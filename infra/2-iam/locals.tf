locals {
  dataset_dir = "${path.module}/../../config/datasets/iam"

  defaults  = yamldecode(file("${local.dataset_dir}/defaults.yaml"))
  overrides = yamldecode(fileexists("${local.dataset_dir}/overrides.yaml") ? file("${local.dataset_dir}/overrides.yaml") : "{}")

  service_account_names = toset(concat(
    keys(lookup(local.defaults, "service_accounts", {})),
    keys(lookup(local.overrides, "service_accounts", {}))
  ))

  service_accounts = {
    for name in local.service_account_names :
    name => merge(
      lookup(lookup(local.defaults, "service_accounts", {}), name, {}),
      lookup(lookup(local.overrides, "service_accounts", {}), name, {})
    )
  }

  role_binding_names = toset(concat(
    keys(lookup(local.defaults, "service_account_project_roles", {})),
    keys(lookup(local.overrides, "service_account_project_roles", {}))
  ))

  service_account_project_roles = {
    for name in local.role_binding_names :
    name => distinct(concat(
      lookup(lookup(local.defaults, "service_account_project_roles", {}), name, []),
      lookup(lookup(local.overrides, "service_account_project_roles", {}), name, [])
    ))
  }

  config = merge(local.defaults, local.overrides, {
    service_accounts              = local.service_accounts
    service_account_project_roles = local.service_account_project_roles
  })

  creatable_accounts = {
    for name, account in local.service_accounts :
    name => account
    if try(local.config.project_id, null) != null && try(account.create, false)
  }

  project_role_pairs = {
    for pair in flatten([
      for account_name, roles in local.service_account_project_roles : [
        for role in roles : {
          key           = "${account_name}/${role}"
          account       = account_name
          role          = role
          create        = try(local.service_accounts[account_name].create, false)
          adopted_email = try(local.service_accounts[account_name].email, null)
        }
      ]
    ]) : pair.key => pair
    if contains(keys(local.service_accounts), pair.account) && (pair.create || pair.adopted_email != null)
  }
}
