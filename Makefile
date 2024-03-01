# Copyright 2024 mlycore. All rights reserved.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

build:
	GOBIN=$(LOCLABIN) go build -o bin/yazi cmd/yazi/yazi.go
	GOBIN=$(LOCALBIN) go build -o bin/yazictl cmd/cli/yazi_ctl.go
	GOBIN=$(LOCALBIN) go build -o bin/yazibench cmd/bench/yazi_bench.go
test:
	GOBIN=$(LOCALBIN) go test ./pkg/... -gcflags=-l -coverprofile=coverage.txt
benchmark:
	GOBIN=$(LOCALBIN) go test ./pkg/... -bench=. -race -count=1 -coverprofile=coverage.txt


.PHONY: lint
lint: check-lint
	GOBIN=$(LOCALBIN) CGO_ENABLED=0 golangci-lint run -v --timeout=5m ./... 

.PHONY: check-lint
check-lint: $(CHECK_LINT) ## Download golangci-lint-setup locally if necessary.
$(CHECK_LINT): $(LOCALBIN)
	GOBIN=$(LOCALBIN) CGO_ENABLED=0 go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest