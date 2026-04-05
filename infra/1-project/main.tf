provider "google" {
  project = try(local.config.project_id, null)
  region  = try(local.config.region, null)
}

data "google_project" "current" {
  count      = try(local.config.project_id, null) == null ? 0 : 1
  project_id = local.config.project_id
}

resource "google_project_service" "required" {
  for_each = try(local.config.project_id, null) == null ? toset([]) : local.enabled_apis

  project            = local.config.project_id
  service            = each.value
  disable_on_destroy = false
}
