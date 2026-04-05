output "project_id" {
  description = "Active project for the GKE stage."
  value       = local.project_id
}

output "region" {
  description = "Target region for registry and cluster resources."
  value       = local.region
}

output "namespace" {
  description = "Namespace default for the app-delivery proof lane."
  value       = try(local.config.namespace, null)
}

output "artifact_registry_repository_id" {
  description = "Artifact Registry repository identifier."
  value       = local.repository_id
}

output "artifact_registry_repository_url" {
  description = "Default Artifact Registry repo URL for Skaffold or CI."
  value       = local.repository_url
}

output "artifact_registry_contract" {
  description = "Artifact Registry contract emitted by this stage."
  value = {
    repository_id  = local.repository_id
    repository_url = local.repository_url
    created        = local.artifact_registry_create
    adopted        = local.artifact_registry_adopt
    readers        = try(local.config.artifact_registry.readers, [])
    writers        = try(local.config.artifact_registry.writers, [])
  }
}

output "cluster_contract" {
  description = "Cluster-side contract emitted by this stage."
  value = {
    name                         = try(local.config.cluster.name, null)
    autopilot                    = try(local.config.cluster.autopilot, true)
    created                      = try(local.config.cluster.create, false)
    adopted                      = local.cluster_adopt
    endpoint                     = try(google_container_cluster.managed[0].endpoint, try(data.google_container_cluster.existing[0].endpoint, null))
    private_nodes_enabled        = try(local.config.hardening.private_nodes_enabled, false)
    master_ipv4_cidr             = try(local.config.hardening.master_ipv4_cidr, null)
    cloud_nat_enabled            = try(local.config.hardening.cloud_nat_enabled, false)
    public_load_balancer_enabled = try(local.config.hardening.public_load_balancer_enabled, true)
  }
}

output "network_contract" {
  description = "Optional network contract for org-managed environments."
  value = {
    managed                       = try(local.config.network.manage, false)
    network_name                  = try(local.config.network.network_name, null)
    subnetwork_name               = try(local.config.network.subnetwork_name, null)
    subnet_cidr                   = try(local.config.network.subnet_cidr, null)
    pods_secondary_range_name     = try(local.config.network.pods_secondary_range_name, null)
    services_secondary_range_name = try(local.config.network.services_secondary_range_name, null)
  }
}
