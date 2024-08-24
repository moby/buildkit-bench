variable "BUILDKIT_REPO" {
  default = "moby/buildkit"
}

variable "BUILDKIT_REF" {
  default = "master"
}

variable "BUILDKIT_TARGET" {
  default = "binaries"
}

group "default" {
  targets = ["tests"]
}

group "validate" {
  targets = ["validate-vendor"]
}

target "_common" {
  args = {
    BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
  }
}

target "buildkit-binaries" {
  inherits = ["_common"]
  context = "https://github.com/${BUILDKIT_REPO}.git#${BUILDKIT_REF}"
  target = BUILDKIT_TARGET
  args = {
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

target "validate-vendor" {
  inherits = ["_common"]
  dockerfile = "./hack/dockerfiles/vendor.Dockerfile"
  target = "validate"
  output = ["type=cacheonly"]
}

target "vendor" {
  inherits = ["_common"]
  dockerfile = "./hack/dockerfiles/vendor.Dockerfile"
  target = "update"
  output = ["."]
}
