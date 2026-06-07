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

// Package s3 is a cloud-vendor persistence backend (the fourth layer of the
// memory system). It implements storage.PersistentStorage by writing the
// encoded store snapshot to an AWS S3 (or S3-compatible) object and reading it
// back. It depends only on the standard library: a small AWS SigV4 signer
// rather than the AWS SDK, to keep the build hermetic and the module minimal.
// It can be swapped for the official SDK later behind the same interface.
package s3

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Config describes the target object and credentials for the S3 backend.
type Config struct {
	Bucket    string
	Region    string
	Key       string
	Endpoint  string // optional; set for S3-compatible stores. Empty => AWS.
	AccessKey string
	SecretKey string
}

// Storage is an S3-backed PersistentStorage. The zero value is not usable; use
// New.
type Storage struct {
	conf   Config
	client *http.Client
	now    func() time.Time // injectable for tests
}

// New validates the config and constructs an S3 backend. It returns an error
// when required fields are missing so startup can fail fast instead of silently
// falling back to local storage.
func New(conf Config) (*Storage, error) {
	if strings.TrimSpace(conf.Bucket) == "" {
		return nil, errors.New("s3: bucket is required")
	}
	if strings.TrimSpace(conf.Region) == "" {
		return nil, errors.New("s3: region is required")
	}
	if strings.TrimSpace(conf.Key) == "" {
		return nil, errors.New("s3: key is required")
	}
	if conf.AccessKey == "" || conf.SecretKey == "" {
		return nil, errors.New("s3: accessKey and secretKey are required (set them in config or AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY)")
	}
	return &Storage{
		conf:   conf,
		client: &http.Client{Timeout: 30 * time.Second},
		now:    time.Now,
	}, nil
}

// Write puts the snapshot bytes to the configured object, replacing it.
func (s *Storage) Write(data []byte) (int, error) {
	req, err := s.signedRequest(http.MethodPut, data)
	if err != nil {
		return 0, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return 0, fmt.Errorf("s3: put failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return len(data), nil
}

// Read fetches the configured object into the provided buffer. A missing object
// (first run, no snapshot yet) is not an error: it returns 0, nil.
func (s *Storage) Read(data []byte) (int, error) {
	req, err := s.signedRequest(http.MethodGet, nil)
	if err != nil {
		return 0, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return 0, nil
	}
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return 0, fmt.Errorf("s3: get failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	n, err := io.ReadFull(resp.Body, data)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		// Object is smaller than the buffer: that's the normal case.
		return n, nil
	}
	return n, err
}

// signedRequest builds and SigV4-signs an HTTP request for the object.
func (s *Storage) signedRequest(method string, body []byte) (*http.Request, error) {
	rawURL, host, canonicalURI := s.objectURL()
	req, err := http.NewRequest(method, rawURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	payloadHash := sha256Hex(body)

	req.Host = host
	req.Header.Set("Host", host)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if method == http.MethodPut {
		req.Header.Set("Content-Type", "application/octet-stream")
		req.ContentLength = int64(len(body))
	}

	const service = "s3"
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalHeaders := "host:" + host + "\n" +
		"x-amz-content-sha256:" + payloadHash + "\n" +
		"x-amz-date:" + amzDate + "\n"

	canonicalRequest := strings.Join([]string{
		method,
		canonicalURI,
		"", // no query string
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	scope := strings.Join([]string{dateStamp, s.conf.Region, service, "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	signingKey := signatureKey(s.conf.SecretKey, dateStamp, s.conf.Region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	authz := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.conf.AccessKey, scope, signedHeaders, signature,
	)
	req.Header.Set("Authorization", authz)
	return req, nil
}

// objectURL returns the request URL, the Host header to sign, and the canonical
// URI (path) for signing. A configured endpoint uses path-style addressing
// (endpoint/bucket/key); AWS uses virtual-hosted-style (bucket.s3.region...).
func (s *Storage) objectURL() (rawURL, host, canonicalURI string) {
	key := strings.TrimPrefix(s.conf.Key, "/")
	encodedKey := encodePath(key)
	if s.conf.Endpoint != "" {
		ep := strings.TrimRight(s.conf.Endpoint, "/")
		host = hostOf(ep)
		canonicalURI = "/" + s.conf.Bucket + "/" + encodedKey
		rawURL = ep + canonicalURI
		return
	}
	host = fmt.Sprintf("%s.s3.%s.amazonaws.com", s.conf.Bucket, s.conf.Region)
	canonicalURI = "/" + encodedKey
	rawURL = "https://" + host + canonicalURI
	return
}

func hostOf(endpoint string) string {
	h := endpoint
	if i := strings.Index(h, "://"); i >= 0 {
		h = h[i+3:]
	}
	if i := strings.IndexByte(h, '/'); i >= 0 {
		h = h[:i]
	}
	return h
}

// encodePath percent-encodes a key per RFC 3986 while preserving '/' separators,
// matching the canonicalization S3 SigV4 expects.
func encodePath(p string) string {
	var b strings.Builder
	for i := 0; i < len(p); i++ {
		c := p[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9',
			c == '-', c == '_', c == '.', c == '~', c == '/':
			b.WriteByte(c)
		default:
			b.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return b.String()
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func signatureKey(secret, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}
