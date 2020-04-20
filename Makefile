Version := $(shell date "+%Y%m%d%H%M")
GitCommit := $(shell git rev-parse HEAD)
DIR := $(shell pwd)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"

run: build
	./build/debug/dev-proxy ./config.yaml

build:
	go build -race -ldflags $(LDFLAGS) -o build/debug/dev-proxy main.go

build-release:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -o build/release/dev-proxy-darwin main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS) -o build/release/dev-proxy.exe main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o build/release/dev-proxy-linux main.go

clean:
	rm -fr build/debug/dev-proxy build/release/dev-proxy*

.PHONY: run build build-release clean
