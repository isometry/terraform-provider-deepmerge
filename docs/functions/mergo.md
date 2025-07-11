---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "mergo function - terraform-provider-deepmerge"
subcategory: ""
description: |-
  Deepmerge of maps with mergo semantics
---

# function: mergo

## Overview

Unlike Terraform's built-in `merge()` function which only performs shallow merging, `mergo` recursively traverses nested structures. The function accepts an arbitrary number of maps or objects, and returns a single merged result.

## Merge Modes

A distinctive feature of `mergo` is its use of string arguments to control merge strategies. Simply append one or more control strings after your maps to change how the merge operates:

| Mode                                 | Description                                | Use Case                                |
| ------------------------------------ | ------------------------------------------ | --------------------------------------- |
| `"override"` / `"replace"` (default) | Later values replace earlier ones          | Standard configuration layering         |
| `"no_override"`                      | Earlier values are preserved               | Setting immutable defaults              |
| `"no_null_override"`                 | Null values don't replace existing values  | Optional configuration fields           |
| `"append"` / `"append_lists"`        | Lists are concatenated instead of replaced | Accumulating features, rules, or tags   |
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

### Multi-Environment Configuration

```hcl
locals {
  # Base configuration for all environments
  base_app_config = {
    app = {
      name = "web-service"
      health_check = {
        path     = "/health"
        interval = 30
        timeout  = 5
      }
      logging = {
        level  = "info"
        format = "json"
      }
    }
  }

  # Environment-specific configurations
  env_configs = {
    dev = {
      app = {
        replicas = 1
        logging  = { level = "debug" }
      }
    }
    staging = {
      app = {
        replicas     = 2
        health_check = { interval = 15 }
      }
    }
    prod = {
      app = {
        replicas     = 5
        health_check = { interval = 10, timeout = 3 }
        logging      = { level = "warn" }
        ssl          = { enabled = true, strict = true }
      }
    }
  }

  # Get the final configuration for the current environment
  final_config = provider::deepmerge::mergo(
    local.base_app_config,
    lookup(local.env_configs, var.environment, {})
  )
}
```

### Kubernetes-Style Resource Merging

```hcl
locals {
  # Base deployment specification
  deployment_base = {
    metadata = {
      labels = {
        app        = "myapp"
        version    = "1.0.0"
        managed_by = "terraform"
      }
      annotations = {
        "deployment.kubernetes.io/revision" = "1"
      }
    }
    spec = {
      selector = {
        match_labels = {
          app = "myapp"
        }
      }
      template = {
        spec = {
          containers = [{
            name  = "app"
            image = "myapp:1.0.0"
            ports = [{ container_port = 8080 }]
            env = [
              { name = "LOG_LEVEL", value = "info" }
            ]
          }]
        }
      }
    }
  }

  # Production overrides with additional configurations
  prod_overrides = {
    metadata = {
      labels = {
        environment = "production"
        tier        = "backend"
      }
      annotations = {
        "prometheus.io/scrape" = "true"
        "prometheus.io/port"   = "9090"
      }
    }
    spec = {
      replicas = 3
      template = {
        spec = {
          containers = [{
            resources = {
              requests = { memory = "256Mi", cpu = "100m" }
              limits   = { memory = "512Mi", cpu = "200m" }
            }
            env = [
              { name = "DATABASE_URL", value = "postgres://prod-db:5432/myapp" },
              { name = "CACHE_ENABLED", value = "true" },
            ]
          }]
        }
      }
    }
  }

  # Merge with append mode for arrays
  deployment = provider::deepmerge::mergo(
    local.deployment_base,
    local.prod_overrides,
    "append"
  )
}
```

### Tag Management with Union Lists

```hcl
locals {
  # Base tags from organization standards
  org_tags = {
    global_tags = ["terraform", "managed", "compliance"]
    security_tags = ["encrypted", "monitored"]
  }

  # Team-specific tags
  team_tags = {
    global_tags = ["backend", "api", "terraform"]  # Some overlap
    team_tags = ["payments", "critical"]
  }

  # Environment-specific tags
  env_tags = {
    global_tags = ["production", "managed"]  # More overlap
    security_tags = ["audited", "encrypted"]  # More overlap
  }

  # Merge all tags with automatic deduplication
  final_tags = provider::deepmerge::mergo(
    local.org_tags,
    local.team_tags,
    local.env_tags,
    "union_lists"
  )
  # Result: {
  #   global_tags = ["terraform", "managed", "compliance", "backend", "api", "production"]
  #   security_tags = ["encrypted", "monitored", "audited"]
  #   team_tags = ["payments", "critical"]
  # }
}
```



## Signature

<!-- signature generated by tfplugindocs -->
```text
mergo(maps dynamic...) dynamic
```

## Arguments

<!-- arguments generated by tfplugindocs -->

<!-- variadic argument generated by tfplugindocs -->
1. `maps` (Variadic, Dynamic, Nullable) Maps to merge
