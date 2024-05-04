// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestMergoFunction_Default(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
					map1 = {
						x1 = {
							y1 = true
							y2 = 1
						}
					}
					map2 = {
						x1 = {
							y2 = 2
							y3 = [1, 2, 3]
						}
						x2 = {
							y4 = {
								a = "hello"
								b = "world"
							}
						}
					}
					map3 = {
						x1 = {
							y1 = false
							y3 = [4, 5, 6]
						}
						x2 = {
							y4 = {
								b = "mergo"
								c = ["a", 2, ["b"]]
							}
						}
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, local.map3)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"x1": knownvalue.MapExact(map[string]knownvalue.Check{
								"y1": knownvalue.Bool(false),
								"y2": knownvalue.Int64Exact(2),
								"y3": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.Int64Exact(4),
									knownvalue.Int64Exact(5),
									knownvalue.Int64Exact(6),
								}),
							}),
							"x2": knownvalue.MapExact(map[string]knownvalue.Check{
								"y4": knownvalue.MapExact(map[string]knownvalue.Check{
									"a": knownvalue.StringExact("hello"),
									"b": knownvalue.StringExact("mergo"),
									"c": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("a"),
										knownvalue.Int64Exact(2),
										knownvalue.ListExact([]knownvalue.Check{
											knownvalue.StringExact("b"),
										}),
									}),
								}),
							}),
						}),
					),
				},
			},
			{
				Config: `
				locals {
					map1 = {
						a = null
						b = "foo"
					}
					map2 = {
						a = "bar"
						b = "baz"
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"a": knownvalue.StringExact("bar"),
							"b": knownvalue.StringExact("baz"),
						}),
					),
				},
			},
			{
				Config: `
				locals {
					map1 = {
						a = "foo"
						b = "bar"
					}
					map2 = {
						a = null
						b = "bam"
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"a": knownvalue.Null(),
							"b": knownvalue.StringExact("bam"),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_Append(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
					map1 = {
						x1 = {
							y1 = true
							y2 = 1
						}
					}
					map2 = {
						x1 = {
							y2 = 2
							y3 = [1, 2, 3]
						}
						x2 = {
							y4 = {
								a = "hello"
								b = "world"
							}
						}
					}
					map3 = {
						x1 = {
							y1 = false
							y3 = [4, 5, 6]
						}
						x2 = {
							y4 = {
								b = "mergo"
								c = ["a", 2, ["b"]]
							}
						}
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, local.map3, "append")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"x1": knownvalue.MapExact(map[string]knownvalue.Check{
								"y1": knownvalue.Bool(false),
								"y2": knownvalue.Int64Exact(2),
								"y3": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.Int64Exact(1),
									knownvalue.Int64Exact(2),
									knownvalue.Int64Exact(3),
									knownvalue.Int64Exact(4),
									knownvalue.Int64Exact(5),
									knownvalue.Int64Exact(6),
								}),
							}),
							"x2": knownvalue.MapExact(map[string]knownvalue.Check{
								"y4": knownvalue.MapExact(map[string]knownvalue.Check{
									"a": knownvalue.StringExact("hello"),
									"b": knownvalue.StringExact("mergo"),
									"c": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("a"),
										knownvalue.Int64Exact(2),
										knownvalue.ListExact([]knownvalue.Check{
											knownvalue.StringExact("b"),
										}),
									}),
								}),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_NoOverride(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
					map1 = {
						x1 = {
							y1 = true
							y2 = 1
						}
					}
					map2 = {
						x1 = {
							y2 = 2
							y3 = [1, 2, 3]
						}
						x2 = {
							y4 = {
								a = "hello"
								b = "world"
							}
						}
					}
					map3 = {
						x1 = {
							y1 = false
							y3 = [4, 5, 6]
						}
						x2 = {
							y4 = {
								b = "mergo"
								c = ["a", 2, ["b"]]
							}
						}
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, local.map3, "no_override")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"x1": knownvalue.MapExact(map[string]knownvalue.Check{
								"y1": knownvalue.Bool(true),
								"y2": knownvalue.Int64Exact(1),
								"y3": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.Int64Exact(1),
									knownvalue.Int64Exact(2),
									knownvalue.Int64Exact(3),
								}),
							}),
							"x2": knownvalue.MapExact(map[string]knownvalue.Check{
								"y4": knownvalue.MapExact(map[string]knownvalue.Check{
									"a": knownvalue.StringExact("hello"),
									"b": knownvalue.StringExact("world"),
									"c": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("a"),
										knownvalue.Int64Exact(2),
										knownvalue.ListExact([]knownvalue.Check{
											knownvalue.StringExact("b"),
										}),
									}),
								}),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_Null(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = provider::deepmerge::mergo(null)
				}
				`,
				// The parameter does not enable AllowNullValue
				ExpectError: regexp.MustCompile(`argument must not be null`),
			},
		},
	})
}
