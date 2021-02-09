module github.com/weaveworks/flux-adapter

go 1.15

require (
	github.com/fluxcd/flux v1.20.2
	github.com/go-kit/kit v0.9.0
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.17.11
	k8s.io/apimachinery v0.17.11
	k8s.io/client-go v11.0.0+incompatible
)

// remove when https://github.com/docker/distribution/pull/2905 is released.
replace github.com/docker/distribution => github.com/fluxcd/distribution v0.0.0-20190419185413-6c9727e5e5de

// this is taken from the fluxcd/flux go.mod
replace github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0

// this is local to flux, and doesn't have a version of its own; the line taken from fluxcd/helm-operator
replace github.com/fluxcd/flux/pkg/install => github.com/fluxcd/flux/pkg/install v0.0.0-20200206191601-8b676b003ab0

// stop Go pulling in client-go v11.0.0
replace k8s.io/client-go => k8s.io/client-go v0.17.11
