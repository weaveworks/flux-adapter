FROM alpine:3.10 as certs
RUN apk --update add -U --no-cache ca-certificates

FROM scratch

WORKDIR /home/flux

# These are pretty static
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="flux-adapter" \
      org.opencontainers.image.description="The Flux adapter connects Flux to Weave Cloud" \
      org.opencontainers.image.url="https://github.com/weaveworks/flux-adapter" \
      org.opencontainers.image.source="git@github.com:weaveworks/flux-adapter" \
      org.opencontainers.image.vendor="Weaveworks" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.name="flux-adapter" \
      org.label-schema.description="The Flux adapter connects Flux to Weave Cloud" \
      org.label-schema.url="https://github.com/weaveworks/flux-adapter" \
      org.label-schema.vcs-url="git@github.com:weaveworks/flux-adapter" \
      org.label-schema.vendor="Weaveworks"

ENTRYPOINT [ "/sbin/tini", "--", "/usr/local/bin/flux-adapter" ]

# Create minimal nsswitch.conf file to prioritize the usage of /etc/hosts over DNS queries.
# This resolves the conflict between:
# * flux-adapter using netgo for static compilation. netgo reads nsswitch.conf to mimic glibc,
#   defaulting to prioritize DNS queries over /etc/hosts if nsswitch.conf is missing:
#   https://github.com/golang/go/issues/22846
# * Alpine not including a nsswitch.conf file. Since Alpine doesn't use glibc
#   (it uses musl), maintainers argue that the need of nsswitch.conf is a Go bug:
#   https://github.com/gliderlabs/docker-alpine/issues/367#issuecomment-354316460
RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ./tini /sbin/
COPY ./flux-adapter /usr/local/bin/

ARG BUILD_DATE
ARG VCS_REF

# These will change for every build
LABEL org.opencontainers.image.revision="$VCS_REF" \
      org.opencontainers.image.created="$BUILD_DATE" \
      org.label-schema.vcs-ref="$VCS_REF" \
      org.label-schema.build-date="$BUILD_DATE"
