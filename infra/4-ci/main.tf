check "trigger_contract_requirements" {
  assert {
    condition     = !try(local.config.triggers_enabled, false) || (try(local.config.project_id, null) != null && try(local.config.repo_owner, null) != null)
    error_message = "When triggers_enabled is true, both project_id and repo_owner must be set in config/datasets/ci/overrides.yaml."
  }
}

check "trigger_mode_contract" {
  assert {
    condition     = !try(local.config.manage_triggers, false) || try(local.config.triggers_enabled, false)
    error_message = "manage_triggers cannot be true when triggers_enabled is false."
  }
}

check "placeholder_mode_contract" {
  assert {
    condition     = try(local.config.trigger_mode, "placeholder") != "placeholder" || !try(local.config.manage_triggers, false)
    error_message = "trigger_mode=placeholder must keep manage_triggers=false."
  }
}

# Placeholder-only stage. Trigger resources are intentionally deferred.
