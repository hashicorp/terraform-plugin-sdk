package hcldec

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/hcl/v2"
)

// Decode interprets the given body using the given specification and returns
// the resulting value. If the given body is not valid per the spec, error
// diagnostics are returned and the returned value is likely to be incomplete.
//
// The ctx argument may be nil, in which case any references to variables or
// functions will produce error diagnostics.
func Decode(body hcl.Body, spec Spec, ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	val, _, diags := decode(body, nil, ctx, spec, false)
	return val, diags
}

// ImpliedType returns the value type that should result from decoding the
// given spec.
func ImpliedType(spec Spec) cty.Type {
	return impliedType(spec)
}
