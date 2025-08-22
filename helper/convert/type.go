package convert

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// An private version of this function to convert cty.Type to tftypes.Type exists in
// internal/plugin/convert/schema.go but this is a simplified and public version
// that maps the types one to one without any error handling, under the assumption
// the incoming data will always be valid
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
		elemTypes := make([]tftypes.Type, 0)

		for _, v := range in.TupleElementTypes() {
			elemTypes = append(elemTypes, ToTfType(v))
		}

		return tftypes.Tuple{ElementTypes: elemTypes}
	}
	return nil
}
