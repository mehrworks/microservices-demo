provider "google" {
  project = try(local.config.project_id, null)
  region  = try(local.config.region, null)
}

check "identity_mode_contract" {
  assert {
    condition = alltrue([
      for name, account in local.service_accounts :
      try(account.create, false) || try(account.email, null) != null || length(lookup(local.service_account_project_roles, name, [])) == 0
    ])
    error_message = "Each service account with role bindings must either set create=true or provide service_accounts.<name>.email for adoption."
  }
}

resource "google_service_account" "managed" {
  for_each = local.creatable_accounts

  project      = coalesce(local.config.project_id, "placeholder-project")
  account_id   = each.value.account_id
  display_name = each.value.display_name
  description  = try(each.value.description, null)
}

resource "google_project_iam_member" "service_account_roles" {
  for_each = local.project_role_pairs

  project = coalesce(local.config.project_id, "placeholder-project")
  role    = each.value.role
  member  = each.value.create ? "serviceAccount:${google_service_account.managed[each.value.account].email}" : "serviceAccount:${each.value.adopted_email}"
}
