output "project_id" {
  description = "Active project identifier for the stage."
  value       = try(local.config.project_id, null)
}

output "project_number" {
  description = "Resolved or explicitly bound project number."
  value       = try(data.google_project.current[0].number, try(local.config.project_number, null))
}

output "enabled_apis" {
  description = "API services managed by this stage."
  value       = sort(tolist(local.enabled_apis))
}

output "project_contract" {
  description = "Project-side contract emitted by this stage for downstream lanes."
  value = {
    project_id     = try(local.config.project_id, null)
    project_number = try(data.google_project.current[0].number, try(local.config.project_number, null))
    region         = try(local.config.region, null)
    enabled_apis   = sort(tolist(local.enabled_apis))
    adopt_only     = true
  }
}
