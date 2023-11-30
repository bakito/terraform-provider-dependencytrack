terraform {
  required_providers {
    dependencytrack = {
      source = "registry.terraform.io/hashicorp/dependencytrack"
    }
  }
}

provider "dependencytrack" {}


resource "dependencytrack_repository" "foo" {
  url        = "https://foo.bar"
  identifier = "foo"
  enabled    = true
  type       = "CPAN"
}
