package replication

import (
	"context"
	"net"
	"testing"

	"github.com/wasp-project/yazi/pkg/config"
	"github.com/wasp-project/yazi/pkg/policy"
	sg "github.com/wasp-project/yazi/pkg/server/grpc"
	"github.com/wasp-project/yazi/pkg/server/pb"
	"github.com/wasp-project/yazi/pkg/storage"
	"google.golang.org/grpc"
)

type testRPCServer struct {
	store storage.KVStore
	pb.UnimplementedRPCServerServer
}

func (s *testRPCServer) Set(ctx context.Context, req *pb.KVRequest) (*pb.KVResponse, error) {
	return &pb.KVResponse{}, s.store.Set(sg.DataKeyPrefix+req.Key, req.Value)
}

func (s *testRPCServer) Get(ctx context.Context, req *pb.KVRequest) (*pb.KVResponse, error) {
	val, err := s.store.Get(sg.DataKeyPrefix + req.Key)
	return &pb.KVResponse{Value: val}, err
}

func (s *testRPCServer) Del(ctx context.Context, req *pb.KVRequest) (*pb.KVResponse, error) {
	return &pb.KVResponse{}, s.store.Del(sg.DataKeyPrefix + req.Key)
}

func (s *testRPCServer) Keys(ctx context.Context, req *pb.MKVRequest) (*pb.MKVResponse, error) {
	keys, err := s.store.Keys()
	return &pb.MKVResponse{Values: keys}, err
}

func (s *testRPCServer) MGet(ctx context.Context, req *pb.MKVRequest) (*pb.MKVResponse, error) {
	nkeys := make([]string, len(req.Keys))
	for i := range req.Keys {
		nkeys[i] = sg.DataKeyPrefix + req.Keys[i]
	}
	vals, err := s.store.MGet(nkeys)
	return &pb.MKVResponse{Values: vals}, err
}

func (s *testRPCServer) MSet(ctx context.Context, req *pb.MKVRequest) (*pb.MKVResponse, error) {
	nkeys := make([]string, len(req.Keys))
	for i := range req.Keys {
		nkeys[i] = sg.DataKeyPrefix + req.Keys[i]
	}
	return &pb.MKVResponse{}, s.store.MSet(nkeys, req.Values)
}

func startTestServer(t *testing.T) (string, storage.KVStore) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	store := storage.NewKVStore(1024, policy.KeyPolicy(""))
	pb.RegisterRPCServerServer(srv, &testRPCServer{store: store})
	go srv.Serve(lis)
	t.Cleanup(func() {
		srv.Stop()
		lis.Close()
	})
	return lis.Addr().String(), store
}

func TestWriteQuorumReplication(t *testing.T) {
	addr1, store1 := startTestServer(t)
	addr2, store2 := startTestServer(t)
	local := storage.NewKVStore(1024, policy.KeyPolicy(""))
	conf := config.ReplicationConfig{
		Enabled:       true,
		Peers:         []string{addr1, addr2},
		WriteQuorum:   2,
		ReadQuorum:    1,
		TimeoutMillis: 500,
	}
	rstore := NewReplicatedStore(local, conf)
	if err := rstore.Set(sg.DataKeyPrefix+"k", "v"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if v, err := store1.Get(sg.DataKeyPrefix + "k"); err != nil || v != "v" {
		t.Fatalf("peer1 value: %v %v", v, err)
	}
	if v, err := store2.Get(sg.DataKeyPrefix + "k"); err != nil || v != "v" {
		t.Fatalf("peer2 value: %v %v", v, err)
	}
}

func TestReadQuorum(t *testing.T) {
	addr1, store1 := startTestServer(t)
	addr2, store2 := startTestServer(t)
	local := storage.NewKVStore(1024, policy.KeyPolicy(""))
	_ = local.Set(sg.DataKeyPrefix+"k", "v1")
	_ = store1.Set(sg.DataKeyPrefix+"k", "v1")
	_ = store2.Set(sg.DataKeyPrefix+"k", "v2")
	conf := config.ReplicationConfig{
		Enabled:       true,
		Peers:         []string{addr1, addr2},
		WriteQuorum:   1,
		ReadQuorum:    2,
		TimeoutMillis: 500,
	}
	rstore := NewReplicatedStore(local, conf)
	val, err := rstore.Get(sg.DataKeyPrefix + "k")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if val != "v1" {
		t.Fatalf("expected v1, got %s", val)
	}
}
