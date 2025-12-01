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
			{
				Config: `
				variable "obj" {
					type = object({
						x1 = any
						x2 = any
					})
					default = {
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
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, var.obj)
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
		},
	})
}

func TestMergoFunction_Append_NoNullOverride(t *testing.T) {
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
							y1 = "foo"
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
							y1 = null
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
					value = provider::deepmerge::mergo(local.map1, local.map2, local.map3, "append", "no_null_override")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"x1": knownvalue.MapExact(map[string]knownvalue.Check{
								"y1": knownvalue.StringExact("foo"),
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
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapSizeExact(0),
					),
				},
			},
			{
				Config: `
				output "test" {
					value = provider::deepmerge::mergo({}, null)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapSizeExact(0),
					),
				},
			},
			{
				Config: `
				variable "null_map" {
					type    = map(any)
					default = null
				}
				output "test" {
					value = provider::deepmerge::mergo(var.null_map)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapSizeExact(0),
					),
				},
			},
			{
				Config: `
				variable "null_map" {
					type    = map(any)
					default = null
				}
				output "test" {
					value = provider::deepmerge::mergo({}, var.null_map)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapSizeExact(0),
					),
				},
			},
		},
	})
}

func TestMergoFunction_NoOverrideWithNull(t *testing.T) {
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
						a = "foo"
						b = "bar"
					}
					map2 = {
						b = "bam"
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, "no_null_override")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"a": knownvalue.StringExact("foo"),
							"b": knownvalue.StringExact("bam"),
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
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, "no_null_override")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"a": knownvalue.StringExact("foo"),
							"b": knownvalue.StringExact("bar"),
						}),
					),
				},
			},
			{
				Config: `
				locals {
					map1 = {
						x1 = {
							y1 = false
							y2 = 1
						}
					}
					map2 = {
						x1 = {
							y2 = 1
							y3 = [4, 5, 6]
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
							y1 = true
							y3 = [1, 2, 3]
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
					value = provider::deepmerge::mergo(local.map1, local.map2, local.map3, "no_null_override")
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
						a = "foo"
						b = "bar"
					}
					map2 = {
						a = null
						b = "bam"
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, "no_null_override")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"a": knownvalue.StringExact("foo"),
							"b": knownvalue.StringExact("bam"),
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
						b = null
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, "no_null_override")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"a": knownvalue.StringExact("foo"),
							"b": knownvalue.StringExact("bar"),
						}),
					),
				},
			},
			{
				Config: `
				locals {
					map1 = {
						x1 = {
							y1 = false
							y2 = 1
						}
					}
					map2 = {
						x1 = {
							y2 = 1
							y3 = [4, 5, 6]
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
							y1 = true
							y3 = [1, 2, 3]
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
					value = provider::deepmerge::mergo(local.map1, local.map2, local.map3, "no_null_override")
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
						x1 = {
							y2 = {
								z1 = 1
								z2 = 2
							}
						}
						x2 = {
							s1 = "hello"
							s2 = "world"
							s3 = null
						}
						x3 = {
							t1 = "foo"
						}
					}
					map2 = {
						x1 = {
							y2 = {
								z2 = null
								z3 = 3
							}
							y2 = "bar"
						}
						x3 = {
							t1 = null
						}
					}
					map3 = {
						x1 = {
							y2 = null
						}
					}
					map4 = {
						x1 = {
							y1 = 4
						}
					}
					map5 = {
						x1 = {
							y2 = {
								z1 = {
									n1 = 1
									n2 = {
										m1 = 1
									}
								}
								z2 = null
								z3 = null
								z4 = 4
							}
						}
						x2 = {
							s2 = "mergo"
							s4 = "today"
						}
						x3 = {
							t2 = "foz"
						}
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, local.map3, local.map4, local.map5, "no_null_override")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"x1": knownvalue.MapExact(map[string]knownvalue.Check{
								"y1": knownvalue.Int64Exact(4),
								"y2": knownvalue.MapExact(map[string]knownvalue.Check{
									"z1": knownvalue.MapExact(map[string]knownvalue.Check{
										"n1": knownvalue.Int64Exact(1),
										"n2": knownvalue.MapExact(map[string]knownvalue.Check{
											"m1": knownvalue.Int64Exact(1),
										}),
									}),
									"z2": knownvalue.Null(),
									"z3": knownvalue.Null(),
									"z4": knownvalue.Int64Exact(4),
								}),
							}),
							"x2": knownvalue.MapExact(map[string]knownvalue.Check{
								"s1": knownvalue.StringExact("hello"),
								"s2": knownvalue.StringExact("mergo"),
								"s3": knownvalue.Null(),
								"s4": knownvalue.StringExact("today"),
							}),
							"x3": knownvalue.MapExact(map[string]knownvalue.Check{
								"t1": knownvalue.StringExact("foo"),
								"t2": knownvalue.StringExact("foz"),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_InvalidType(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = provider::deepmerge::mergo(true)
				}
				`,
				ExpectError: regexp.MustCompile(`unsupported bool argument`),
			},
			{
				Config: `
				output "test" {
					value = provider::deepmerge::mergo(99.9)
				}
				`,
				ExpectError: regexp.MustCompile(`unsupported number argument`),
			},
			{
				Config: `
				output "test" {
					value = provider::deepmerge::mergo(["a", "b", "c"])
				}
				`,
				ExpectError: regexp.MustCompile(`unsupported tuple argument`),
			},

			{
				Config: `
				output "test" {
					value = provider::deepmerge::mergo(tolist(["a", "b", "c"]))
				}
				`,
				ExpectError: regexp.MustCompile(`unsupported list argument`),
			},
			{
				Config: `
				output "test" {
					value = provider::deepmerge::mergo(toset(["a", "b", "c"]))
				}
				`,
				ExpectError: regexp.MustCompile(`unsupported set argument`),
			},
		},
	})
}

func TestMergoFunction_UnionLists(t *testing.T) {
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
						tags = ["app", "terraform", "prod"]
						ports = [80, 443, 22]
						features = {
							security = ["ssl", "firewall"]
							monitoring = ["logs", "metrics"]
						}
					}
					map2 = {
						tags = ["monitoring", "prod", "security"]
						ports = [8080, 80, 9090]
						features = {
							security = ["encryption", "ssl"]
							monitoring = ["alerts", "metrics"]
							caching = ["redis"]
						}
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, "union_lists")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"tags": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("app"),
								knownvalue.StringExact("terraform"),
								knownvalue.StringExact("prod"),
								knownvalue.StringExact("monitoring"),
								knownvalue.StringExact("security"),
							}),
							"ports": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.Int64Exact(80),
								knownvalue.Int64Exact(443),
								knownvalue.Int64Exact(22),
								knownvalue.Int64Exact(8080),
								knownvalue.Int64Exact(9090),
							}),
							"features": knownvalue.MapExact(map[string]knownvalue.Check{
								"security": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("ssl"),
									knownvalue.StringExact("firewall"),
									knownvalue.StringExact("encryption"),
								}),
								"monitoring": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("logs"),
									knownvalue.StringExact("metrics"),
									knownvalue.StringExact("alerts"),
								}),
								"caching": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("redis"),
								}),
							}),
						}),
					),
				},
			},
			{
				Config: `
				locals {
					base = {
						allowed_cidrs = ["10.0.0.0/8", "192.168.1.0/24"]
						environments = ["dev", "staging"]
					}
					overrides = {
						allowed_cidrs = ["172.16.0.0/12", "10.0.0.0/8"]
						environments = ["prod", "staging"]
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.base, local.overrides, "union_lists")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"allowed_cidrs": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("10.0.0.0/8"),
								knownvalue.StringExact("192.168.1.0/24"),
								knownvalue.StringExact("172.16.0.0/12"),
							}),
							"environments": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("dev"),
								knownvalue.StringExact("staging"),
								knownvalue.StringExact("prod"),
							}),
						}),
					),
				},
			},
			{
				Config: `
				locals {
					map1 = {
						mixed_types = [1, "string", true]
						numbers = [1, 2, 3]
					}
					map2 = {
						mixed_types = [true, "another", 1]
						numbers = [3, 4, 5]
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, "union_lists")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"mixed_types": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.Int64Exact(1),
								knownvalue.StringExact("string"),
								knownvalue.Bool(true),
								knownvalue.StringExact("another"),
							}),
							"numbers": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.Int64Exact(1),
								knownvalue.Int64Exact(2),
								knownvalue.Int64Exact(3),
								knownvalue.Int64Exact(4),
								knownvalue.Int64Exact(5),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_UnionListsWithNullOverride(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
					base = {
						tags = ["app", "base"]
						important_setting = "keep_this"
					}
					overrides = {
						tags = ["override", "app"]
						important_setting = null
						new_field = "added"
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.base, local.overrides, "union_lists", "no_null_override")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"tags": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("app"),
								knownvalue.StringExact("base"),
								knownvalue.StringExact("override"),
							}),
							"important_setting": knownvalue.StringExact("keep_this"),
							"new_field":         knownvalue.StringExact("added"),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_UnionListsWithDefaultNullBehavior(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
					base = {
						tags = ["app", "base"]
						important_setting = "replace_this"
					}
					overrides = {
						tags = ["override", "app"]
						important_setting = null
						new_field = "added"
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.base, local.overrides, "union_lists")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"tags": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("app"),
								knownvalue.StringExact("base"),
								knownvalue.StringExact("override"),
							}),
							"new_field": knownvalue.StringExact("added"),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_Union_NoNullOverride(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
					base = {
						tags = ["app", "base"]
						important_setting = "keep_this"
					}
					overrides = {
						tags = ["override", "app"]
						important_setting = null
						new_field = "added"
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.base, local.overrides, "union", "no_null_override")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"tags": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("app"),
								knownvalue.StringExact("base"),
								knownvalue.StringExact("override"),
							}),
							"important_setting": knownvalue.StringExact("keep_this"),
							"new_field":         knownvalue.StringExact("added"),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_UnionListsWithNonComparableElements(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
					base = {
						nested_lists = [
							["a", "b"],
							["c", "d"]
						]
					}
					additional = {
						nested_lists = [
							["a", "b"],
							["e", "f"]
						]
					}
				}
				output "test" {
					value = provider::deepmerge::mergo(local.base, local.additional, "union_lists")
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"nested_lists": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("a"),
									knownvalue.StringExact("b"),
								}),
								knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("c"),
									knownvalue.StringExact("d"),
								}),
								knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("e"),
									knownvalue.StringExact("f"),
								}),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_VariableOption(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				variable "merge_mode" {
					type    = string
					default = "union"
				}
				locals {
					map1 = { a = [1, 2, 3] }
					map2 = { a = [3, 4, 5] }
				}
				output "test" {
					value = provider::deepmerge::mergo(local.map1, local.map2, var.merge_mode)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"a": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.Int64Exact(1),
								knownvalue.Int64Exact(2),
								knownvalue.Int64Exact(3),
								knownvalue.Int64Exact(4),
								knownvalue.Int64Exact(5),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestMergoFunction_NestedVariableOption(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				variable "merge_options" {
					type = object({
						mode = optional(string, "no_override")
					})
					default = {}
				}
				locals {
					defaults    = { timeout = 30, retries = 3 }
					user_config = { timeout = 60 }
				}
				output "test" {
					value = provider::deepmerge::mergo(local.defaults, local.user_config, var.merge_options.mode)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"timeout": knownvalue.Int64Exact(30), // no_override keeps original
							"retries": knownvalue.Int64Exact(3),
						}),
					),
				},
			},
		},
	})
}
