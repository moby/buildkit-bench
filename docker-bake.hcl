variable "BUILDKIT_REPO" {
  default = "moby/buildkit"
}

variable "BUILDKIT_REFS" {
  default = "master"
}

variable "BUILDKIT_TARGET" {
  default = "binaries"
}

# https://github.com/docker/buildx/blob/8411a763d99274c7585553f0354a7fdd0df679eb/bake/bake.go#L35
# TODO: use sanitize func once buildx 0.17.0 is released https://github.com/docker/buildx/pull/2649
function "sanitize_target" {
  params = [in]
  result = regex_replace(in, "[^a-zA-Z0-9_-]+", "-")
}

function "parse_refs" {
  params = [refs]
  result = [
    for ref in split(",", refs) :
    {
      key = (can(regex("^[^=]+=", ref)) ? split("=", ref)[0] : ref),
      value = (can(regex("^[^=]+=", ref)) ? split("=", ref)[1] : ref)
    }
  ]
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

target "buildkit-build" {
  inherits = ["_common"]
  name = "buildkit-build-${sanitize_target(ref)}"
  matrix = {
    ref = [for item in parse_refs(BUILDKIT_REFS) : item.value]
  }
  context = "https://github.com/${BUILDKIT_REPO}.git#${ref}"
  target = BUILDKIT_TARGET
  args = {
    BUILDKIT_DEBUG = 1
  }
}

target "buildkit-binaries" {
  contexts = { for ref in parse_refs(BUILDKIT_REFS) :
    format("buildkit-build-%s", sanitize_target(ref.value)) => format("target:buildkit-build-%s", sanitize_target(ref.value))
  }
  dockerfile-inline = <<EOT
FROM scratch
${join("\n", [for ref in parse_refs(BUILDKIT_REFS) :
  format("COPY --link --from=buildkit-build-%s / /%s", sanitize_target(ref.value), ref.key)
])}
EOT
  output = ["type=cacheonly"]
}

target "tests-base" {
  inherits = ["_common"]
  target = "tests-base"
  output = ["type=cacheonly"]
}

target "tests" {
  inherits = ["tests-base"]
  contexts = {
    buildkit-binaries = "target:buildkit-binaries"
  }
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
