variable "BUILDKIT_REPO" {
  default = "moby/buildkit"
}

variable "BUILDKIT_CACHE_REPO" {
  default = "ghcr.io/moby/buildkit-bench-cache"
}

variable "BUILDKIT_REF" {
  default = "master"
}

variable "BUILDKIT_REF_TYPE" {
  default = "ref"
}

variable "BUILDKIT_REF_COMMIT" {
  default = null
}

variable "BUILDKIT_TARGET" {
  default = "binaries"
}

group "default" {
  targets = ["tests"]
}

target "_common" {
  cache-from = [
    "type=registry,ref=${BUILDKIT_CACHE_REPO}:${BUILDKIT_REF_TYPE}-${BUILDKIT_REF}",
    "${BUILDKIT_REF_COMMIT != null ? "type=registry,ref=${BUILDKIT_CACHE_REPO}:${BUILDKIT_REF_COMMIT}" : ""}"
  ]
  cache-to = ["type=inline"]
}

target "buildkit-binaries" {
  inherits = ["_common"]
  context = "https://github.com/${BUILDKIT_REPO}.git#${BUILDKIT_REF}"
  target = BUILDKIT_TARGET
  args = {
    BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
    BUILDKIT_DEBUG = 1
  }
}

target "tests-base" {
  inherits = ["_common"]
  contexts = {
    buildkit-binaries = "target:buildkit-binaries"
  }
  target = "tests-base"
  args = {
    BUILDKIT_REF = BUILDKIT_REF
  }
  output = ["type=cacheonly"]
}

target "tests" {
  inherits = ["tests-base"]
  target = "tests"
}
