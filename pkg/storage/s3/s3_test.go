// Copyright 2024 mlycore. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s3

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeS3 is a minimal in-process S3-compatible object store for testing the
// backend offline (path-style: /<bucket>/<key>).
func fakeS3(t *testing.T) (*httptest.Server, map[string][]byte) {
	t.Helper()
	objects := map[string][]byte{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Require that the request was signed.
		if !strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256 ") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.Header.Get("X-Amz-Content-Sha256") == "" || r.Header.Get("X-Amz-Date") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			objects[r.URL.Path] = body
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			body, ok := objects[r.URL.Path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(body)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, objects
}

func newTestStorage(t *testing.T, endpoint string) *Storage {
	t.Helper()
	s, err := New(Config{
		Bucket:    "memory",
		Region:    "us-east-1",
		Key:       "yazi.data",
		Endpoint:  endpoint,
		AccessKey: "AKIDEXAMPLE",
		SecretKey: "secret",
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	return s
}

func TestWriteReadRoundTrip(t *testing.T) {
	srv, objects := fakeS3(t)
	s := newTestStorage(t, srv.URL)

	payload := []byte(`{"/_memory/basic/1":"hello"}`)
	if n, err := s.Write(payload); err != nil || n != len(payload) {
		t.Fatalf("write: n=%d err=%v", n, err)
	}

	// The object landed under path-style /<bucket>/<key>.
	if _, ok := objects["/memory/yazi.data"]; !ok {
		t.Fatalf("object not stored at expected path; have %v", keysOf(objects))
	}

	buf := make([]byte, 1024)
	n, err := s.Read(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(buf[:n]) != string(payload) {
		t.Fatalf("round-trip mismatch: got %q want %q", string(buf[:n]), string(payload))
	}
}

func TestReadMissingObjectIsEmpty(t *testing.T) {
	srv, _ := fakeS3(t)
	s := newTestStorage(t, srv.URL)

	buf := make([]byte, 1024)
	n, err := s.Read(buf)
	if err != nil {
		t.Fatalf("read missing should not error, got %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 bytes for missing object, got %d", n)
	}
}

func TestCustomEndpointHonored(t *testing.T) {
	srv, _ := fakeS3(t)
	s := newTestStorage(t, srv.URL)
	_, host, uri := s.objectURL()
	if !strings.Contains(srv.URL, host) {
		t.Fatalf("expected host derived from endpoint %s, got %s", srv.URL, host)
	}
	if uri != "/memory/yazi.data" {
		t.Fatalf("expected path-style canonical URI, got %s", uri)
	}
}

func TestNewValidatesRequiredFields(t *testing.T) {
	cases := []struct {
		name string
		conf Config
	}{
		{"missing bucket", Config{Region: "us-east-1", Key: "k", AccessKey: "a", SecretKey: "s"}},
		{"missing region", Config{Bucket: "b", Key: "k", AccessKey: "a", SecretKey: "s"}},
		{"missing key", Config{Bucket: "b", Region: "us-east-1", AccessKey: "a", SecretKey: "s"}},
		{"missing creds", Config{Bucket: "b", Region: "us-east-1", Key: "k"}},
	}
	for _, c := range cases {
		if _, err := New(c.conf); err == nil {
			t.Fatalf("%s: expected error", c.name)
		}
	}
}

func keysOf(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
