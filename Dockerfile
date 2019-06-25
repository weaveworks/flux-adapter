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

COPY ./tini /sbin/
COPY ./flux-adapter /usr/local/bin/

ARG BUILD_DATE
ARG VCS_REF

# These will change for every build
LABEL org.opencontainers.image.revision="$VCS_REF" \
      org.opencontainers.image.created="$BUILD_DATE" \
      org.label-schema.vcs-ref="$VCS_REF" \
      org.label-schema.build-date="$BUILD_DATE"
