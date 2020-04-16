package configschema

import (
	"github.com/hashicorp/go-cty/cty"
)

// ImpliedType returns the cty.Type that would result from decoding a
// configuration block using the receiving block schema.
//
// ImpliedType always returns a result, even if the given schema is
// inconsistent. Code that creates configschema.Block objects should be
// tested using the InternalValidate method to detect any inconsistencies
// that would cause this method to fall back on defaults and assumptions.
func (b *Block) ImpliedType() cty.Type {
	if b == nil {
		return cty.EmptyObject
	}

	atys := make(map[string]cty.Type)

	for name, attrS := range b.Attributes {
		atys[name] = attrS.Type
	}

	for name, blockS := range b.BlockTypes {
		if _, exists := atys[name]; exists {
			// This indicates an invalid schema, since it's not valid to
			// define both an attribute and a block type of the same name.
			// However, we don't raise this here since it's checked by
			// InternalValidate.
			continue
		}

		childType := blockS.Block.ImpliedType()

		switch blockS.Nesting {
		case NestingSingle, NestingGroup:
			atys[name] = childType
		case NestingList:
			// We prefer to use a list where possible, since it makes our
			// implied type more complete, but if there are any
			// dynamically-typed attributes inside we must use a tuple
			// instead, which means our type _constraint_ must be
			// cty.DynamicPseudoType to allow the tuple type to be decided
			// separately for each value.
			if childType.HasDynamicTypes() {
				atys[name] = cty.DynamicPseudoType
			} else {
				atys[name] = cty.List(childType)
			}
		case NestingSet:
			// We forbid dynamically-typed attributes inside NestingSet in
			// InternalValidate, so we will consider that a bug in the caller
			// if we see it here. (There is no set equivalent to tuple and
			// object types, because cty's set implementation depends on
			// knowing the static type in order to properly compute its
			// internal hashes.)
			if childType.HasDynamicTypes() {
				panic("can't use cty.DynamicPseudoType inside a block type with NestingSet")
			}
			atys[name] = cty.Set(childType)
		case NestingMap:
			// We prefer to use a map where possible, since it makes our
			// implied type more complete, but if there are any
			// dynamically-typed attributes inside we must use an object
			// instead, which means our type _constraint_ must be
			// cty.DynamicPseudoType to allow the tuple type to be decided
			// separately for each value.
			if childType.HasDynamicTypes() {
				atys[name] = cty.DynamicPseudoType
			} else {
				atys[name] = cty.Map(childType)
			}
		default:
			// Invalid nesting type is just ignored. It's checked by
			// InternalValidate.
			continue
		}
	}

	return cty.Object(atys)
}
