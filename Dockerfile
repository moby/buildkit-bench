# syntax=docker/dockerfile-upstream:master

ARG GO_VERSION=1.22
ARG ALPINE_VERSION=3.20
ARG XX_VERSION=1.4.0

ARG REGISTRY_VERSION=v2.8.3
ARG GOTESTSUM_VERSION=v1.9.0

# named context for buildkit binaries
FROM scratch AS buildkit-binaries

# xx is a helper for cross-compilation
FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

# go base image
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS golatest

# gobuild is base stage for compiling go/cgo
FROM golatest AS gobuild-base
RUN apk add --no-cache file bash clang lld musl-dev pkgconfig git make
COPY --link --from=xx / /

FROM gobuild-base AS registry
WORKDIR /go/src/github.com/docker/distribution
ARG REGISTRY_VERSION
ADD --keep-git-dir=true "https://github.com/distribution/distribution.git#$REGISTRY_VERSION" .
ARG TARGETPLATFORM
RUN --mount=target=/root/.cache,type=cache <<EOT
  set -ex
  mkdir /out
  export GOPATH="$(pwd)/Godeps/_workspace:$GOPATH"
  GO111MODULE=off CGO_ENABLED=0 xx-go build -o /out/registry ./cmd/registry
  xx-verify --static /out/registry
  if [ "$(xx-info os)" = "windows" ]; then
    mv /out/registry /out/registry.exe
  fi
EOT

FROM gobuild-base AS gotestsum
ARG GOTESTSUM_VERSION
ARG TARGETPLATFORM
RUN --mount=target=/root/.cache,type=cache <<EOT
  set -ex
  xx-go install "gotest.tools/gotestsum@${GOTESTSUM_VERSION}"
  xx-go install "github.com/wadey/gocovmerge@latest"
  mkdir /out
  if ! xx-info is-cross; then
    /go/bin/gotestsum --version
    mv /go/bin/gotestsum /out
    mv /go/bin/gocovmerge /out
  else
    mv /go/bin/*/gotestsum* /out
    mv /go/bin/*/gocovmerge* /out
  fi
EOT
COPY --chmod=755 <<"EOF" /out/gotestsumandcover
#!/bin/sh
set -x
if [ -z "$GO_TEST_COVERPROFILE" ]; then
  exec gotestsum "$@"
fi
coverdir="$(dirname "$GO_TEST_COVERPROFILE")"
mkdir -p "$coverdir/helpers"
gotestsum "$@" "-coverprofile=$GO_TEST_COVERPROFILE"
ecode=$?
go tool covdata textfmt -i=$coverdir/helpers -o=$coverdir/helpers-report.txt
gocovmerge "$coverdir/helpers-report.txt" "$GO_TEST_COVERPROFILE" > "$coverdir/merged-report.txt"
mv "$coverdir/merged-report.txt" "$GO_TEST_COVERPROFILE"
rm "$coverdir/helpers-report.txt"
for f in "$coverdir/helpers"/*; do
  rm "$f"
done
rmdir "$coverdir/helpers"
exit $ecode
EOF

FROM scratch AS binaries
COPY --link --from=gotestsum /out /
COPY --link --from=registry /out /
COPY --link --from=buildkit-binaries / /

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
ENV GOTESTSUM_FORMAT=standard-verbose
COPY --link --from=binaries / /usr/bin/

# tests prepares an image suitable for running tests
FROM tests-base AS tests
COPY . .
