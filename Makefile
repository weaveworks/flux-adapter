.PHONY: all FORCE

all: flux-adapter

VERSION := $(shell git describe --tags --dirty --always)

flux-adapter: FORCE
	go build -o "$@" -ldflags "-X main.version=$(VERSION)" ./cmd/flux-adapter
