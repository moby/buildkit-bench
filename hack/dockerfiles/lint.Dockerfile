# syntax=docker/dockerfile:1

ARG GO_VERSION=1.27
# renovate: datasource=docker depName=alpine versioning=docker
ARG ALPINE_VERSION=3.24
# renovate: datasource=docker depName=tonistiigi/xx versioning=docker
ARG XX_VERSION=1.10.0

# renovate: datasource=github-releases depName=golangci/golangci-lint
ARG GOLANGCI_LINT_VERSION=v2.15.0
ARG GOLANGCI_FROM_SOURCE=true
# renovate: datasource=go depName=github.com/crazy-max/gopls-analyzers
ARG GOPLS_ANALYZERS_VERSION=v0.3.0

FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base
RUN apk add --no-cache git gcc musl-dev binutils-gold

FROM base AS golangci-build
WORKDIR /src
ARG GOLANGCI_LINT_VERSION
ADD --keep-git-dir=true https://github.com/golangci/golangci-lint.git#${GOLANGCI_LINT_VERSION} .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/ go mod download
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/ mkdir -p out && go build -o /out/golangci-lint ./cmd/golangci-lint

FROM --platform=$BUILDPLATFORM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint
FROM scratch AS golangci-binary-false
COPY --from=golangci-lint /usr/bin/golangci-lint golangci-lint
FROM scratch AS golangci-binary-true
COPY --from=golangci-build /out/golangci-lint golangci-lint
FROM golangci-binary-${GOLANGCI_FROM_SOURCE} AS golangci-binary

FROM base AS lint-base
ENV GOFLAGS="-buildvcs=false"
COPY --link --from=golangci-binary / /usr/bin/
COPY --link --from=xx / /
WORKDIR /go/src/github.com/moby/buildkit-bench
ARG TARGETPLATFORM

FROM lint-base AS lint
RUN --mount=target=/go/src/github.com/moby/buildkit-bench \
    --mount=target=/root/.cache,type=cache,id=lint-cache-$TARGETPLATFORM \
    xx-go --wrap && \
    golangci-lint run

FROM lint-base AS validate-golangci
RUN --mount=target=/go/src/github.com/moby/buildkit-bench \
  golangci-lint config verify

FROM base AS gopls-analyzers
ARG GOPLS_ANALYZERS_VERSION
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/ \
  GOBIN=/out go install github.com/crazy-max/gopls-analyzers/cmd/gopls-analyzers@${GOPLS_ANALYZERS_VERSION}

FROM base AS gopls-analyze
COPY --link --from=xx / /
ARG TARGETNAME
ARG TARGETPLATFORM
WORKDIR /go/src/github.com/moby/buildkit-bench
RUN --mount=target=. \
  --mount=target=/root/.cache,type=cache,id=lint-cache-${TARGETNAME}-${TARGETPLATFORM} \
  --mount=target=/usr/bin/gopls-analyzers,from=gopls-analyzers,source=/out/gopls-analyzers <<EOF
  set -ex
  xx-go --wrap
  gopls-analyzers -list-analyzers
  go vet -vettool=/usr/bin/gopls-analyzers -verbose -exclude-generated ./...
EOF

FROM base AS gopls-fix-run
COPY --link --from=xx / /
ARG TARGETNAME
ARG TARGETPLATFORM
WORKDIR /go/src/github.com/moby/buildkit-bench
RUN --mount=target=.,rw \
  --mount=target=/root/.cache,type=cache,id=lint-cache-${TARGETNAME}-${TARGETPLATFORM} \
  --mount=target=/usr/bin/gopls-analyzers,from=gopls-analyzers,source=/out/gopls-analyzers <<EOF
  set -ex
  xx-go --wrap
  mkdir /out
  /usr/bin/gopls-analyzers -fix -exclude-generated ./...
  for file in $(git status --porcelain | awk '/^ M/ {print $2}'); do
    mkdir -p /out/$(dirname $file)
    cp $file /out/$file
  done
EOF

FROM scratch AS gopls-fix
COPY --link --from=gopls-fix-run /out /

FROM lint
