// Copyright (c) Robin Breathe and contributors
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestIsUnknownSentinel(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "sentinel struct",
			input:    UnknownSentinel{Type: types.StringType},
			expected: true,
		},
		{
			name:     "string value",
			input:    "foo",
			expected: false,
		},
		{
			name:     "nil value",
			input:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUnknownSentinel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateUnknownSentinel(t *testing.T) {
	tests := []struct {
		name         string
		input        attr.Value
		expectedType attr.Type
	}{
		{
			name:         "string unknown",
			input:        types.StringUnknown(),
			expectedType: types.StringType,
		},
		{
			name:         "number unknown",
			input:        types.NumberUnknown(),
			expectedType: types.NumberType,
		},
		{
			name:         "bool unknown",
			input:        types.BoolUnknown(),
			expectedType: types.BoolType,
		},
		{
			name:         "map unknown",
			input:        types.MapUnknown(types.StringType),
			expectedType: types.MapType{ElemType: types.StringType},
		},
		{
			name:         "list unknown with string element",
			input:        types.ListUnknown(types.StringType),
			expectedType: types.ListType{ElemType: types.StringType},
		},
		{
			name:         "set unknown with bool element",
			input:        types.SetUnknown(types.BoolType),
			expectedType: types.SetType{ElemType: types.BoolType},
		},
		{
			name:  "object unknown",
			input: types.ObjectUnknown(map[string]attr.Type{"foo": types.StringType, "bar": types.NumberType}),
			expectedType: types.ObjectType{AttrTypes: map[string]attr.Type{
				"foo": types.StringType,
				"bar": types.NumberType,
			}},
		},
		{
			name:         "tuple unknown",
			input:        types.TupleUnknown([]attr.Type{types.StringType, types.NumberType, types.BoolType}),
			expectedType: types.TupleType{ElemTypes: []attr.Type{types.StringType, types.NumberType, types.BoolType}},
		},
		{
			name:         "dynamic unknown",
			input:        types.DynamicUnknown(),
			expectedType: types.DynamicType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentinel := CreateUnknownSentinel(t.Context(), tt.input)
			assert.Equal(t, tt.expectedType, sentinel.Type)
		})
	}
}

func TestUnknownSentinel_ToUnknownValue(t *testing.T) {
	tests := []struct {
		name     string
		sentinel UnknownSentinel
		validate func(t *testing.T, result attr.Value)
	}{
		{
			name:     "string type",
			sentinel: UnknownSentinel{Type: types.StringType},
			validate: func(t *testing.T, result attr.Value) {
				assert.True(t, result.IsUnknown())
				_, ok := result.(types.String)
				assert.True(t, ok, "expected types.String")
			},
		},
		{
			name:     "number type",
			sentinel: UnknownSentinel{Type: types.NumberType},
			validate: func(t *testing.T, result attr.Value) {
				assert.True(t, result.IsUnknown())
				_, ok := result.(types.Number)
				assert.True(t, ok, "expected types.Number")
			},
		},
		{
			name:     "bool type",
			sentinel: UnknownSentinel{Type: types.BoolType},
			validate: func(t *testing.T, result attr.Value) {
				assert.True(t, result.IsUnknown())
				_, ok := result.(types.Bool)
				assert.True(t, ok, "expected types.Bool")
			},
		},
		{
			name:     "map type with element type",
			sentinel: UnknownSentinel{Type: types.MapType{ElemType: types.StringType}},
			validate: func(t *testing.T, result attr.Value) {
				assert.True(t, result.IsUnknown())
				m, ok := result.(types.Map)
				assert.True(t, ok, "expected types.Map")
				assert.Equal(t, types.StringType, m.ElementType(t.Context()))
			},
		},
		{
			name:     "list type with element type",
			sentinel: UnknownSentinel{Type: types.ListType{ElemType: types.NumberType}},
			validate: func(t *testing.T, result attr.Value) {
				assert.True(t, result.IsUnknown())
				l, ok := result.(types.List)
				assert.True(t, ok, "expected types.List")
				assert.Equal(t, types.NumberType, l.ElementType(t.Context()))
			},
		},
		{
			name:     "set type with element type",
			sentinel: UnknownSentinel{Type: types.SetType{ElemType: types.BoolType}},
			validate: func(t *testing.T, result attr.Value) {
				assert.True(t, result.IsUnknown())
				s, ok := result.(types.Set)
				assert.True(t, ok, "expected types.Set")
				assert.Equal(t, types.BoolType, s.ElementType(t.Context()))
			},
		},
		{
			name: "object type with attribute types",
			sentinel: UnknownSentinel{
				Type: types.ObjectType{AttrTypes: map[string]attr.Type{"foo": types.StringType}},
			},
			validate: func(t *testing.T, result attr.Value) {
				assert.True(t, result.IsUnknown())
				o, ok := result.(types.Object)
				assert.True(t, ok, "expected types.Object")
				attrTypes := o.AttributeTypes(t.Context())
				assert.Equal(t, types.StringType, attrTypes["foo"])
			},
		},
		{
			name: "tuple type with element types",
			sentinel: UnknownSentinel{
				Type: types.TupleType{ElemTypes: []attr.Type{types.StringType, types.NumberType}},
			},
			validate: func(t *testing.T, result attr.Value) {
				assert.True(t, result.IsUnknown())
				tu, ok := result.(types.Tuple)
				assert.True(t, ok, "expected types.Tuple")
				elemTypes := tu.ElementTypes(t.Context())
				assert.Equal(t, []attr.Type{types.StringType, types.NumberType}, elemTypes)
			},
		},
		{
			name:     "dynamic type",
			sentinel: UnknownSentinel{Type: types.DynamicType},
			validate: func(t *testing.T, result attr.Value) {
				assert.True(t, result.IsUnknown())
				_, ok := result.(types.Dynamic)
				assert.True(t, ok, "expected types.Dynamic")
			},
		},
		{
			name:     "nil type falls back to dynamic",
			sentinel: UnknownSentinel{Type: nil},
			validate: func(t *testing.T, result attr.Value) {
				assert.True(t, result.IsUnknown())
				_, ok := result.(types.Dynamic)
				assert.True(t, ok, "expected types.Dynamic fallback")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sentinel.ToUnknownValue()
			tt.validate(t, result)
		})
	}
}
