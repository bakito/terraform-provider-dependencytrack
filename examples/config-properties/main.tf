terraform {
  required_providers {
    dependencytrack = {
      source = "registry.terraform.io/hashicorp/dependencytrack"
    }
  }
}

provider "dependencytrack" {}

resource "dependencytrack_config_property" "general_base-url" {
  group = "general"
  name  = "base.url"
  type  = "URL"
  value = "http://localhost:8080/"
}

