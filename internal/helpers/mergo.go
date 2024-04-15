package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	"dario.cat/mergo"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Mergo(ctx context.Context, objs []types.Dynamic, opts ...func(*mergo.Config)) (merged types.Dynamic, diags diag.Diagnostics) {
	maps := make([]map[string]any, len(objs))
	for i, obj := range objs {
		// TODO: don't rely on String encoding
		if err := json.Unmarshal([]byte(obj.String()), &maps[i]); err != nil {
			diags.Append(diag.NewErrorDiagnostic(fmt.Sprintf("Error unmarshalling argument %d", i+1), err.Error()))
			return
		}
	}

	dst := maps[0]
	for i, m := range maps[1:] {
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
