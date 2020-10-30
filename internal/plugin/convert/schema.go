package convert

import (
	"fmt"
	"log"
	"reflect"
	"sort"

	"github.com/hashicorp/go-cty/cty"
	proto "github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/configschema"
)

func tftypeFromCtyType(in cty.Type) (tftypes.Type, error) {
	switch {
	case in.Equals(cty.String):
		return tftypes.String, nil
	case in.Equals(cty.Number):
		return tftypes.Number, nil
	case in.Equals(cty.Bool):
		return tftypes.Bool, nil
	case in.Equals(cty.DynamicPseudoType):
		return tftypes.DynamicPseudoType, nil
	case in.IsSetType():
		elemType, err := tftypeFromCtyType(in.ElementType())
		if err != nil {
			return nil, err
		}
		return tftypes.Set{
			ElementType: elemType,
		}, nil
	case in.IsListType():
		elemType, err := tftypeFromCtyType(in.ElementType())
		if err != nil {
			return nil, err
		}
		return tftypes.List{
			ElementType: elemType,
		}, nil
	case in.IsTupleType():
		elemTypes := make([]tftypes.Type, 0, in.Length())
		for _, typ := range in.TupleElementTypes() {
			elemType, err := tftypeFromCtyType(typ)
			if err != nil {
				return nil, err
			}
			elemTypes = append(elemTypes, elemType)
		}
		return tftypes.Tuple{
			ElementTypes: elemTypes,
		}, nil
	case in.IsMapType():
		elemType, err := tftypeFromCtyType(in.ElementType())
		if err != nil {
			return nil, err
		}
		return tftypes.Map{
			AttributeType: elemType,
		}, nil
	case in.IsObjectType():
		attrTypes := make(map[string]tftypes.Type)
		for key, typ := range in.AttributeTypes() {
			attrType, err := tftypeFromCtyType(typ)
			if err != nil {
				return nil, err
			}
			attrTypes[key] = attrType
		}
		return tftypes.Object{
			AttributeTypes: attrTypes,
		}, nil
	}
	return nil, fmt.Errorf("unknown cty type %s", in.GoString())
}

func ctyTypeFromTFType(in tftypes.Type) (cty.Type, error) {
	switch {
	case in.Is(tftypes.String):
		return cty.String, nil
	case in.Is(tftypes.Bool):
		return cty.Bool, nil
	case in.Is(tftypes.Number):
		return cty.Number, nil
	case in.Is(tftypes.DynamicPseudoType):
		return cty.DynamicPseudoType, nil
	case in.Is(tftypes.List{}):
		elemType, err := ctyTypeFromTFType(in.(tftypes.List).ElementType)
		if err != nil {
			return cty.Type{}, err
		}
		return cty.List(elemType), nil
	case in.Is(tftypes.Set{}):
		elemType, err := ctyTypeFromTFType(in.(tftypes.Set).ElementType)
		if err != nil {
			return cty.Type{}, err
		}
		return cty.Set(elemType), nil
	case in.Is(tftypes.Map{}):
		elemType, err := ctyTypeFromTFType(in.(tftypes.Map).AttributeType)
		if err != nil {
			return cty.Type{}, err
		}
		return cty.Map(elemType), nil
	case in.Is(tftypes.Tuple{}):
		elemTypes := make([]cty.Type, 0, len(in.(tftypes.Tuple).ElementTypes))
		for _, typ := range in.(tftypes.Tuple).ElementTypes {
			elemType, err := ctyTypeFromTFType(typ)
			if err != nil {
				return cty.Type{}, err
			}
			elemTypes = append(elemTypes, elemType)
		}
		return cty.Tuple(elemTypes), nil
	case in.Is(tftypes.Object{}):
		attrTypes := make(map[string]cty.Type, len(in.(tftypes.Object).AttributeTypes))
		for k, v := range in.(tftypes.Object).AttributeTypes {
			attrType, err := ctyTypeFromTFType(v)
			if err != nil {
				return cty.Type{}, err
			}
			attrTypes[k] = attrType
		}
		return cty.Object(attrTypes), nil
	}
	return cty.Type{}, fmt.Errorf("unknown tftypes.Type %s", in)
}

// ConfigSchemaToProto takes a *configschema.Block and converts it to a
// proto.SchemaBlock for a grpc response.
func ConfigSchemaToProto(b *configschema.Block) *proto.SchemaBlock {
	block := &proto.SchemaBlock{
		Description:     b.Description,
		DescriptionKind: protoStringKind(b.DescriptionKind),
		Deprecated:      b.Deprecated,
	}

	for _, name := range sortedKeys(b.Attributes) {
		a := b.Attributes[name]

		attr := &proto.SchemaAttribute{
			Name:            name,
			Description:     a.Description,
			DescriptionKind: protoStringKind(a.DescriptionKind),
			Optional:        a.Optional,
			Computed:        a.Computed,
			Required:        a.Required,
			Sensitive:       a.Sensitive,
			Deprecated:      a.Deprecated,
		}

		var err error
		attr.Type, err = tftypeFromCtyType(a.Type)
		if err != nil {
			panic(err)
		}

		block.Attributes = append(block.Attributes, attr)
	}

	for _, name := range sortedKeys(b.BlockTypes) {
		b := b.BlockTypes[name]
		block.BlockTypes = append(block.BlockTypes, protoSchemaNestedBlock(name, b))
	}

	return block
}

