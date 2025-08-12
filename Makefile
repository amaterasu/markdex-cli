APP_NAME=markdex
PKG=github.com/amaterasu/markdex-cli
GOBIN?=$(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN:=$(shell go env GOPATH)/bin
endif
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE?=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-s -w -X '$(PKG)/cmd.version=$(VERSION)' -X '$(PKG)/cmd.commit=$(COMMIT)' -X '$(PKG)/cmd.date=$(DATE)'

.PHONY: build install clean run

build:
	GOFLAGS="-trimpath" go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME) .

install:
	@mkdir -p $(GOBIN)
	GOFLAGS="-trimpath" go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/$(APP_NAME) .

run: build
	./bin/$(APP_NAME) $(ARGS)

clean:
	rm -rf bin

# Cross compile example (linux/amd64): make build GOOS=linux GOARCH=amd64
