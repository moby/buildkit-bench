variable "BUILDKIT_REPO" {
  default = "moby/buildkit"
}

variable "BUILDKIT_REFS" {
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

# https://github.com/docker/buildx/blob/8411a763d99274c7585553f0354a7fdd0df679eb/bake/bake.go#L35
function "sanitize_target" {
  params = [in]
  result = regex_replace(in, "[^a-zA-Z0-9_-]+", "-")
}

# generate named context for each ref
function "buildkit_build_contexts" {
  params = [refs]
  result = { for ref in split(",", refs) :
    format("buildkit-build-%s", sanitize_target(ref)) => format("target:buildkit-build-%s", sanitize_target(ref))
  }
}

# generate COPY instructions for each ref
function "buildkit_build_copy_froms" {
  params = [refs]
  result = join("\n", [for ref in split(",", refs) :
    format("COPY --link --from=buildkit-build-%s / /%s", sanitize_target(ref), ref)
  ])
}

group "default" {
  targets = ["tests"]
}

target "buildkit-build" {
  name = "buildkit-build-${sanitize_target(ref)}"
  matrix = {
    ref = split(",", BUILDKIT_REFS)
  }
  context = "https://github.com/${BUILDKIT_REPO}.git#${ref}"
  target = BUILDKIT_TARGET
  args = {
    BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
    BUILDKIT_DEBUG = 1
  }
}

target "buildkit-binaries" {
  contexts = buildkit_build_contexts(BUILDKIT_REFS)
  dockerfile-inline = <<EOT
FROM scratch
${buildkit_build_copy_froms(BUILDKIT_REFS)}
EOT
}

target "tests-base" {
  contexts = {
    buildkit-binaries = "target:buildkit-binaries"
  }
  target = "tests-base"
  args = {
    BUILDKIT_REFS = BUILDKIT_REFS
  }
  output = ["type=cacheonly"]
}

target "tests" {
  inherits = ["tests-base"]
  target = "tests"
  args = {
    GOBUILDFLAGS = TEST_COVERAGE == "1" ? "-cover" : null
  }
}