func protoStringKind(k configschema.StringKind) proto.StringKind {
	switch k {
	default:
		log.Printf("[TRACE] unexpected configschema.StringKind: %d", k)
		return proto.StringKindPlain
	case configschema.StringPlain:
		return proto.StringKindPlain
	case configschema.StringMarkdown:
		return proto.StringKindMarkdown
	}
}

func protoSchemaNestedBlock(name string, b *configschema.NestedBlock) *proto.SchemaNestedBlock {
	var nesting proto.SchemaNestedBlockNestingMode
	switch b.Nesting {
	case configschema.NestingSingle:
		nesting = proto.SchemaNestedBlockNestingModeSingle
	case configschema.NestingGroup:
		nesting = proto.SchemaNestedBlockNestingModeGroup
	case configschema.NestingList:
		nesting = proto.SchemaNestedBlockNestingModeList
	case configschema.NestingSet:
		nesting = proto.SchemaNestedBlockNestingModeSet
	case configschema.NestingMap:
		nesting = proto.SchemaNestedBlockNestingModeMap
	default:
		nesting = proto.SchemaNestedBlockNestingModeInvalid
	}
	return &proto.SchemaNestedBlock{
		TypeName: name,
		Block:    ConfigSchemaToProto(&b.Block),
		Nesting:  nesting,
		MinItems: int64(b.MinItems),
		MaxItems: int64(b.MaxItems),
	}
}

// ProtoToConfigSchema takes the GetSchema_Block from a grpc response and converts it
// to a terraform *configschema.Block.
func ProtoToConfigSchema(b *proto.SchemaBlock) *configschema.Block {
	block := &configschema.Block{
		Attributes: make(map[string]*configschema.Attribute),
		BlockTypes: make(map[string]*configschema.NestedBlock),

		Description:     b.Description,
		DescriptionKind: schemaStringKind(b.DescriptionKind),
		Deprecated:      b.Deprecated,
	}

	for _, a := range b.Attributes {
		attr := &configschema.Attribute{
			Description:     a.Description,
			DescriptionKind: schemaStringKind(a.DescriptionKind),
			Required:        a.Required,
			Optional:        a.Optional,
			Computed:        a.Computed,
			Sensitive:       a.Sensitive,
			Deprecated:      a.Deprecated,
		}

		var err error
		attr.Type, err = ctyTypeFromTFType(a.Type)
		if err != nil {
			panic(err)
		}

		block.Attributes[a.Name] = attr
	}

	for _, b := range b.BlockTypes {
		block.BlockTypes[b.TypeName] = schemaNestedBlock(b)
	}

	return block
}

func schemaStringKind(k proto.StringKind) configschema.StringKind {
	switch k {
	default:
		log.Printf("[TRACE] unexpected proto.StringKind: %d", k)
		return configschema.StringPlain
	case proto.StringKindPlain:
		return configschema.StringPlain
	case proto.StringKindMarkdown:
		return configschema.StringMarkdown
	}
}

func schemaNestedBlock(b *proto.SchemaNestedBlock) *configschema.NestedBlock {
	var nesting configschema.NestingMode
	switch b.Nesting {
	case proto.SchemaNestedBlockNestingModeSingle:
		nesting = configschema.NestingSingle
	case proto.SchemaNestedBlockNestingModeGroup:
		nesting = configschema.NestingGroup
	case proto.SchemaNestedBlockNestingModeList:
		nesting = configschema.NestingList
	case proto.SchemaNestedBlockNestingModeMap:
		nesting = configschema.NestingMap
	case proto.SchemaNestedBlockNestingModeSet:
		nesting = configschema.NestingSet
	default:
		// In all other cases we'll leave it as the zero value (invalid) and
		// let the caller validate it and deal with this.
	}

	nb := &configschema.NestedBlock{
		Nesting:  nesting,
		MinItems: int(b.MinItems),
		MaxItems: int(b.MaxItems),
	}

	nested := ProtoToConfigSchema(b.Block)
	nb.Block = *nested
	return nb
}

// sortedKeys returns the lexically sorted keys from the given map. This is
// used to make schema conversions are deterministic. This panics if map keys
// are not a string.
func sortedKeys(m interface{}) []string {
	v := reflect.ValueOf(m)
	keys := make([]string, v.Len())

	mapKeys := v.MapKeys()
	for i, k := range mapKeys {
		keys[i] = k.Interface().(string)
	}

	sort.Strings(keys)
	return keys
}
