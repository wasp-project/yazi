package replication

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/wasp-project/yazi/pkg/config"
	sg "github.com/wasp-project/yazi/pkg/server/grpc"
	"github.com/wasp-project/yazi/pkg/server/pb"
	"github.com/wasp-project/yazi/pkg/storage"
	"google.golang.org/grpc"
	"strings"
)

type clientPool struct {
	mu      sync.Mutex
	clients map[string]*grpc.ClientConn
}

func newClientPool() *clientPool {
	return &clientPool{clients: map[string]*grpc.ClientConn{}}
}

func (p *clientPool) get(addr string, timeout time.Duration) (*grpc.ClientConn, error) {
	p.mu.Lock()
	if conn, ok := p.clients[addr]; ok {
		p.mu.Unlock()
		return conn, nil
	}
	p.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	p.clients[addr] = conn
	p.mu.Unlock()
	return conn, nil
}

type Store struct {
	local       storage.KVStore
	peers       []string
	writeQuorum int
	readQuorum  int
	timeout     time.Duration
	pool        *clientPool
}

func NewReplicatedStore(local storage.KVStore, conf config.ReplicationConfig) *Store {
	total := len(conf.Peers) + 1
	writeQuorum := conf.WriteQuorum
	readQuorum := conf.ReadQuorum
	if writeQuorum <= 0 {
		writeQuorum = total/2 + 1
	}
	if readQuorum <= 0 {
		readQuorum = total/2 + 1
	}
	if writeQuorum > total {
		writeQuorum = total
	}
	if readQuorum > total {
		readQuorum = total
	}
	timeout := time.Duration(conf.TimeoutMillis) * time.Millisecond
	if timeout == 0 {
		timeout = 500 * time.Millisecond
	}
	return &Store{
		local:       local,
		peers:       conf.Peers,
		writeQuorum: writeQuorum,
		readQuorum:  readQuorum,
		timeout:     timeout,
		pool:        newClientPool(),
	}
}

func (s *Store) Get(key string) (string, error) {
	localVal, localErr := s.local.Get(key)
	if s.readQuorum <= 1 || len(s.peers) == 0 {
		return localVal, localErr
	}
	peerKey, ok := normalizeKeyForPeer(key)
	if !ok {
		return localVal, localErr
	}
	type result struct {
		val string
		err error
	}
	results := make(chan result, len(s.peers))
	for _, peer := range s.peers {
		go func(addr string) {
			val, err := s.remoteGet(addr, peerKey)
			results <- result{val: val, err: err}
		}(peer)
	}
	success := 0
	values := map[string]int{}
	if localErr == nil {
		values[localVal]++
		success++
	}
	for i := 0; i < len(s.peers); i++ {
		r := <-results
		if r.err == nil {
			values[r.val]++
			success++
		}
		if success >= s.readQuorum {
			break
		}
	}
	if success < s.readQuorum {
		if localErr != nil {
			return "", errors.New("read quorum not met")
		}
		return localVal, nil
	}
	var bestVal string
	bestCount := -1
	for v, c := range values {
		if c > bestCount {
			bestVal = v
			bestCount = c
		}
	}
	return bestVal, nil
}

func (s *Store) Set(key, val string) error {
	if err := s.local.Set(key, val); err != nil {
		return err
	}
	peerKey, ok := normalizeKeyForPeer(key)
	if !ok {
		return nil
	}
	return s.replicateWrite(func(ctx context.Context, cli pb.RPCServerClient) error {
		_, err := cli.Set(ctx, &pb.KVRequest{Key: peerKey, Value: val})
		return err
	})
}

func (s *Store) Expire(key string, ttl time.Duration) error {
	return s.local.Expire(key, ttl)
}

func (s *Store) Del(key string) error {
	if err := s.local.Del(key); err != nil {
		return err
	}
	peerKey, ok := normalizeKeyForPeer(key)
	if !ok {
		return nil
	}
	return s.replicateWrite(func(ctx context.Context, cli pb.RPCServerClient) error {
		_, err := cli.Del(ctx, &pb.KVRequest{Key: peerKey})
		return err
	})
}

func (s *Store) MSet(keys, values []string) error {
	if err := s.local.MSet(keys, values); err != nil {
		return err
	}
	peerKeys, ok := normalizeKeysForPeer(keys)
	if !ok {
		return nil
	}
	return s.replicateWrite(func(ctx context.Context, cli pb.RPCServerClient) error {
		_, err := cli.MSet(ctx, &pb.MKVRequest{Keys: peerKeys, Values: values})
		return err
	})
}

