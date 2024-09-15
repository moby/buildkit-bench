# syntax=docker/dockerfile:1

ARG NODE_VERSION=20

FROM node:${NODE_VERSION}-alpine AS base
RUN apk add --no-cache cpio findutils git
WORKDIR /src
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/src/.yarn/cache <<EOT
  corepack enable
  yarn --version
  yarn config set --home enableTelemetry 0
EOT

FROM base AS deps
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/src/.yarn/cache \
  --mount=type=cache,target=/src/node_modules \
  yarn install

FROM deps AS build
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/src/.yarn/cache \
  --mount=type=cache,target=/src/node_modules \
  yarn run build && cp -Rf ./dist /out

FROM scratch
COPY --from=build /out /
