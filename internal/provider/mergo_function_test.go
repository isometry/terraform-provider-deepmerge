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
					value = provider::deepmerge::mergo(false, local.map1, local.map2, local.map3)
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
					value = provider::deepmerge::mergo(false, local.map1, local.map2)
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
					value = provider::deepmerge::mergo(false, local.map1, local.map2)
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
					value = provider::deepmerge::mergo(false, local.map1, local.map2, local.map3, "append")
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
					value = provider::deepmerge::mergo(false, local.map1, local.map2, local.map3, "no_override")
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
					value = provider::deepmerge::mergo(null, null)
				}
				`,
				// The allow_null parameter is set to false
				ExpectError: regexp.MustCompile(`argument must not be null`),
			},
			{
				Config: `
				output "test" {
					value = provider::deepmerge::mergo(true, {}, null)
				}
				`,
				// The allow_null parameter is set to true
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.MapSizeExact(0)),
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
					value = provider::deepmerge::mergo(false, local.map1, local.map2, "no_null_override")
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
					value = provider::deepmerge::mergo(false, local.map1, local.map2, "no_null_override")
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
					value = provider::deepmerge::mergo(false, local.map1, local.map2, local.map3, "no_null_override")
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
					value = provider::deepmerge::mergo(false, local.map1, local.map2, "no_null_override")
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
					value = provider::deepmerge::mergo(false, local.map1, local.map2, "no_null_override")
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
					value = provider::deepmerge::mergo(false, local.map1, local.map2, local.map3, "no_null_override")
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
					value = provider::deepmerge::mergo(false, local.map1, local.map2, local.map3, local.map4, local.map5, "no_null_override")
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
