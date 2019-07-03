.PHONY: all clean image FORCE

all: build/flux-adapter build/image.tar

clean:
	rm -rf ./build

image: build/image.tar

TINI_VERSION:=v0.18.0
TINI_CHECKSUM:=eadb9d6e2dc960655481d78a92d2c8bc021861045987ccd3e27c7eae5af0cf33

VERSION:=$(shell git describe --tags --dirty --always)
VCS_REF:=$(shell git rev-parse HEAD)
BUILD_DATE:=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
STATIC=-tags netgo -ldflags '-extldflags "-static"'

IMAGE_NAME ?= "local/flux-adapter"

build/flux-adapter: FORCE
	GOOS=linux GOARCH=amd64 go build -o "$@" $(STATIC) -ldflags "-X main.version=$(VERSION)" ./cmd/flux-adapter

build/tini: build/tini_$(TINI_VERSION)
	echo "$(TINI_CHECKSUM)  $^" | shasum -a 256 -c
	cp "$^" "$@"
	chmod a+x "$@"

build/tini_$(TINI_VERSION):
	curl -Ls -o $@ "https://github.com/krallin/tini/releases/download/$(TINI_VERSION)/tini-static-amd64"

build/image.tar: Dockerfile build/flux-adapter build/tini
	mkdir -p ./build/docker/
	cp $^ ./build/docker/
	docker build -t $(IMAGE_NAME) \
		--build-arg VCS_REF="$(VCS_REF)" \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		-f build/docker/Dockerfile ./build/docker/
	docker save $(IMAGE_NAME) > "$@"
