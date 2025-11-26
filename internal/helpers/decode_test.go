package helpers

import (
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDecodeScalar(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected attr.Value
		hasError bool
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: types.DynamicNull(),
			hasError: false,
		},
		{
			name:     "float64 value",
			input:    3.14,
			expected: types.NumberValue(big.NewFloat(3.14)),
			hasError: false,
		},
		{
			name:     "bool value",
			input:    true,
			expected: types.BoolValue(true),
			hasError: false,
		},
		{
			name:     "string value",
			input:    "test",
			expected: types.StringValue("test"),
			hasError: false,
		},
		{
			name:  "slice value",
			input: []any{1.23, "foo", false},
			expected: func() attr.Value {
				value, _ := types.TupleValue(
					[]attr.Type{types.NumberType, types.StringType, types.BoolType},
					[]attr.Value{
						types.NumberValue(big.NewFloat(1.23)),
						types.StringValue("foo"),
						types.BoolValue(false),
					})
				return value
			}(),
			hasError: false,
		},
		{
			name: "map value",
			input: map[string]any{
				"key1": nil,
				"key2": 3.14,
				"key3": true,
				"key4": "test",
			},
			expected: func() attr.Value {
				value, _ := types.ObjectValue(
					map[string]attr.Type{
						"key1": types.DynamicType,
						"key2": types.NumberType,
						"key3": types.BoolType,
						"key4": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.DynamicNull(),
						"key2": types.NumberValue(big.NewFloat(3.14)),
						"key3": types.BoolValue(true),
						"key4": types.StringValue("test"),
					},
				)
				return value
			}(),
			hasError: false,
		},
		{
			name:     "unexpected type",
			input:    struct{}{},
			expected: nil,
			hasError: true,
		},
		{
			name:     "unknown string sentinel",
			input:    UnknownSentinel{Type: types.StringType},
			expected: types.StringUnknown(),
			hasError: false,
		},
		{
			name:     "unknown number sentinel",
			input:    UnknownSentinel{Type: types.NumberType},
			expected: types.NumberUnknown(),
			hasError: false,
		},
		{
			name:     "unknown bool sentinel",
			input:    UnknownSentinel{Type: types.BoolType},
			expected: types.BoolUnknown(),
			hasError: false,
		},
		{
			name:     "unknown map sentinel",
			input:    UnknownSentinel{Type: types.MapType{ElemType: types.StringType}},
			expected: types.MapUnknown(types.StringType),
			hasError: false,
		},
		{
			name:     "unknown list sentinel",
			input:    UnknownSentinel{Type: types.ListType{ElemType: types.NumberType}},
			expected: types.ListUnknown(types.NumberType),
			hasError: false,
		},
		{
			name:     "unknown dynamic sentinel",
			input:    UnknownSentinel{Type: types.DynamicType},
			expected: types.DynamicUnknown(),
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, diags := DecodeScalar(t.Context(), tt.input)
			if tt.hasError {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
				assert.Equal(t, tt.expected, decoded)
			}
		})
	}
}
