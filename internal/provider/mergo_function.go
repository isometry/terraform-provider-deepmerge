// Copyright (c) Robin Breathe and contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"reflect"

	"dario.cat/mergo"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/isometry/terraform-provider-deepmerge/internal/helpers"
)

var (
	_ function.Function = MergoFunction{}
)

func NewMergoFunction() function.Function {
	return MergoFunction{}
}

type MergoFunction struct{}

func (r MergoFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "mergo"
}

func (r MergoFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "Deepmerge of maps with mergo semantics",
		MarkdownDescription: "`mergo` takes an arbitrary number of maps or objects, and returns a single map or object that contains a recursively merged set of elements from all arguments.\n\nBy default, values in later arguments override those in earlier arguments, in accordance with standard `mergo` semantics. The merge behaviour can be adjusted by passing additional string arguments to the function:\n\n* `\"override\"` or `\"replace\"` (default): New values override existing values.\n* `\"no_override\"`: New values do not override existing values.\n* `\"no_null_override\"`: Explicit null values do not override existing values.\n* `\"append\"` or `\"append_lists\"`: Append list values instead of replacing them.",
		VariadicParameter: function.DynamicParameter{
			Name:                "maps",
			MarkdownDescription: "Maps to merge",
			AllowNullValue:      true,
		},
		Return: function.DynamicReturn{},
	}
}

func (r MergoFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	args := make([]types.Dynamic, 0)

	if resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &args)); resp.Error != nil {
		return
	}

	if len(args) == 0 {
		resp.Error = function.NewFuncError("at least one map must be provided")
		return
	}

	objs := make([]types.Dynamic, 0)
	opts := make([]func(*mergo.Config), 0)
	with_override := true
	no_null_override := false

	for i, arg := range args {
		if arg.IsNull() || arg.IsUnderlyingValueNull() {
			continue
		}
		value := arg.UnderlyingValue()
		switch ty := value.Type(ctx); ty {
		case types.StringType:
			switch option := arg.String(); option {
			case `"no_override"`:
				with_override = false
			case `"no_null_override"`:
				no_null_override = true
			case `"override"`, `"replace"`:
				with_override = true
			case `"append"`, `"append_lists"`:
				opts = append(opts, mergo.WithAppendSlice)
			default:
				resp.Error = function.NewArgumentFuncError(int64(i), "unrecognised option")
				return
			}
		default:
			objs = append(objs, arg)
		}
	}

	if len(objs) == 0 {
		resp.Error = function.NewFuncError("at least one non-null map required")
		return
	}

	if with_override {
		opts = append(opts, mergo.WithOverride)
	}

	if no_null_override {
		opts = append(opts, mergo.WithTransformers(noNullOverrideTransformer{}))
	}

	merged, diags := helpers.Mergo(ctx, objs, opts...)
	if diags.HasError() {
		resp.Error = function.FuncErrorFromDiags(ctx, diags)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, &merged))
}

type noNullOverrideTransformer struct {
}

func (t noNullOverrideTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ.Kind() == reflect.Map {
		return func(dst, src reflect.Value) error {
			deepMergeMaps(dst, src)
			return nil
		}
	}
	return nil
}

func deepMergeMaps(dst, src reflect.Value) reflect.Value {
	for _, key := range src.MapKeys() {
		srcElem := src.MapIndex(key)
		dstElem := dst.MapIndex(key)

		// Unwrap the interfaces of srcElem and dstElem
		if srcElem.Kind() == reflect.Interface {
			srcElem = srcElem.Elem()
		}

		if dstElem.Kind() == reflect.Interface {
			dstElem = dstElem.Elem()
		}

		if srcElem.Kind() == reflect.Map && dstElem.Kind() == reflect.Map {
			// recursive call
			newValue := deepMergeMaps(dstElem, srcElem)
			dst.SetMapIndex(key, newValue)
		} else {
			if !srcElem.IsValid() { // skip override of nil values
				continue
			}
			dst.SetMapIndex(key, srcElem)
		}
	}
	return dst
}
