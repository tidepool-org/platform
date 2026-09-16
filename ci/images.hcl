// Make supplies the service list, existing release tags, and compiled binaries.
variable "CI_IMAGE_PREFIX" {
  default = "tidepool/platform"
}

variable "CI_IMAGE_SUFFIX" {
  default = ""
}

variable "CI_SERVICES" {
  default = ""
  validation {
    condition = CI_SERVICES != ""
    error_message = "Run make ci-docker to select services and build their binaries."
  }
}

variable "CI_TAGS" {
  default = ""
  validation {
    condition = CI_TAGS != ""
    error_message = "CI image publishing requires explicit release tags."
  }
}

variable "CI_PLUGIN_VISIBILITY" {
  default = "public"
}

variable "CI_BIN_DIRECTORY" {
  default = "./_bin"
}

variable "CI_PLATFORM" {
  default = ""
}

group "default" {
  targets = ["services"]
}

target "services" {
  name = "platform-${service}"
  matrix = {
    service = split(" ", CI_SERVICES)
  }
  context = "."
  dockerfile = "Dockerfile"
  target = "platform-${service}"
  contexts = {
    platform-binaries = CI_BIN_DIRECTORY
  }
  args = {
    PLUGIN_VISIBILITY = CI_PLUGIN_VISIBILITY
  }
  platforms = [CI_PLATFORM]
  tags = [for tag in split(" ", CI_TAGS) : "${CI_IMAGE_PREFIX}-${service}${CI_IMAGE_SUFFIX}:${tag}"]
  // Preserve the existing single-platform image format.
  attest = ["type=provenance,disabled=true"]
}
