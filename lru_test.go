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

package yazi

import (
	"testing"
)

func TestCacheSetGet(t *testing.T) {
	l := New[int, int](8192)

	for i := 0; i < 8192; i++ {
		l.Set(i, i, 0)
	}

	for i := 0; i < 8192; i++ {
		if i%2 == 0 {
			l.Set(i, i+1, 0)
		}
	}

	if l.size != 8192 {
		t.Errorf("expected size 8192, got %d", l.size)
	}

	for i := 0; i < 8192; i++ {
		if _, ok := l.Get(i); ok {
			if i%2 == 0 {
				if l.cache[i].Val != i+1 {
					t.Errorf("expected %d, got %d", i, l.cache[i].Val)
				}
			} else {
				if l.cache[i].Val != i {
					t.Errorf("expected %d, got %d", i, l.cache[i].Val)
				}
			}
		} else {
			t.Errorf("expected %d, got nothing", i)
		}
	}
}

func BenchmarkCacheSet(b *testing.B) {
	l := New[int, int](8192)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Set(i, i, 0)
	}
}

func BenchmarkCacheUpdate(b *testing.B) {
	l := New[int, int](8192)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Set(i, i, 0)
	}
	for i := 0; i < b.N; i++ {
		l.Set(i, i+1, 0)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	l := New[int, int](8192)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Set(i, i, 0)
	}

	for i := 0; i < b.N; i++ {
		l.Get(i)
	}
}
