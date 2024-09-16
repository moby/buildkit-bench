# syntax=docker/dockerfile-upstream:master

ARG GO_VERSION=1.22
ARG ALPINE_VERSION=3.20
ARG XX_VERSION=1.4.0

ARG BUILDX_VERSION=0.17.0
ARG REGISTRY_VERSION=v2.8.3

# named contexts
FROM scratch AS buildkit-binaries
FROM scratch AS tests-results

# xx is a helper for cross-compilation
FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

# buildx client
FROM docker/buildx-bin:${BUILDX_VERSION} AS buildx

# go base image
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS golatest

# gobuild is base stage for compiling go/cgo
FROM golatest AS gobuild-base
RUN apk add --no-cache file bash clang lld musl-dev pkgconfig git make tree
COPY --link --from=xx / /

FROM gobuild-base AS registry
WORKDIR /go/src/github.com/docker/distribution
ARG REGISTRY_VERSION
ADD --keep-git-dir=true "https://github.com/distribution/distribution.git#$REGISTRY_VERSION" .
ARG TARGETPLATFORM
RUN --mount=type=cache,target=/root/.cache <<EOT
  set -ex
  mkdir /out
  export GOPATH="$(pwd)/Godeps/_workspace:$GOPATH"
  GO111MODULE=off CGO_ENABLED=0 xx-go build -o /out/registry ./cmd/registry
  xx-verify --static /out/registry
  if [ "$(xx-info os)" = "windows" ]; then
    mv /out/registry /out/registry.exe
  fi
EOT

FROM gobuild-base AS gotestmetrics
WORKDIR /src
ENV GOFLAGS=-mod=vendor
ENV CGO_ENABLED=0
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
  --mount=type=cache,target=/root/.cache \
  --mount=type=cache,target=/go/pkg/mod <<EOT
  set -ex
  xx-go build -mod=vendor -ldflags '-extldflags -static' -o /usr/bin/gotestmetrics ./cmd/gotestmetrics
  xx-verify --static /usr/bin/gotestmetrics
  if ! xx-info is-cross; then
    gotestmetrics --help
  fi
EOT

FROM gobuild-base AS tests-gen-run
COPY --link --from=gotestmetrics /usr/bin/gotestmetrics /usr/bin/
COPY --from=tests-results . /tests-results
ARG GEN_VALIDATION_MODE
RUN --mount=type=bind,target=. <<EOT
  set -e
  args="gen --output /tmp/benchmarks.html"
  if [ -f /tests-results/candidates.json ]; then
    args="$args --candidates /tests-results/candidates.json"
  fi
  if [ -f /tests-results/testconfig.yml ]; then
    args="$args --config /tests-results/testconfig.yml"
  fi
  if [ -n "$GEN_VALIDATION_MODE" ]; then
    args="$args --validation-mode $GEN_VALIDATION_MODE"
  fi
  set -x
  gotestmetrics $args "/tests-results/*.json"
EOT

FROM scratch AS tests-gen
COPY --from=tests-gen-run /tmp/benchmarks.html /index.html

FROM scratch AS binaries
COPY --link --from=registry /out /
COPY --link --from=buildx /buildx /
COPY --link --from=gotestmetrics /usr/bin/gotestmetrics /

FROM gobuild-base AS tests-base
WORKDIR /src
ENV GOFLAGS=-mod=vendor
RUN apk add --no-cache shadow shadow-uidmap sudo vim iptables ip6tables dnsmasq fuse curl git-daemon openssh-client slirp4netns iproute2 \
  && useradd --create-home --home-dir /home/user --uid 1000 -s /bin/sh user \
  && echo "XDG_RUNTIME_DIR=/run/user/1000; export XDG_RUNTIME_DIR" >> /home/user/.profile \
  && mkdir -m 0700 -p /run/user/1000 \
  && chown -R user /run/user/1000 /home/user \
  && ln -s /sbin/iptables-legacy /usr/bin/iptables \
  && xx-go --wrap
# The entrypoint script is needed for enabling nested cgroup v2 (https://github.com/moby/buildkit/issues/3265#issuecomment-1309631736)
RUN curl -Ls https://raw.githubusercontent.com/moby/moby/v25.0.1/hack/dind > /docker-entrypoint.sh && chmod 0755 /docker-entrypoint.sh
ENTRYPOINT ["/docker-entrypoint.sh"]
ENV CGO_ENABLED=0
COPY --link --from=binaries / /usr/bin/

# tests prepares an image suitable for running tests
FROM tests-base AS tests
COPY --link --from=buildkit-binaries / /buildkit-binaries
RUN tree -nph /buildkit-binaries
COPY . .
