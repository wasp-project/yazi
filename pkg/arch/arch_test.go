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

// Package arch holds architecture-conformance tests for the layered memory
// system. They enforce the dependency direction between layers so the storage
// engine stays the reusable core, independent of the tenant and UI layers.
package arch

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const modulePath = "github.com/wasp-project/yazi"

// repoRoot returns the module root by walking up from this test file.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	// file = <root>/pkg/arch/arch_test.go
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

// imports collects all import paths used by non-test .go files under dir.
func imports(t *testing.T, dir string) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	fset := token.NewFileSet()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			out[path] = append(out[path], p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return out
}

// TestStorageEngineLayerIsIndependent asserts the storage-engine layer
// (pkg/storage and its subpackages) does not depend on the tenant-service layer
// (pkg/tenant), the memory model (pkg/memory), or the user-interface layers
// (cmd/cli, pkg/server). The storage engine must remain the reusable core.
func TestStorageEngineLayerIsIndependent(t *testing.T) {
	root := repoRoot(t)
	storageDir := filepath.Join(root, "pkg", "storage")

	forbidden := []string{
		modulePath + "/pkg/tenant",
		modulePath + "/pkg/memory",
		modulePath + "/pkg/server",
		modulePath + "/cmd/cli",
	}

	for file, imps := range imports(t, storageDir) {
		for _, imp := range imps {
			for _, bad := range forbidden {
				if imp == bad || strings.HasPrefix(imp, bad+"/") {
					rel, _ := filepath.Rel(root, file)
					t.Errorf("storage layer must not import %s (found in %s)", bad, rel)
				}
			}
		}
	}
}

// TestTenantLayerIsStorageAgnostic asserts the tenant-service layer does not
// depend on the storage engine, keeping its interface implementable by an
// external project without coupling to internal persistence details.
func TestTenantLayerIsStorageAgnostic(t *testing.T) {
	root := repoRoot(t)
	tenantDir := filepath.Join(root, "pkg", "tenant")

	for file, imps := range imports(t, tenantDir) {
		for _, imp := range imps {
			if strings.HasPrefix(imp, modulePath+"/pkg/storage") ||
				strings.HasPrefix(imp, modulePath+"/pkg/memory") {
				rel, _ := filepath.Rel(root, file)
				t.Errorf("tenant layer must stay storage-agnostic, but %s imports %s", rel, imp)
			}
		}
	}
}
