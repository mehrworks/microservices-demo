provider "google" {
  project = local.project_id
  region  = local.region
}

check "private_nodes_require_master_cidr" {
  assert {
    condition     = !try(local.config.hardening.private_nodes_enabled, false) || try(local.config.hardening.master_ipv4_cidr, null) != null
    error_message = "hardening.master_ipv4_cidr must be set when private_nodes_enabled is true."
  }
}

check "managed_network_requires_name" {
  assert {
    condition     = !try(local.config.network.manage, false) || try(local.config.network.network_name, null) != null
    error_message = "network.network_name must be set when network.manage is true."
  }
}

resource "google_compute_network" "managed" {
  count = local.network_enabled ? 1 : 0

  project                 = coalesce(local.project_id, "placeholder-project")
  name                    = coalesce(local.config.network.network_name, "placeholder-network")
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "managed" {
  count = local.subnetwork_enabled ? 1 : 0

  project       = coalesce(local.project_id, "placeholder-project")
  region        = coalesce(local.region, "europe-west3")
  name          = coalesce(local.config.network.subnetwork_name, "placeholder-subnetwork")
  ip_cidr_range = coalesce(local.config.network.subnet_cidr, "10.255.255.0/24")
  network       = google_compute_network.managed[0].id

  dynamic "secondary_ip_range" {
    for_each = local.secondary_ranges
    content {
      range_name    = secondary_ip_range.value.range_name
      ip_cidr_range = secondary_ip_range.value.ip_cidr_range
    }
  }
}

resource "google_artifact_registry_repository" "managed" {
  count = local.artifact_registry_create ? 1 : 0

  project       = coalesce(local.project_id, "placeholder-project")
  location      = coalesce(local.region, "europe-west3")
  repository_id = coalesce(local.repository_id, "placeholder-repository")
  description   = try(local.config.artifact_registry.description, null)
  format        = local.config.artifact_registry.format
}

data "google_artifact_registry_repository" "existing" {
  count = local.artifact_registry_adopt ? 1 : 0

  project       = coalesce(local.project_id, "placeholder-project")
  location      = coalesce(local.region, "europe-west3")
  repository_id = coalesce(local.repository_id, "placeholder-repository")
}

resource "google_artifact_registry_repository_iam_member" "readers" {
  for_each = local.repo_reader_bindings

  project    = coalesce(local.project_id, "placeholder-project")
  location   = coalesce(local.region, "europe-west3")
  repository = coalesce(local.repository_id, "placeholder-repository")
  role       = "roles/artifactregistry.reader"
  member     = each.value
}

resource "google_artifact_registry_repository_iam_member" "writers" {
  for_each = local.repo_writer_bindings

  project    = coalesce(local.project_id, "placeholder-project")
  location   = coalesce(local.region, "europe-west3")
  repository = coalesce(local.repository_id, "placeholder-repository")
  role       = "roles/artifactregistry.writer"
  member     = each.value
}

resource "google_container_cluster" "managed" {
  count = local.cluster_enabled ? 1 : 0

  project             = coalesce(local.project_id, "placeholder-project")
  name                = local.config.cluster.name
  location            = coalesce(local.region, "europe-west3")
  deletion_protection = false
  enable_autopilot    = try(local.config.cluster.autopilot, true)
  network             = local.network_enabled ? google_compute_network.managed[0].self_link : try(local.config.network.network_name, null)
  subnetwork          = local.subnetwork_enabled ? google_compute_subnetwork.managed[0].self_link : try(local.config.network.subnetwork_name, null)

  dynamic "ip_allocation_policy" {
    for_each = length(local.secondary_ranges) == 2 ? [1] : []
    content {
      cluster_secondary_range_name  = local.config.network.pods_secondary_range_name
      services_secondary_range_name = local.config.network.services_secondary_range_name
    }
  }

  dynamic "private_cluster_config" {
    for_each = try(local.config.hardening.private_nodes_enabled, false) ? [1] : []
    content {
      enable_private_nodes   = true
      master_ipv4_cidr_block = local.config.hardening.master_ipv4_cidr
    }
  }

  lifecycle {
    precondition {
      condition     = !try(local.config.hardening.private_nodes_enabled, false) || try(local.config.hardening.master_ipv4_cidr, null) != null
      error_message = "hardening.master_ipv4_cidr must be set when private_nodes_enabled is true."
    }
  }

  depends_on = [
    google_compute_network.managed,
    google_compute_subnetwork.managed,
    google_artifact_registry_repository.managed
  ]
}

data "google_container_cluster" "existing" {
  count = local.cluster_adopt ? 1 : 0

  project  = coalesce(local.project_id, "placeholder-project")
  name     = local.config.cluster.name
  location = coalesce(local.region, "europe-west3")
}
