locals {
  dataset_dir = "${path.module}/../../config/datasets/gke"

  defaults  = yamldecode(file("${local.dataset_dir}/defaults.yaml"))
  overrides = yamldecode(fileexists("${local.dataset_dir}/overrides.yaml") ? file("${local.dataset_dir}/overrides.yaml") : "{}")

  artifact_registry = merge(
    lookup(local.defaults, "artifact_registry", {}),
    lookup(local.overrides, "artifact_registry", {})
  )

  cluster = merge(
    lookup(local.defaults, "cluster", {}),
    lookup(local.overrides, "cluster", {})
  )

  network = merge(
    lookup(local.defaults, "network", {}),
    lookup(local.overrides, "network", {})
  )

  hardening = merge(
    lookup(local.defaults, "hardening", {}),
    lookup(local.overrides, "hardening", {})
  )

  proof = merge(
    lookup(local.defaults, "proof", {}),
    lookup(local.overrides, "proof", {})
  )

  config = merge(local.defaults, local.overrides, {
    artifact_registry = local.artifact_registry
    cluster           = local.cluster
    network           = local.network
    hardening         = local.hardening
    proof             = local.proof
  })

  project_id = try(local.config.project_id, null)
  region     = try(local.config.region, null)

  network_enabled    = local.project_id != null && try(local.config.network.manage, false) && try(local.config.network.network_name, null) != null
  subnetwork_enabled = local.network_enabled && try(local.config.network.subnetwork_name, null) != null && try(local.config.network.subnet_cidr, null) != null

  secondary_ranges = concat(
    try(local.config.network.pods_secondary_range_name, null) != null && try(local.config.network.pods_secondary_cidr, null) != null ? [{
      range_name    = local.config.network.pods_secondary_range_name
      ip_cidr_range = local.config.network.pods_secondary_cidr
    }] : [],
    try(local.config.network.services_secondary_range_name, null) != null && try(local.config.network.services_secondary_cidr, null) != null ? [{
      range_name    = local.config.network.services_secondary_range_name
      ip_cidr_range = local.config.network.services_secondary_cidr
    }] : []
  )

  artifact_registry_enabled = local.project_id != null && try(local.config.artifact_registry.manage, false) && try(local.config.artifact_registry.repository_id, null) != null
  artifact_registry_create  = local.artifact_registry_enabled && try(local.config.artifact_registry.create, false)
  artifact_registry_adopt   = local.artifact_registry_enabled && !try(local.config.artifact_registry.create, false)

  cluster_enabled = local.project_id != null && try(local.config.cluster.manage, false) && try(local.config.cluster.create, false)
  cluster_adopt   = local.project_id != null && try(local.config.cluster.manage, false) && !try(local.config.cluster.create, false) && try(local.config.cluster.name, null) != null

  repository_id  = try(local.config.artifact_registry.repository_id, null)
  repository_url = local.project_id == null || local.repository_id == null ? null : "${local.region}-docker.pkg.dev/${local.project_id}/${local.repository_id}"

  repo_reader_bindings = {
    for member in try(local.config.artifact_registry.readers, []) :
    member => member
    if local.artifact_registry_enabled
  }

  repo_writer_bindings = {
    for member in try(local.config.artifact_registry.writers, []) :
    member => member
    if local.artifact_registry_enabled
  }
}
