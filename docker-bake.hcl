# always set to true when GitHub Actions is running the workflow.
variable "GITHUB_ACTIONS" {
  default = null
}

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
      ref == "v0.17.0" || ref == "fd61877fa73693dcd4ef64c538f894ec216409a3" ? "v0.17.0-grpc-fix" :
      ref == "v0.17.1" || ref == "8b1b83ef4947c03062cdcdb40c69989d8fe3fd04" ? "v0.17.1-grpc-fix" :
      ref
    ),
    context = (
      ref == "v0.12.0" || ref == "18fc875d9bfd6e065cd8211abc639434ba65aa56" ? "https://github.com/crazy-max/buildkit.git#v0.12.0-pick-pr-4361" :
      ref == "v0.17.0" || ref == "fd61877fa73693dcd4ef64c538f894ec216409a3" ? "https://github.com/crazy-max/buildkit.git#0.17.0-grpc-fix" :
      ref == "v0.17.1" || ref == "8b1b83ef4947c03062cdcdb40c69989d8fe3fd04" ? "https://github.com/crazy-max/buildkit.git#0.17.1-grpc-fix" :
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
  name = "buildkit-build-${sanitize(ref)}"
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
    format("buildkit-build-%s", sanitize(ref.value)) => format("target:buildkit-build-%s", sanitize(ref.value))
  }
  dockerfile-inline = <<EOT
FROM scratch
${join("\n", [for ref in parse_refs(BUILDKIT_REFS) :
  format("COPY --link --from=buildkit-build-%s / /%s", sanitize(ref.value), ref.key)
])}
EOT
  output = ["type=cacheonly"]
}

target "buildx-build" {
  inherits = ["_common"]
  name = "buildx-build-${sanitize(ref)}"
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
    format("buildx-build-%s", sanitize(ref.value)) => format("target:buildx-build-%s", sanitize(ref.value))
  }
  dockerfile-inline = <<EOT
FROM scratch
${join("\n", [for ref in parse_refs(BUILDX_REFS) :
  format("COPY --link --from=buildx-build-%s / /%s", sanitize(ref.value), ref.key)
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

target "tests-buildkit" {
  inherits = ["tests-base"]
  contexts = {
    buildkit-binaries = "target:buildkit-binaries"
  }
  target = "tests"
}

target "tests-buildx" {
  inherits = ["tests-base"]
  contexts = {
    buildx-binaries = "target:buildx-binaries"
  }
  target = "tests"
}

variable "GEN_VALIDATION_MODE" {
  default = null
}

target "tests-gen" {
  inherits = ["_common"]
  contexts = {
    tests-results = "./bin/results"
  }
  target = "tests-gen"
  output = ["./bin/gen"]
  args = {
    GITHUB_ACTIONS = GITHUB_ACTIONS
    GEN_VALIDATION_MODE = GEN_VALIDATION_MODE
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
