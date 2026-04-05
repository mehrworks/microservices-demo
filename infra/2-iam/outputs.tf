output "service_accounts" {
  description = "Managed service-account contract for downstream stages."
  value = {
    for name, account in local.service_accounts :
    name => {
      create       = try(account.create, false)
      adopted      = !try(account.create, false) && try(account.email, null) != null
      account_id   = try(account.account_id, null)
      display_name = try(account.display_name, null)
      email        = try(google_service_account.managed[name].email, try(account.email, null))
      member       = try("serviceAccount:${google_service_account.managed[name].email}", try(account.email, null) == null ? null : "serviceAccount:${account.email}")
    }
  }
}

output "iam_contract" {
  description = "Compact identity contract for downstream lanes."
  value = {
    project_id = try(local.config.project_id, null)
    region     = try(local.config.region, null)
    service_accounts = {
      for name, account in local.service_accounts :
      name => {
        create  = try(account.create, false)
        adopted = !try(account.create, false) && try(account.email, null) != null
        email   = try(google_service_account.managed[name].email, try(account.email, null))
        member  = try("serviceAccount:${google_service_account.managed[name].email}", try(account.email, null) == null ? null : "serviceAccount:${account.email}")
        roles   = lookup(local.service_account_project_roles, name, [])
      }
    }
  }
}

output "service_account_project_roles" {
  description = "Configured project-role bindings for the managed identities."
  value       = local.service_account_project_roles
}
