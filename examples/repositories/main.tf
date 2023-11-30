terraform {
  required_providers {
    dependencytrack = {
      source = "registry.terraform.io/hashicorp/dependencytrack"
    }
  }
}

provider "dependencytrack" {}

data "dependencytrack_repositories" "repos" {}

output "repos" {
  value = data.dependencytrack_repositories.repos
}
