locals {
  dataset_dir = "${path.module}/../../config/datasets/project"

  defaults  = yamldecode(file("${local.dataset_dir}/defaults.yaml"))
  overrides = yamldecode(fileexists("${local.dataset_dir}/overrides.yaml") ? file("${local.dataset_dir}/overrides.yaml") : "{}")

  config = merge(local.defaults, local.overrides)

  enabled_apis = toset(try(local.config.enabled_apis, []))
}
