[![CodeQL](https://github.com/isometry/terraform-provider-deepmerge/actions/workflows/codeql.yml/badge.svg)](https://github.com/isometry/terraform-provider-deepmerge/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/isometry/terraform-provider-deepmerge)](https://goreportcard.com/report/github.com/isometry/terraform-provider-deepmerge)
[![Terraform Registry Version](https://img.shields.io/badge/dynamic/json?color=blue&label=registry&query=$.version&url=https://registry.terraform.io/v1/providers/isometry/deepmerge)](https://registry.terraform.io/providers/isometry/deepmerge/latest)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/isometry/terraform-provider-deepmerge/total)

# Terraform Provider Deepmerge

Deep merge functionality for complex Terraform configurations, supporting recursive merging of maps and objects with fine-grained control over merge behavior.

## Why Use This Provider?

Terraform's built-in `merge()` function only performs shallow merging, which means nested structures are completely replaced rather than intelligently combined. This provider enables:

- **Recursive merging** of deeply nested structures
- **Flexible merge strategies** with multiple modes of operation
- **List concatenation** instead of replacement
- **Null value handling** for optional configurations
- **Type-safe merging** that preserves your data structure integrity

### Common Use Cases

- **Environment-specific configurations**: Merge base configurations with environment overrides
- **Module defaults**: Provide sensible defaults that users can partially override
- **Multi-team configurations**: Combine configurations from different teams or sources
- **Dynamic resource generation**: Build complex resources from modular configuration fragments

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.8
- [OpenTofu](https://opentofu.org/docs/intro/install/) >= 1.7

## Installation

```hcl
terraform {
  required_providers {
    deepmerge = {
      source  = "isometry/deepmerge"
      version = "~> 1.0"
    }
  }
}

provider "deepmerge" {}
```

## Merge Modes

A distinctive feature of `mergo` is its use of string arguments to control merge strategies. Simply append one or more control strings after your maps to change how the merge operates:

| Mode                                 | Description                                | Use Case                              |
| ------------------------------------ | ------------------------------------------ | ------------------------------------- |
| `"override"` / `"replace"` (default) | Later values replace earlier ones          | Standard configuration layering       |
| `"no_override"`                      | Earlier values are preserved               | Setting immutable defaults            |
| `"no_null_override"`                 | Null values don't replace existing values  | Optional configuration fields         |
| `"append"` / `"append_lists"`        | Lists are concatenated instead of replaced | Accumulating features, rules, or tags |
| `"union"` / `"union_lists"`          | Lists are merged as sets (unique elements) | Deduplicating tags, IPs, or identifiers |

### Examples by Mode

#### Default Behavior (Override)

```hcl
locals {
  config1 = { a = 1, b = { x = 10, y = 20 } }
  config2 = { a = 2, b = { y = 30, z = 40 } }

  result = provider::deepmerge::mergo(local.config1, local.config2)
  # Result: { a = 2, b = { x = 10, y = 30, z = 40 } }
}
```

#### No Override Mode

```hcl
locals {
  defaults    = { timeout = 30, retries = 3, debug = false }
  user_config = { timeout = 60, debug = true, custom = "value" }

  result = provider::deepmerge::mergo(local.defaults, local.user_config, "no_override")
  # Result: { timeout = 30, retries = 3, debug = false, custom = "value" }
  # Note: defaults are preserved, only new keys from user_config are added
}
```

#### No Null Override Mode

```hcl
locals {
  base      = { name = "service", port = 8080, optional_setting = "enabled" }
  overrides = { port = 9090, optional_setting = null }

  result = provider::deepmerge::mergo(local.base, local.overrides, "no_null_override")
  # Result: { name = "service", port = 9090, optional_setting = "enabled" }
  # Note: null doesn't override the existing value
}
```

#### Append Mode

```hcl
locals {
  base_rules = {
    firewall = {
      allowed_ports = [22, 80, 443]
      denied_ips    = ["192.168.1.100"]
    }
  }

  additional_rules = {
    firewall = {
      allowed_ports = [8080, 9090]
      denied_ips    = ["192.168.1.101", "192.168.1.102"]
    }
  }

  result = provider::deepmerge::mergo(local.base_rules, local.additional_rules, "append")
  # Result: {
  #   firewall = {
  #     allowed_ports = [22, 80, 443, 8080, 9090]
  #     denied_ips    = ["192.168.1.100", "192.168.1.101", "192.168.1.102"]
  #   }
  # }
}
```

#### Union Lists Mode

```hcl
locals {
  base_tags = {
    security = {
      allowed_ports = [22, 80, 443]
      tags = ["security", "prod", "critical"]
    }
  }

  additional_tags = {
    security = {
      allowed_ports = [8080, 80, 9090]
      tags = ["monitoring", "prod", "audit"]
    }
  }

  result = provider::deepmerge::mergo(local.base_tags, local.additional_tags, "union_lists")
  # Result: {
  #   security = {
  #     allowed_ports = [22, 80, 443, 8080, 9090]  # Unique elements only
  #     tags = ["security", "prod", "critical", "monitoring", "audit"]  # Duplicates removed
  #   }
  # }
}
```

## Practical Examples

See [docs/functions/mergo.md](docs/functions/mergo.md) for detailed examples.

## Comparison with Terraform's `merge()` Function

| Feature          | Terraform `merge()`        | `deepmerge::mergo()`            |
| ---------------- | -------------------------- | ------------------------------- |
| Shallow merge    | ✅                         | ✅                              |
| Deep merge       | ❌                         | ✅                              |
| List handling    | Replace only               | Replace or append               |
| Null handling    | Overwrites                 | Configurable                    |
| Merge strategies | None                       | Multiple modes                  |
| Performance      | Faster for flat structures | Optimized for nested structures |

### When to Use Each

**Use Terraform's `merge()` when:**

- Working with flat maps
- Performance is critical for large, flat structures
- You want complete replacement of nested values

**Use `deepmerge::mergo()` when:**

- Working with nested configurations
- You need fine-grained control over merge behavior
- Building layered configurations
- Handling optional/nullable values

## Troubleshooting

### Common Issues

- **Unexpected list behavior**: By default, lists are replaced. Use `"append"` mode to concatenate them.

- **Null values overriding**: Use `"no_null_override"` mode to prevent nulls from replacing existing values.

- **Large data structures**: For *very* large nested structures, consider breaking them into smaller, manageable pieces.

## Documentation

- [Provider Documentation](docs/index.md)
- [Function Reference](docs/functions/mergo.md)

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This provider is distributed under the [Mozilla Public License 2.0](LICENSE).
