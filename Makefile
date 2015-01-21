VERSION := 0.1.0
LDFLAGS := -ldflags "-X main.Version $(VERSION)"

SRC=$(wildcard *.go)
TGT=web

OS=$(subst Darwin,darwin,$(subst Linux,linux,$(shell uname)))
ARCH=$(shell uname -m)

GOVER=go1.3.1
GOOS=$(subst Darwin,darwin,$(subst Linux,linux,$(OS)))
GOARCH=$(subst x86_64,amd64,$(ARCH))
GOPKG=$(subst darwin-amd64,darwin-amd64-osx10.8,$(GOVER).$(GOOS)-$(GOARCH).tar.gz)
GOROOT=$(CURDIR)/.deps/go
GOPATH=$(CURDIR)/.deps/gopath
GOCC=$(GOROOT)/bin/go
GO=GOROOT=$(GOROOT) GOPATH=$(GOPATH) $(GOCC)

default:
			go build

build: $(TGT)

test: $(GOCC) $(SRC)
			$(GO) test

.deps/$(GOPKG):
			mkdir -p .deps
						curl -o .deps/$(GOPKG) https://storage.googleapis.com/golang/$(GOPKG)

$(GOCC): .deps/$(GOPKG)
			tar -C .deps -xzf .deps/$(GOPKG)
						touch $@

dependencies: $(SRC)
			$(GO) get -d

$(TGT): $(GOCC) $(SRC) dependencies
			$(GO) build $(LDFLAGS) -v -o $(TGT)

clean:
			rm -rf .deps/

format:
			find . -iname '*.go' -exec gofmt -w -s=true '{}' ';'

advice:
			go tool vet .

.PHONY: advice build clean dependencies format test
