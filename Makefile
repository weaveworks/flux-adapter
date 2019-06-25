.PHONY: all clean FORCE

all: build/flux-adapter build/image.tar

clean:
	rm -rf ./build

TINI_VERSION:=v0.18.0
TINI_CHECKSUM:=12d20136605531b09a2c2dac02ccee85e1b874eb322ef6baf7561cd93f93c855

VERSION := $(shell git describe --tags --dirty --always)
VCS_REF:=$(shell git rev-parse HEAD)
BUILD_DATE:=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

build/flux-adapter: FORCE
	go build -o "$@" -ldflags "-X main.version=$(VERSION)" ./cmd/flux-adapter

build/tini: build/tini_$(TINI_VERSION)
	echo "$(TINI_CHECKSUM)  $^" | shasum -a 256 -c
	cp "$^" "$@"
	chmod a+x "$@"

build/tini_$(TINI_VERSION):
	curl -Ls -o $@ "https://github.com/krallin/tini/releases/download/$(TINI_VERSION)/tini"

build/image.tar: Dockerfile build/flux-adapter build/tini
	mkdir -p ./build/docker/
	cp $^ ./build/docker/
	docker build -t docker.io/weaveworks/flux-adapter -t docker.io/weaveworks/flux-adapter:$(VCS_REF) \
		--build-arg VCS_REF="$(VCS_REF)" \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		-f build/docker/Dockerfile ./build/docker/
	docker save docker.io/weaveworks/flux-adapter > "$@"
