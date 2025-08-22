package convert

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func PrimitiveTfValue(in cty.Value) tftypes.Value {
	if in.IsNull() {
		return emptyTfValue(ToTfType(in.Type()))
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

	if in.IsNull() || in.LengthInt() == 0 {
		return emptyTfValue(listType)
	}

	vals := make([]tftypes.Value, 0)

	for _, v := range in.AsValueSlice() {
		vals = append(vals, ToTfValue(v))
	}

	return tftypes.NewValue(listType, vals)
}

func MapTfValue(in cty.Value) tftypes.Value {
	mapType := ToTfType(in.Type())

	if in.IsNull() || in.LengthInt() == 0 {
		return emptyTfValue(mapType)
	}

	vals := make(map[string]tftypes.Value)

	for k, v := range in.AsValueMap() {
		vals[k] = ToTfValue(v)
	}

	return tftypes.NewValue(mapType, vals)
}

func SetTfValue(in cty.Value) tftypes.Value {
	setType := ToTfType(in.Type())

	if in.IsNull() || in.LengthInt() == 0 {
		return emptyTfValue(setType)
	}

	vals := make([]tftypes.Value, 0)

	for _, v := range in.AsValueSlice() {
		vals = append(vals, ToTfValue(v))
	}

	return tftypes.NewValue(setType, vals)
}

func ObjectTfValue(in cty.Value) tftypes.Value {
	objType := ToTfType(in.Type())

	if in.IsNull() || in.LengthInt() == 0 {
		return emptyTfValue(objType)
	}

	vals := make(map[string]tftypes.Value)

	for k, v := range in.AsValueMap() {
		vals[k] = ToTfValue(v)
	}

	return tftypes.NewValue(objType, vals)
}

func TupleTfValue(in cty.Value) tftypes.Value {
	tupleType := ToTfType(in.Type())

	if in.IsNull() || in.LengthInt() == 0 {
		return emptyTfValue(tupleType)
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
	}

	return emptyTfValue(ToTfType(in.Type()))
}

func emptyTfValue(ty tftypes.Type) tftypes.Value {
	return tftypes.NewValue(ty, nil)
}
