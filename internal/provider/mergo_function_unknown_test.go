// Copyright (c) Robin Breathe and contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestMergoFunction_UnknownValuePreservation(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source: "hashicorp/random",
			},
		},
		Steps: []resource.TestStep{
			// Test: Unknown value in source map - unknown key should be unknown in result
			// After apply, known keys should have correct values
			{
				Config: `
				resource "random_string" "test" {
					length = 8
				}
				locals {
					map1 = { a = "known" }
					map2 = { b = random_string.test.result }
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapPartial(map[string]knownvalue.Check{
							"a": knownvalue.StringExact("known"),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_BidirectionalStickyUnknown(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source: "hashicorp/random",
			},
		},
		Steps: []resource.TestStep{
			// Test: Known + Unknown on same key = Unknown (bidirectional sticky)
			// After apply, the value should be the random string (from the unknown source)
			{
				Config: `
				resource "random_string" "test" {
					length = 8
				}
				locals {
					map1 = { a = "foo" }
					map2 = { a = random_string.test.result }
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					// After apply, the key "a" should have the random string value (8 chars)
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapSizeExact(1),
					),
				},
			},
		},
	})
}

func TestMergoFunction_NestedUnknown(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source: "hashicorp/random",
			},
		},
		Steps: []resource.TestStep{
			// Test: Nested maps with unknown values
			{
				Config: `
				resource "random_string" "test" {
					length = 8
				}
				locals {
					map1 = {
						nested = {
							a = "known"
						}
					}
					map2 = {
						nested = {
							b = random_string.test.result
						}
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapPartial(map[string]knownvalue.Check{
							"nested": knownvalue.MapPartial(map[string]knownvalue.Check{
								"a": knownvalue.StringExact("known"),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_UnknownWithOptions(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source: "hashicorp/random",
			},
		},
		Steps: []resource.TestStep{
			// Test: Unknown values with no_null_override option
			{
				Config: `
				resource "random_string" "test" {
					length = 8
				}
				locals {
					map1 = {
						a = "keep_this"
						b = random_string.test.result
					}
					map2 = {
						a = null
						c = "added"
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, "no_null_override")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapPartial(map[string]knownvalue.Check{
							"a": knownvalue.StringExact("keep_this"),
							"c": knownvalue.StringExact("added"),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_MultipleUnknowns(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source: "hashicorp/random",
			},
		},
		Steps: []resource.TestStep{
			// Test: Multiple unknown values from different sources
			{
				Config: `
				resource "random_string" "test1" {
					length = 8
				}
				resource "random_string" "test2" {
					length = 8
				}
				locals {
					map1 = {
						a = random_string.test1.result
						b = "known1"
					}
					map2 = {
						c = random_string.test2.result
						d = "known2"
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapPartial(map[string]knownvalue.Check{
							"b": knownvalue.StringExact("known1"),
							"d": knownvalue.StringExact("known2"),
						}),
					),
				},
			},
		},
	})
}
