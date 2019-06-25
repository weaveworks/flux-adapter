# Flux adapter for Weave Cloud

This component is used to connect a [fluxd](https://github.com/weaveworks/flux) to Weave Cloud.

It's intended to be run as a sidecar in the same pod as fluxd. By default, it will connect to
Weave Cloud using a websocket, and proxy commands received over the websocket, to fluxd on
`localhost:3030`. It will also act as the receiver for events fluxd POSTs, forwarding them to
Weave Cloud.
