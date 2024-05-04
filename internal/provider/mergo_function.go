// Copyright (c) Robin Breathe and contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

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
		MarkdownDescription: "`mergo` takes an arbitrary number of maps or objects, and returns a single map or object that contains a recursively merged set of elements from all arguments.",
		VariadicParameter: function.DynamicParameter{
			Name:                "maps",
			MarkdownDescription: "Maps to merge",
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

	for i, arg := range args {
		if arg.IsNull() {
			resp.Error = function.NewArgumentFuncError(int64(i), "argument must not be null")
			return
		}
		value := arg.UnderlyingValue()
		switch ty := value.Type(ctx); ty {
		case types.StringType:
			switch option := arg.String(); option {
			case `"no_override"`:
				with_override = false
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
		resp.Error = function.NewFuncError("at least one map must be provided")
		return
	}

	if with_override {
		opts = append(opts, mergo.WithOverride)
	}

	merged, diags := helpers.Mergo(ctx, objs, opts...)
	if diags.HasError() {
		resp.Error = function.FuncErrorFromDiags(ctx, diags)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, &merged))
}
