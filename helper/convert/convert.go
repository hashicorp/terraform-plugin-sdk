package convert

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func PrimitiveTfType(in cty.Value) tftypes.Value {
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

func ListTfType(in cty.Value) tftypes.Value {
	out := make([]tftypes.Value, 0)

	for _, v := range in.AsValueSlice() {
		switch {
		case v.Type().IsPrimitiveType():
			out = append(out, PrimitiveTfType(v))
		case v.Type().IsObjectType():
			out = append(out, ObjectTfType(v))
		}
	}

	elemType, _ := tftypes.TypeFromElements(out)

	return tftypes.NewValue(tftypes.List{ElementType: elemType}, out)
}

func MapTfType(in cty.Value) tftypes.Value {
	out := make(map[string]tftypes.Value)

	mapObj := tftypes.Map{}

	for k, v := range in.AsValueMap() {
		val := tftypes.Value{}

		ty := v.Type()
		switch {
		case ty.IsPrimitiveType():
			val = PrimitiveTfType(v)
			mapObj.ElementType = val.Type()
			out[k] = val
		case ty.IsObjectType():
			out[k] = ObjectTfType(v)
		case ty.IsListType():
			out[k] = ListTfType(v)
		case ty.IsSetType():
			out[k] = SetTfType(v)
		}

		mapObj.ElementType = val.Type()
	}

	return tftypes.NewValue(mapObj, out)
}

func SetTfType(in cty.Value) tftypes.Value {
	out := make([]tftypes.Value, 0)

	for _, v := range in.AsValueSlice() {
		switch {
		case v.Type().IsPrimitiveType():
			out = append(out, PrimitiveTfType(v))
		case v.Type().IsObjectType():
			out = append(out, ObjectTfType(v))
		}
	}

	elemType, _ := tftypes.TypeFromElements(out)

	return tftypes.NewValue(tftypes.Set{ElementType: elemType}, out)
}

func ObjectTfType(in cty.Value) tftypes.Value {
	obj := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}

	vals := make(map[string]tftypes.Value)

	for k, v := range in.AsValueMap() {
		val := tftypes.Value{}

		ty := v.Type()
		switch {
		case ty.IsPrimitiveType():
			val = PrimitiveTfType(v)
		case ty.IsListType():
			val = ListTfType(v)
		case ty.IsObjectType():
			val = ObjectTfType(v)
		case ty.IsMapType():
			val = MapTfType(v)
		case ty.IsSetType():
			val = SetTfType(v)
		}

		obj.AttributeTypes[k] = val.Type()
		vals[k] = val
	}

	return tftypes.NewValue(obj, vals)
}
