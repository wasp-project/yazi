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

package local

import "os"

type DiskWriter struct{}

const (
	file = "yazi.data"
)

func NewWriter() *DiskWriter {
	return &DiskWriter{}
}

func (w *DiskWriter) Write(data []byte) (int, error) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}

	if err := f.Truncate(0); err != nil {
		return 0, err
	}

	if _, err := f.Seek(0, 0); err != nil {
		return 0, err
	}

	return f.Write(data)
}

type DiskReader struct{}

func NewReader() *DiskReader {
	return &DiskReader{}
}

func (r *DiskReader) Read(data []byte) (int, error) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}

	return f.Read(data)
}

type LocalStorage struct {
	*DiskWriter
	*DiskReader
}

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		DiskWriter: NewWriter(),
		DiskReader: NewReader(),
	}
}

func (s *LocalStorage) Writer() *DiskWriter {
	return s.DiskWriter
}

func (s *LocalStorage) Reader() *DiskReader {
	return s.DiskReader
}
