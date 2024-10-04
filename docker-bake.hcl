variable "BUILDKIT_CACHE_REPO" {
  default = "moby/buildkit-bench-cache"
}

variable "BUILDKIT_REPO" {
  default = "moby/buildkit"
}

variable "BUILDKIT_REFS" {
  default = "master"
}

variable "BUILDKIT_TARGET" {
  default = "binaries"
}

variable "BUILDX_REPO" {
  default = "docker/buildx"
}

variable "BUILDX_REFS" {
  default = "master"
}

variable "BUILDX_TARGET" {
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

function "buildkit_ref_info" {
  params = [ref]
  result = {
    cache_tag = (
      ref == "v0.12.0" || ref == "18fc875d9bfd6e065cd8211abc639434ba65aa56" ? "v0.12.0-pick-pr-4361" :
      ref
    ),
    context = (
      ref == "v0.12.0" || ref == "18fc875d9bfd6e065cd8211abc639434ba65aa56" ? "https://github.com/crazy-max/buildkit.git#v0.12.0-pick-pr-4361" :
      can(regex("^pr-(\\d+)$", ref)) ? "https://github.com/${BUILDKIT_REPO}.git#refs/pull/${regex_replace(ref, "^pr-(\\d+)$", "$1")}/merge" :
      "https://github.com/${BUILDKIT_REPO}.git#${ref}"
    )
  }
}

function "buildx_ref_info" {
  params = [ref]
  result = {
    cache_tag = ref,
    context = (
      can(regex("^pr-(\\d+)$", ref)) ? "https://github.com/${BUILDX_REPO}.git#refs/pull/${regex_replace(ref, "^pr-(\\d+)$", "$1")}/merge" :
      "https://github.com/${BUILDX_REPO}.git#${ref}"
    )
  }
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
  context = buildkit_ref_info(ref).context
  target = BUILDKIT_TARGET
  cache-from = ["type=registry,ref=${BUILDKIT_CACHE_REPO}:bkbins-${buildkit_ref_info(ref).cache_tag}"]
  cache-to = ["type=inline"]
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

target "buildx-build" {
  inherits = ["_common"]
  name = "buildx-build-${sanitize_target(ref)}"
  matrix = {
    ref = [for item in parse_refs(BUILDX_REFS) : item.value]
  }
  context = buildx_ref_info(ref).context
  target = BUILDX_TARGET
  cache-from = ["type=registry,ref=${BUILDKIT_CACHE_REPO}:bxbins-${buildx_ref_info(ref).cache_tag}"]
  cache-to = ["type=inline"]
}

target "buildx-binaries" {
  contexts = { for ref in parse_refs(BUILDX_REFS) :
    format("buildx-build-%s", sanitize_target(ref.value)) => format("target:buildx-build-%s", sanitize_target(ref.value))
  }
  dockerfile-inline = <<EOT
FROM scratch
${join("\n", [for ref in parse_refs(BUILDX_REFS) :
  format("COPY --link --from=buildx-build-%s / /%s", sanitize_target(ref.value), ref.key)
])}
EOT
  output = ["type=cacheonly"]
}

target "tests-base" {
  inherits = ["_common"]
  target = "tests-base"
  cache-from = ["type=registry,ref=${BUILDKIT_CACHE_REPO}:tests-base"]
  cache-to = ["type=inline"]
  output = ["type=cacheonly"]
}

target "tests" {
  inherits = ["tests-base"]
  contexts = {
    buildkit-binaries = "target:buildkit-binaries"
    buildx-binaries = "target:buildx-binaries"
  }
  target = "tests"
}

variable "GEN_VALIDATION_MODE" {
  default = null
}
variable "CANDIDATES_KEY" {
  default = "candidates"
}

target "tests-gen" {
  inherits = ["_common"]
  contexts = {
    tests-results = "./bin/results"
  }
  target = "tests-gen"
  output = ["./bin/gen"]
  args = {
    GEN_VALIDATION_MODE = GEN_VALIDATION_MODE
    CANDIDATES_KEY = CANDIDATES_KEY
  }
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

variable "WEBSITE_PUBLIC_PATH" {
  default = null
}

target "website" {
  context = "./website"
  target = "build-update"
  output = ["./bin/website"]
  args = {
    WEBSITE_PUBLIC_PATH = WEBSITE_PUBLIC_PATH
  }
}
