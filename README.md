[![CodeQL](https://github.com/isometry/terraform-provider-deepmerge/actions/workflows/codeql.yml/badge.svg)](https://github.com/isometry/terraform-provider-deepmerge/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/isometry/terraform-provider-deepmerge)](https://goreportcard.com/report/github.com/isometry/terraform-provider-deepmerge)

# Terraform Provider Deepmerge

Deepmerge functions for Terraform 1.8+ and OpenTofu 1.7+.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.8
- [OpenTofu](https://opentofu.org/docs/intro/install/) >= 1.7

## Using the provider

```hcl
terraform {
  required_providers {
    deepmerge = {
      source = "isometry/deepmerge"
      version = "1.0.0"
    }
  }
}

provider "deepmerge" {}

output "example" {
  value = provider::deepmerge::mergo(local.map1, local.map2, local.map3)
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
