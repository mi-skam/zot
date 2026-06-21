# Local fork builds keep the upstream tag visible and append -fork.
# Release builds are driven by goreleaser which overrides VERSION from the git tag.
UPSTREAM_VERSION ?= $(shell git describe --tags --abbrev=0 --match 'v*' 2>/dev/null | sed 's/^v//')
VERSION ?= $(if $(UPSTREAM_VERSION),$(UPSTREAM_VERSION)-fork,0.0.0-fork)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE ?= $(shell git log -1 --format=%cI 2>/dev/null)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build install test lint fmt clean release run

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/zot ./cmd/zot

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/zot

test:
	go test -race ./...

run: build
	./bin/zot

lint:
	go vet ./...
	@test -z "$$(gofmt -l . | tee /dev/stderr)" || (echo "gofmt issues"; exit 1)

fmt:
	gofmt -w .

clean:
	rm -rf bin

release:
	@mkdir -p bin
	GOOS=linux   GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/zot-linux-amd64   ./cmd/zot
	GOOS=linux   GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/zot-linux-arm64   ./cmd/zot
	GOOS=darwin  GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/zot-darwin-amd64  ./cmd/zot
	GOOS=darwin  GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/zot-darwin-arm64  ./cmd/zot
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/zot-windows-amd64.exe ./cmd/zot
