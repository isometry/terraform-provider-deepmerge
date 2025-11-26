// Copyright (c) Robin Breathe and contributors
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// UnknownSentinel represents an unknown Terraform value during encoding.
// It carries type information needed to reconstruct the appropriate unknown
// value during decoding.
type UnknownSentinel struct {
	Type attr.Type
}

// IsUnknownSentinel checks if a value is our sentinel type.
func IsUnknownSentinel(v any) bool {
	_, ok := v.(UnknownSentinel)
	return ok
}

// CreateUnknownSentinel creates a sentinel from a Terraform unknown value.
// The sentinel preserves type information so the unknown can be reconstructed
// with the correct Terraform type during decoding.
func CreateUnknownSentinel(ctx context.Context, v attr.Value) UnknownSentinel {
	return UnknownSentinel{Type: v.Type(ctx)}
}

// ToUnknownValue converts a sentinel back to a Terraform unknown value
// with the appropriate type.
func (s UnknownSentinel) ToUnknownValue() attr.Value {
	switch t := s.Type.(type) {
	case basetypes.StringType:
		return types.StringUnknown()
	case basetypes.NumberType:
		return types.NumberUnknown()
	case basetypes.BoolType:
		return types.BoolUnknown()
	case basetypes.MapType:
		return types.MapUnknown(t.ElemType)
	case basetypes.ObjectType:
		return types.ObjectUnknown(t.AttrTypes)
	case basetypes.ListType:
		return types.ListUnknown(t.ElemType)
	case basetypes.SetType:
		return types.SetUnknown(t.ElemType)
	case basetypes.TupleType:
		return types.TupleUnknown(t.ElemTypes)
	case basetypes.DynamicType:
		return types.DynamicUnknown()
	default:
		return types.DynamicUnknown()
	}
}
