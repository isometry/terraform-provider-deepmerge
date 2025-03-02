package helpers

import (
	"context"
	"fmt"

	"dario.cat/mergo"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Mergo(ctx context.Context, objs []types.Dynamic, opts ...func(*mergo.Config)) (merged types.Dynamic, diags diag.Diagnostics) {
	maps := make([]map[string]any, len(objs))
	for i, obj := range objs {
		x, err := EncodeValue(obj)
		if err != nil {
			diags.Append(diag.NewErrorDiagnostic(fmt.Sprintf("Error encoding argument %d", i+1), err.Error()))
			return
		}

		if y, ok := x.(map[string]any); !ok {
			diags.Append(diag.NewErrorDiagnostic(fmt.Sprintf("Error converting argument %d to map", i+1), fmt.Sprintf("unexpected type: %T for value %#v", x, x)))
			return
		} else {
			maps[i] = y
		}
	}

	dst := make(map[string]any)
	for i, m := range maps {
		if err := mergo.Merge(&dst, m, opts...); err != nil {
			diags.Append(diag.NewErrorDiagnostic(fmt.Sprintf("Error merging argument %d", i+1), err.Error()))
			return
		}
	}

	var mergedValue attr.Value

	mergedValue, diags = DecodeScalar(ctx, dst)
	if diags.HasError() {
		return
	}

	return types.DynamicValue(mergedValue), nil
}
