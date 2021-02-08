# Flux adapter for Weave Cloud

[![CircleCI](https://circleci.com/gh/weaveworks/flux-adapter.svg?style=svg)](https://circleci.com/gh/weaveworks/flux-adapter)

This component is used to connect a [fluxd](https://github.com/weaveworks/flux) to Weave Cloud.

It's intended to be run as a sidecar in the same pod as fluxd. By default, it will connect to
Weave Cloud using a websocket, and proxy commands received over the websocket, to fluxd on
`localhost:3030`. It will also act as the receiver for events fluxd POSTs, forwarding them to
Weave Cloud.

## <a name="help"></a>Getting Help

If you have any questions about, feedback for or problems with `flux-adapter`:

- Invite yourself to the <a href="https://slack.weave.works/" target="_blank">Weave Users Slack</a>.
- Ask a question on the [#general](https://weave-community.slack.com/messages/general/) slack channel.
- [File an issue](https://github.com/weaveworks/flux-adapter/issues/new).

Weaveworks follows the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md). Instances of abusive, harassing, or otherwise unacceptable behavior may be reported by contacting a Weaveworks project maintainer, or Alexis Richardson (alexis@weave.works).

Your feedback is always welcome!
