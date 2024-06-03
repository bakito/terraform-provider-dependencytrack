terraform {
  required_providers {
    dependencytrack = {
      source = "registry.terraform.io/hashicorp/dependencytrack"
    }
  }
}

provider "dependencytrack" {}

data "dependencytrack_oidc_groups" "oidc_groups" {
}

output "oidc_groups" {
  value = data.dependencytrack_oidc_groups.oidc_groups
}
