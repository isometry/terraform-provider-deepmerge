// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Originally copied from https://github.com/hashicorp/terraform-provider-kubernetes/blob/main/internal/framework/provider/functions/encode.go

package helpers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func EncodeValue(ctx context.Context, v attr.Value) (any, error) {
	// Avoid nil pointer deref with broken OpenTofu custom function
	// implementation that passes unknown values as zero values.
	if v == nil || v.IsNull() {
		return nil, nil
	}

	// Handle unknown values by returning a sentinel that preserves type info
	if v.IsUnknown() {
		return CreateUnknownSentinel(ctx, v), nil
	}

	switch vv := v.(type) {
	case basetypes.StringValue:
		return vv.ValueString(), nil

	case basetypes.NumberValue:
		f, _ := vv.ValueBigFloat().Float64()
		return f, nil

	case basetypes.BoolValue:
		return vv.ValueBool(), nil

	case basetypes.ObjectValue:
		return EncodeObject(ctx, vv)

	case basetypes.TupleValue:
		return EncodeTuple(ctx, vv)

	case basetypes.MapValue:
		return EncodeMap(ctx, vv)

	case basetypes.ListValue:
		return EncodeList(ctx, vv)

	case basetypes.SetValue:
		return EncodeSet(ctx, vv)

	case basetypes.DynamicValue:
		return EncodeValue(ctx, vv.UnderlyingValue())

	default:
		return nil, fmt.Errorf("tried to encode unsupported type: %T: %v", v, vv)
	}
}

func EncodeSet(ctx context.Context, sv basetypes.SetValue) (es []any, err error) {
	elems := sv.Elements()
	size := len(elems)
	es = make([]any, size)

	for i := range size {
		es[i], err = EncodeValue(ctx, elems[i])
		if err != nil {
			return nil, err
		}
	}

	return es, nil
}

func EncodeList(ctx context.Context, lv basetypes.ListValue) (el []any, err error) {
	elems := lv.Elements()
	size := len(elems)
	el = make([]any, size)

	for i := range size {
		el[i], err = EncodeValue(ctx, elems[i])
		if err != nil {
			return nil, err
		}
	}

	return el, nil
}

func EncodeTuple(ctx context.Context, tv basetypes.TupleValue) (et []any, err error) {
	elems := tv.Elements()
	size := len(elems)
	et = make([]any, size)

	for i := range size {
		et[i], err = EncodeValue(ctx, elems[i])
		if err != nil {
			return nil, err
		}
	}

	return et, nil
}

func EncodeObject(ctx context.Context, ov basetypes.ObjectValue) (eo map[string]any, err error) {
	attrs := ov.Attributes()
	eo = make(map[string]any, len(attrs))

	for k, v := range attrs {
		eo[k], err = EncodeValue(ctx, v)
		if err != nil {
			return nil, err
		}
	}

	return eo, nil
}

func EncodeMap(ctx context.Context, mv basetypes.MapValue) (em map[string]any, err error) {
	elems := mv.Elements()
	em = make(map[string]any, len(elems))

	for k, v := range elems {
		em[k], err = EncodeValue(ctx, v)
		if err != nil {
			return nil, err
		}
	}

	return em, nil
}
