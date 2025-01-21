# syntax=docker/dockerfile-upstream:master

ARG GO_VERSION=1.23
ARG ALPINE_VERSION=3.21
ARG XX_VERSION=1.6.1

ARG BUILDKIT_VERSION=v0.19.0
ARG BUILDX_VERSION=v0.20.0
ARG REGISTRY_VERSION=v2.8.3

# named contexts
FROM scratch AS buildkit-binaries
FROM scratch AS buildx-binaries
FROM scratch AS tests-results

# xx is a helper for cross-compilation
FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

# default buildkit
FROM moby/buildkit:${BUILDKIT_VERSION} AS buildkit

# default buildx
FROM docker/buildx-bin:${BUILDX_VERSION#v} AS buildx

# go base image
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS golatest

# gobuild is base stage for compiling go/cgo
FROM golatest AS gobuild-base
RUN apk add --no-cache file bash clang lld musl-dev pkgconfig git make tree findutils
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
  xx-go build -mod=vendor -ldflags '-extldflags -static' -o /out/gotestmetrics ./cmd/gotestmetrics
  xx-verify --static /out/gotestmetrics
  if ! xx-info is-cross; then
    /out/gotestmetrics --help
  fi
EOT
COPY --chmod=755 <<"EOF" /out/gotestmetrics-gen
#!/bin/bash
set -e
project=$1
if [ -z "$project" ]; then
  echo "project argument is required"
  exit 1
fi
args=("gen" "--output=/out/$project.html" "--project=$project")
if [ -f /tests-results/name.txt ]; then
  args+=("--name=$(cat /tests-results/name.txt)")
fi
if [ -f /tests-results/testconfig.yml ]; then
  args+=("--config=/tests-results/testconfig.yml")
fi
if [ -f /tests-results/gha-event.json ]; then
  args+=("--gha-event=/tests-results/gha-event.json")
fi
if [ -f /tests-results/env.txt ]; then
  args+=("--envs=/tests-results/env.txt")
fi
if [ -n "$GEN_VALIDATION_MODE" ]; then
  args+=("--validation-mode=$GEN_VALIDATION_MODE")
fi
case "$project" in
  buildkit)
    if [ -f "/tests-results/buildkit-candidates.json" ]; then
      args+=("--candidates=/tests-results/buildkit-candidates.json")
    elif [ -f "/tests-results/candidates.json" ]; then
      args+=("--candidates=/tests-results/candidates.json")
    fi
    if find /tests-results -maxdepth 1 -type f -name 'gotestoutput-buildkit-*' | grep -q .; then
      args+=("/tests-results/gotestoutput-buildkit-*.json")
    else
      args+=("/tests-results/gotestoutput*.json")
    fi
  ;;
  buildx)
    if [ -f "/tests-results/buildx-candidates.json" ]; then
      args+=("--candidates=/tests-results/buildx-candidates.json")
    fi
    if find /tests-results -maxdepth 1 -type f -name 'gotestoutput-buildx-*' | grep -q .; then
      args+=("/tests-results/gotestoutput-buildx-*.json")
    else
      if [ "$GITHUB_ACTIONS" = "true" ]; then
        # for backward compatibility with old test results in GitHub Pages
        exit 0
      else
        args+=("/tests-results/gotestoutput*.json")
      fi
    fi
    ;;
esac
set -x
gotestmetrics "${args[@]}"
EOF

FROM gobuild-base AS tests-gen-run
COPY --link --from=gotestmetrics /out /usr/bin/
COPY --from=tests-results . /tests-results
ARG GITHUB_ACTIONS
ARG GEN_VALIDATION_MODE
RUN --mount=type=bind,target=. <<EOT
  set -e
  mkdir -p /out
  if [ -d /tests-results/logs ]; then
    tar -czvf /out/logs.tar.gz -C /tests-results/logs .
  fi
  gotestmetrics-gen buildkit
  gotestmetrics-gen buildx
EOT

FROM scratch AS tests-gen
COPY --from=tests-gen-run /out /

FROM scratch AS binaries
COPY --link --from=registry /out /
COPY --link --from=gotestmetrics /out /

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
ARG BUILDKIT_VERSION
COPY --link --from=buildkit /usr/bin/buildctl* /opt/buildkit/${BUILDKIT_VERSION}/
COPY --link --from=buildkit /usr/bin/buildkit* /opt/buildkit/${BUILDKIT_VERSION}/
ARG BUILDX_VERSION
COPY --link --from=buildx /buildx /opt/buildx/${BUILDX_VERSION}/
COPY --link --from=binaries / /usr/bin/

# tests prepares an image suitable for running tests
FROM tests-base AS tests
COPY --link --from=buildkit-binaries / /buildkit-binaries
COPY --link --from=buildx-binaries / /buildx-binaries
RUN tree -nph /buildkit-binaries
RUN tree -nph /buildx-binaries
COPY . .
