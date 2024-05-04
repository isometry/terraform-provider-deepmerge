// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Originally copied from https://github.com/hashicorp/terraform-provider-kubernetes/blob/main/internal/framework/provider/functions/encode.go

package helpers

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func EncodeValue(v attr.Value) (any, error) {
	if v.IsNull() {
		return nil, nil
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
		return EncodeObject(vv)
	case basetypes.TupleValue:
		return EncodeTuple(vv)
	case basetypes.MapValue:
		return EncodeMap(vv)
	case basetypes.ListValue:
		return EncodeList(vv)
	case basetypes.SetValue:
		return EncodeSet(vv)
	case basetypes.DynamicValue:
		return EncodeValue(vv.UnderlyingValue())
	default:
		return nil, fmt.Errorf("tried to encode unsupported type: %T: %v", v, vv)
	}
}

func EncodeSet(sv basetypes.SetValue) ([]any, error) {
	elems := sv.Elements()
	size := len(elems)
	l := make([]any, size)
	for i := 0; i < size; i++ {
		var err error
		l[i], err = EncodeValue(elems[i])
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}

func EncodeList(lv basetypes.ListValue) ([]any, error) {
	elems := lv.Elements()
	size := len(elems)
	l := make([]any, size)
	for i := 0; i < size; i++ {
		var err error
		l[i], err = EncodeValue(elems[i])
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}

func EncodeTuple(tv basetypes.TupleValue) ([]any, error) {
	elems := tv.Elements()
	size := len(elems)
	l := make([]any, size)
	for i := 0; i < size; i++ {
		var err error
		l[i], err = EncodeValue(elems[i])
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}

func EncodeObject(ov basetypes.ObjectValue) (map[string]any, error) {
	attrs := ov.Attributes()
	m := make(map[string]any, len(attrs))
	for k, v := range attrs {
		var err error
		m[k], err = EncodeValue(v)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func EncodeMap(mv basetypes.MapValue) (map[string]any, error) {
	elems := mv.Elements()
	m := make(map[string]any, len(elems))
	for k, v := range elems {
		var err error
		m[k], err = EncodeValue(v)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}
