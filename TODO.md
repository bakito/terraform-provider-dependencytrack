# TODOs

## refresh

```
dependencytrack_repository.npm["npm-bison"]: Refreshing state... [id=dc458ae7-066a-45bf-8b61-bbf49683e0d1]
2023-12-01T16:35:41.703+0100 [WARN]  Provider "registry.terraform.io/hashicorp/dependencytrack" produced an unexpected new value for dependencytrack_repository.npm["npm-bison"] during refresh.
      - .identifier: was null, but now cty.StringVal("repo.bison-group.com -> registry.npmjs.org")
      - .url: was null, but now cty.StringVal("https://repo.bison-group.com/artifactory/api/npm/npm-cache")
      - .enabled: was null, but now cty.True
      - .resolution_order: was null, but now cty.NumberIntVal(2)
      - .type: was null, but now cty.StringVal("NPM")
```


produced an unexpected new value for  during refresh