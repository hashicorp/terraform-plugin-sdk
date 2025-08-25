// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package convert

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"log"
)

func PrimitiveTfValue(in cty.Value) tftypes.Value {
	if in.IsNull() || !in.IsKnown() {
		return nullTfValue(ToTfType(in.Type()))
	}

	var val tftypes.Value
	switch in.Type() {
	case cty.String:
		val = tftypes.NewValue(tftypes.String, in.AsString())
	case cty.Bool:
		val = tftypes.NewValue(tftypes.Bool, in.True())
	case cty.Number:
		val = tftypes.NewValue(tftypes.Number, in.AsBigFloat())
	}

	return val
}

func ListTfValue(in cty.Value) tftypes.Value {
	listType := ToTfType(in.Type())

	if in.IsNull() || !in.IsKnown() {
		return nullTfValue(listType)
	}

	vals := make([]tftypes.Value, 0)

	for _, v := range in.AsValueSlice() {
		vals = append(vals, ToTfValue(v))
	}

	return tftypes.NewValue(listType, vals)
}

func MapTfValue(in cty.Value) tftypes.Value {
	mapType := ToTfType(in.Type())

	if in.IsNull() || !in.IsKnown() {
		return nullTfValue(mapType)
	}

	vals := make(map[string]tftypes.Value)

	for k, v := range in.AsValueMap() {
		vals[k] = ToTfValue(v)
	}

	return tftypes.NewValue(mapType, vals)
}

func SetTfValue(in cty.Value) tftypes.Value {
	setType := ToTfType(in.Type())

	if in.IsNull() || !in.IsKnown() {
		return nullTfValue(setType)
	}

	vals := make([]tftypes.Value, 0)

	for _, v := range in.AsValueSlice() {
		vals = append(vals, ToTfValue(v))
	}

	return tftypes.NewValue(setType, vals)
}

func ObjectTfValue(in cty.Value) tftypes.Value {
	objType := ToTfType(in.Type())

	if in.IsNull() || !in.IsKnown() {
		return nullTfValue(objType)
	}

	vals := make(map[string]tftypes.Value)

	for k, v := range in.AsValueMap() {
		vals[k] = ToTfValue(v)
	}

	return tftypes.NewValue(objType, vals)
}

func TupleTfValue(in cty.Value) tftypes.Value {
	tupleType := ToTfType(in.Type())

	if in.IsNull() || !in.IsKnown() {
		return nullTfValue(tupleType)
	}

	vals := make([]tftypes.Value, 0)

	for _, v := range in.AsValueSlice() {
		vals = append(vals, ToTfValue(v))
	}

	return tftypes.NewValue(tupleType, vals)
}

func ToTfValue(in cty.Value) tftypes.Value {
	ty := in.Type()
	switch {
	case ty.IsPrimitiveType():
		return PrimitiveTfValue(in)
	case ty.IsListType():
		return ListTfValue(in)
	case ty.IsObjectType():
		return ObjectTfValue(in)
	case ty.IsMapType():
		return MapTfValue(in)
	case ty.IsSetType():
		return SetTfValue(in)
	case ty.IsTupleType():
		return TupleTfValue(in)
	default:
		log.Panicf("unknown type: %s", ty)
	}

	return nullTfValue(ToTfType(in.Type()))
}

func nullTfValue(ty tftypes.Type) tftypes.Value {
	return tftypes.NewValue(ty, nil)
}
