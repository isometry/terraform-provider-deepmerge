// Copyright (c) Robin Breathe and contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	_ "embed"
	"fmt"
	"reflect"
	"strings"

	"dario.cat/mergo"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

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

//go:embed mergo_function.md
var mergoFunctionDescription string

func (r MergoFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "Deepmerge of maps with mergo semantics",
		MarkdownDescription: mergoFunctionDescription,
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
	with_append := false

	for i, arg := range args {
		if arg.IsNull() {
			continue
		}

		value := arg.UnderlyingValue()

		switch vv := value.(type) {
		case basetypes.StringValue:
			switch option := vv.ValueString(); option {
			case "no_override":
				with_override = false

			case "no_null_override":
				no_null_override = true

			case "override", "replace":
				with_override = true

			case "append", "append_lists":
				opts = append(opts, mergo.WithAppendSlice)
				with_append = true

			default:
				resp.Error = function.NewArgumentFuncError(int64(i), "unrecognised option")
				return
			}

		case basetypes.MapValue, basetypes.ObjectValue:
			if !vv.IsNull() {
				objs = append(objs, arg)
			}

		default:
			typeName := strings.ToLower(strings.TrimSuffix(reflect.TypeOf(value).Name(), "Value"))
			resp.Error = function.NewArgumentFuncError(int64(i), fmt.Sprintf("unsupported %s argument", typeName))
			return
		}
	}

	if with_override {
		opts = append(opts, mergo.WithOverride)
	}

	if no_null_override {
		opts = append(opts, mergo.WithTransformers(noNullOverrideTransformer{with_append: with_append}))
	}

	merged, diags := helpers.Mergo(ctx, objs, opts...)
	if diags.HasError() {
		resp.Error = function.FuncErrorFromDiags(ctx, diags)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, &merged))
}

type noNullOverrideTransformer struct {
	with_append bool
}

func (t noNullOverrideTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ.Kind() == reflect.Map {
		return func(dst, src reflect.Value) error {
			deepMergeMaps(dst, src, t.with_append)
			return nil
		}
	}
	return nil
}

func deepMergeMaps(dst, src reflect.Value, appendSlice bool) reflect.Value {
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
			newValue := deepMergeMaps(dstElem, srcElem, appendSlice)
			dst.SetMapIndex(key, newValue)
		} else if !srcElem.IsValid() { // skip override of nil values
			continue
		} else if srcElem.Kind() == reflect.Slice && dstElem.Kind() == reflect.Slice && appendSlice { // handle append
			dst.SetMapIndex(key, reflect.AppendSlice(dstElem, srcElem))
		} else {
			dst.SetMapIndex(key, srcElem)
		}
	}

	return dst
}
