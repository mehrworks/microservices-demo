output "ci_contract" {
  description = "Dataset-driven CI contract reserved for later trigger provisioning work."
  value       = local.config
}

output "ci_contract_summary" {
  description = "Compact CI handoff contract for later trigger-provisioning work."
  value = {
    project_id       = try(local.config.project_id, null)
    repo_owner       = try(local.config.repo_owner, null)
    repo_name        = try(local.config.repo_name, null)
    provider         = try(local.config.provider, null)
    trigger_mode     = try(local.config.trigger_mode, null)
    manage_triggers  = try(local.config.manage_triggers, false)
    triggers_enabled = try(local.config.triggers_enabled, false)
    manual_approval  = try(local.config.manual_approval_required, true)
    included_paths   = try(local.config.included_paths, [])
    excluded_paths   = try(local.config.excluded_paths, [])
    placeholder_only = try(local.config.trigger_mode, "placeholder") == "placeholder"
  }
}
