package configschema

import (
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/hcl/v2/hcldec"
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

	return hcldec.ImpliedType(b.DecoderSpec())
}
