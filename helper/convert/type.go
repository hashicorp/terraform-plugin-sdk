package convert

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func ToTfType(in cty.Type) tftypes.Type {
	switch {
	case in.IsPrimitiveType():
		if in == cty.String {
			return tftypes.String
		}
		if in == cty.Bool {
			return tftypes.Bool
		}
		if in == cty.Number {
			return tftypes.Number
		}
	case in.IsListType():
		elemType := ToTfType(in.ElementType())
		return tftypes.List{ElementType: elemType}
	case in.IsSetType():
		elemType := ToTfType(in.ElementType())
		return tftypes.Set{ElementType: elemType}
	case in.IsMapType():
		elemType := ToTfType(in.ElementType())
		return tftypes.Map{ElementType: elemType}
	case in.IsObjectType():
		attrTypes := map[string]tftypes.Type{}

		for k, v := range in.AttributeTypes() {
			attrTypes[k] = ToTfType(v)
		}
		return tftypes.Object{AttributeTypes: attrTypes}
	case in.IsTupleType():
		// TODO
	}
	return nil
}
