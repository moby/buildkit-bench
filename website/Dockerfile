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

FROM base AS install
COPY .yarnrc.yml package.json yarn.lock ./
RUN yarn install

FROM base AS dev
COPY --from=install /src/node_modules /src/node_modules
COPY . .

FROM base AS build
ARG WEBSITE_PUBLIC_PATH
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/src/.yarn/cache \
  --mount=from=install,source=/src/node_modules,target=/src/node_modules \
  yarn run build && cp -Rf ./dist /out

FROM scratch AS build-update
COPY --from=build /out /
