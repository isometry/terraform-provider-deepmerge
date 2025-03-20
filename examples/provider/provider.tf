terraform {
  required_providers {
    deepmerge = {
      source = "isometry/deepmerge"
    }
  }
}

provider "deepmerge" {}

locals {
  map1 = {
    a = {
      x = [1, 2, 3]
      y = false
    }
    b = {
      s = "hello, world"
      n = 17
    }
  }
  map2 = {
    a = { x = [4, 5, 6] }
    b = { n = 42 }
  }

  merged = provider::deepmerge::mergo(local.map1, local.map2)
}
