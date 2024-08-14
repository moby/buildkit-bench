variable "BUILDKIT_REPO" {
  default = "moby/buildkit"
}

variable "BUILDKIT_REF" {
  default = "master"
}

variable "BUILDKIT_TARGET" {
  default = "binaries"
}

variable "GOBUILDFLAGS" {
  default = null
}

variable "TEST_COVERAGE" {
  default = null
}

group "default" {
  targets = ["tests"]
}

target "buildkit-binaries" {
  context = "https://github.com/${BUILDKIT_REPO}.git#${BUILDKIT_REF}"
  target = BUILDKIT_TARGET
  args = {
    BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
    BUILDKIT_DEBUG = 1
  }
}

target "tests-base" {
  contexts = {
    buildkit-binaries = "target:buildkit-binaries"
  }
  target = "tests-base"
  output = ["type=cacheonly"]
}

target "tests" {
  inherits = ["tests-base"]
  target = "tests"
  args = {
    GOBUILDFLAGS = TEST_COVERAGE == "1" ? "-cover" : null
  }
}