func (s *Store) MGet(keys []string) ([]string, error) {
	localVals, localErr := s.local.MGet(keys)
	if s.readQuorum <= 1 || len(s.peers) == 0 {
		return localVals, localErr
	}
	peerKeys, ok := normalizeKeysForPeer(keys)
	if !ok {
		return localVals, localErr
	}
	type result struct {
		vals []string
		err  error
	}
	results := make(chan result, len(s.peers))
	for _, peer := range s.peers {
		go func(addr string) {
			vals, err := s.remoteMGet(addr, peerKeys)
			results <- result{vals: vals, err: err}
		}(peer)
	}
	success := 0
	valueCounts := make([]map[string]int, len(keys))
	for i := range valueCounts {
		valueCounts[i] = map[string]int{}
	}
	if localErr == nil {
		success++
		for i, v := range localVals {
			valueCounts[i][v]++
		}
	}
	for i := 0; i < len(s.peers); i++ {
		r := <-results
		if r.err != nil {
			continue
		}
		success++
		for idx, v := range r.vals {
			valueCounts[idx][v]++
		}
		if success >= s.readQuorum {
			break
		}
	}
	if success < s.readQuorum {
		if localErr != nil {
			return nil, errors.New("read quorum not met")
		}
		return localVals, nil
	}
	final := make([]string, len(keys))
	for i := range valueCounts {
		bestVal := ""
		bestCount := -1
		for v, c := range valueCounts[i] {
			if c > bestCount {
				bestVal = v
				bestCount = c
			}
		}
		final[i] = bestVal
	}
	return final, nil
}

func (s *Store) Keys() ([]string, error) {
	return s.local.Keys()
}

func (s *Store) Encode() []byte {
	return s.local.Encode()
}

func (s *Store) Decode(data []byte) error {
	return s.local.Decode(data)
}

func (s *Store) replicateWrite(apply func(ctx context.Context, cli pb.RPCServerClient) error) error {
	if len(s.peers) == 0 || s.writeQuorum <= 1 {
		return nil
	}
	results := make(chan error, len(s.peers))
	for _, peer := range s.peers {
		go func(addr string) {
			results <- s.remoteApply(addr, apply)
		}(peer)
	}
	success := 1
	for i := 0; i < len(s.peers); i++ {
		if err := <-results; err == nil {
			success++
		}
		if success >= s.writeQuorum {
			return nil
		}
	}
	return errors.New("write quorum not met")
}

func (s *Store) remoteApply(addr string, apply func(ctx context.Context, cli pb.RPCServerClient) error) error {
	conn, err := s.pool.get(addr, s.timeout)
	if err != nil {
		return err
	}
	cli := pb.NewRPCServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return apply(ctx, cli)
}

func (s *Store) remoteGet(addr, key string) (string, error) {
	return s.remoteValue(addr, func(ctx context.Context, cli pb.RPCServerClient) (string, error) {
		resp, err := cli.Get(ctx, &pb.KVRequest{Key: key})
		if err != nil {
			return "", err
		}
		return resp.Value, nil
	})
}

func (s *Store) remoteMGet(addr string, keys []string) ([]string, error) {
	conn, err := s.pool.get(addr, s.timeout)
	if err != nil {
		return nil, err
	}
	cli := pb.NewRPCServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	resp, err := cli.MGet(ctx, &pb.MKVRequest{Keys: keys})
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

func (s *Store) remoteValue(addr string, call func(ctx context.Context, cli pb.RPCServerClient) (string, error)) (string, error) {
	conn, err := s.pool.get(addr, s.timeout)
	if err != nil {
		return "", err
	}
	cli := pb.NewRPCServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return call(ctx, cli)
}

func normalizeKeyForPeer(key string) (string, bool) {
	if strings.HasPrefix(key, sg.MetadataKeyPrefix) || strings.HasPrefix(key, sg.RaftKeyPrefix) {
		return "", false
	}
	if strings.HasPrefix(key, sg.DataKeyPrefix) {
		return strings.TrimPrefix(key, sg.DataKeyPrefix), true
	}
	if strings.HasPrefix(key, "/_") {
		return "", false
	}
	return key, true
}

func normalizeKeysForPeer(keys []string) ([]string, bool) {
	peerKeys := make([]string, len(keys))
	for i, key := range keys {
		nkey, ok := normalizeKeyForPeer(key)
		if !ok {
			return nil, false
		}
		peerKeys[i] = nkey
	}
	return peerKeys, true
}
