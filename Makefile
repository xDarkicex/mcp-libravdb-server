VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")

BIN            := bin/mcp-memory-libravdb
MAIN           := ./cmd/mcp-memory-libravdb
LDFLAGS        := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)
INSTALL_PREFIX ?= /usr/local/bin
COVERAGE_MIN   ?= 70
CC_MAX         ?= 8

.PHONY: all verify lint test test-race coverage build build-linux run run-http run-stdio clean install-systemd install-launchd

all: verify build

verify: lint test-race coverage

lint:
	golangci-lint run ./...

test:
	go test ./...

test-race:
	go test -race ./...

coverage:
	go test -coverprofile=bin/coverage.out ./...
	go tool cover -func=bin/coverage.out

build:
	go build -ldflags="$(LDFLAGS)" -o $(BIN) $(MAIN)

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BIN)-linux $(MAIN)

run: run-stdio

run-stdio:
	go run -ldflags="$(LDFLAGS)" $(MAIN) stdio

run-http:
	go run -ldflags="$(LDFLAGS)" $(MAIN) http

clean:
	rm -rf bin/ dist/

install-systemd: build
	install -m 0755 $(BIN) $(INSTALL_PREFIX)/mcp-memory-libravdb
	install -m 0644 deploy/systemd/mcp-memory-libravdb.service /etc/systemd/system/mcp-memory-libravdb.service
	systemctl daemon-reload

install-launchd: build
	install -m 0755 $(BIN) $(INSTALL_PREFIX)/mcp-memory-libravdb
	install -m 0644 deploy/launchd/com.zephyr-systems.mcp-memory-libravdb.plist /Library/LaunchDaemons/com.zephyr-systems.mcp-memory-libravdb.plist
