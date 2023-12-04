terraform {
  required_providers {
    dependencytrack = {
      source = "registry.terraform.io/hashicorp/dependencytrack"
    }
  }
}

provider "dependencytrack" {}

data "dependencytrack_repositories" "repos" {}
data "dependencytrack_oidc_groups" "groups" {}
data "dependencytrack_teams" "teams" {}

output "teams" {
  value = data.dependencytrack_teams.teams
}
output "groups" {
  value = data.dependencytrack_oidc_groups.groups
}
