package helpers

import (
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestEncodeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    attr.Value
		expected any
	}{
		{
			name:     "null value",
			input:    types.StringNull(),
			expected: nil,
		},
		{
			name:     "string value",
			input:    types.StringValue("test"),
			expected: "test",
		},
		{
			name:     "number value",
			input:    types.NumberValue(big.NewFloat(123.45)),
			expected: 123.45,
		},
		{
			name:     "bool value",
			input:    types.BoolValue(true),
			expected: true,
		},
		{
			name: "object value",
			input: func() attr.Value {
				value, _ := types.ObjectValue(map[string]attr.Type{
					"key1": types.StringType,
					"key2": types.BoolType,
				}, map[string]attr.Value{
					"key1": types.StringValue("value"),
					"key2": types.BoolValue(false),
				})
				return value
			}(),
			expected: map[string]any{
				"key1": "value",
				"key2": false,
			},
		},
		{
			name: "tuple value",
			input: func() attr.Value {
				value, _ := types.TupleValue([]attr.Type{
					types.StringType,
					types.NumberType,
				}, []attr.Value{
					types.StringValue("test"),
					types.NumberValue(big.NewFloat(123.45)),
				})
				return value
			}(),
			expected: []any{"test", 123.45},
		},
		{
			name: "map value",
			input: types.MapValueMust(types.StringType, map[string]attr.Value{
				"key": types.StringValue("value"),
			}),
			expected: map[string]any{
				"key": "value",
			},
		},
		{
			name: "list value",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("value1"),
				types.StringValue("value2"),
			}),
			expected: []any{"value1", "value2"},
		},
		{
			name: "set value",
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("value1"),
				types.StringValue("value2"),
			}),
			expected: []any{"value1", "value2"},
		},
		{
			name:     "dynamic value",
			input:    types.DynamicValue(types.StringValue("dynamic")),
			expected: "dynamic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodeValue(t.Context(), tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncodeValue_Unknown(t *testing.T) {
	tests := []struct {
		name         string
		input        attr.Value
		expectedType attr.Type
	}{
		{
			name:         "unknown string",
			input:        types.StringUnknown(),
			expectedType: types.StringType,
		},
		{
			name:         "unknown number",
			input:        types.NumberUnknown(),
			expectedType: types.NumberType,
		},
		{
			name:         "unknown bool",
			input:        types.BoolUnknown(),
			expectedType: types.BoolType,
		},
		{
			name:         "unknown map",
			input:        types.MapUnknown(types.StringType),
			expectedType: types.MapType{ElemType: types.StringType},
		},
		{
			name:         "unknown list",
			input:        types.ListUnknown(types.StringType),
			expectedType: types.ListType{ElemType: types.StringType},
		},
		{
			name:         "unknown dynamic",
			input:        types.DynamicUnknown(),
			expectedType: types.DynamicType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodeValue(t.Context(), tt.input)
			assert.NoError(t, err)
			assert.True(t, IsUnknownSentinel(result), "expected UnknownSentinel")

			sentinel, ok := result.(UnknownSentinel)
			assert.True(t, ok, "expected result to be UnknownSentinel")
			assert.Equal(t, tt.expectedType, sentinel.Type)
		})
	}
}
