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
			AllowUnknownValues:  true,
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
	with_union := false

	for i, arg := range args {
		if arg.IsNull() {
			continue
		}

		// Handle unknown arguments - return unknown result
		if arg.IsUnknown() {
			resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, types.DynamicUnknown()))
			return
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

			case "union", "union_lists":
				with_union = true

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

	if no_null_override || with_union {
		opts = append(opts, mergo.WithTransformers(customTransformer{
			with_append:        with_append,
			with_union:         with_union,
			with_null_override: !no_null_override,
		}))
	}

	merged, diags := helpers.Mergo(ctx, objs, opts...)
	if diags.HasError() {
		resp.Error = function.FuncErrorFromDiags(ctx, diags)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, &merged))
}

type customTransformer struct {
	with_append        bool
	with_union         bool
	with_null_override bool
}

func (t customTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ.Kind() == reflect.Map {
		return func(dst, src reflect.Value) error {
			deepMergeMaps(dst, src, t.with_append, t.with_union, t.with_null_override)
			return nil
		}
	}
	return nil
}

func deepMergeMaps(dst, src reflect.Value, appendSlice bool, uniqueSlice bool, nullOverride bool) reflect.Value {
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

		// BIDIRECTIONAL STICKY UNKNOWN: If either side is unknown, result is unknown
		srcIsUnknown := srcElem.IsValid() && isUnknownSentinel(srcElem)
		dstIsUnknown := dstElem.IsValid() && isUnknownSentinel(dstElem)

		if srcIsUnknown || dstIsUnknown {
			// Prefer src's sentinel if available (more recent type info), otherwise keep dst's
			if srcIsUnknown {
				dst.SetMapIndex(key, srcElem)
			}
			// If only dst is unknown, it stays (already in dst)
			continue
		}

		if srcElem.Kind() == reflect.Map && dstElem.Kind() == reflect.Map {
			// recursive call
			newValue := deepMergeMaps(dstElem, srcElem, appendSlice, uniqueSlice, nullOverride)
			dst.SetMapIndex(key, newValue)
		} else if !srcElem.IsValid() && !nullOverride { // skip override of nil values only if nullOverride is false
			continue
		} else if srcElem.Kind() == reflect.Slice && dstElem.Kind() == reflect.Slice && uniqueSlice { // handle union
			dst.SetMapIndex(key, unionSlices(dstElem, srcElem))
		} else if srcElem.Kind() == reflect.Slice && dstElem.Kind() == reflect.Slice && appendSlice { // handle append
			dst.SetMapIndex(key, reflect.AppendSlice(dstElem, srcElem))
		} else {
			dst.SetMapIndex(key, srcElem)
		}
	}

	return dst
}

// isUnknownSentinel checks if a reflect.Value contains an UnknownSentinel.
func isUnknownSentinel(v reflect.Value) bool {
	if !v.IsValid() || !v.CanInterface() {
		return false
	}
	return helpers.IsUnknownSentinel(v.Interface())
}

func unionSlices(dst, src reflect.Value) reflect.Value {
	result := reflect.MakeSlice(dst.Type(), 0, dst.Len()+src.Len())

	// Add elements from dst (preserving order)
	for i := 0; i < dst.Len(); i++ {
		if !containsElement(result, dst.Index(i)) {
			result = reflect.Append(result, dst.Index(i))
		}
	}

	// Add new elements from src
	for i := 0; i < src.Len(); i++ {
		if !containsElement(result, src.Index(i)) {
			result = reflect.Append(result, src.Index(i))
		}
	}

	return result
}

// containsElement checks if a slice contains a specific element using reflect.DeepEqual.
func containsElement(slice, elem reflect.Value) bool {
	elemInterface := elem.Interface()
	for i := 0; i < slice.Len(); i++ {
		if reflect.DeepEqual(slice.Index(i).Interface(), elemInterface) {
			return true
		}
	}
	return false
}
