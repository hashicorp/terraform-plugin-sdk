package hcltest

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/hcl/v2"
)

// MockBody returns a hcl.Body implementation that works in terms of a
// caller-constructed hcl.BodyContent, thus avoiding the need to parse
// a "real" HCL config file to use as input to a test.
func MockBody(content *hcl.BodyContent) hcl.Body {
	return mockBody{content}
}

type mockBody struct {
	C *hcl.BodyContent
}

func (b mockBody) Content(schema *hcl.BodySchema) (*hcl.BodyContent, hcl.Diagnostics) {
	content, remainI, diags := b.PartialContent(schema)
	remain := remainI.(mockBody)
	for _, attr := range remain.C.Attributes {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Extraneous argument in mock body",
			Detail:   fmt.Sprintf("Mock body has extraneous argument %q.", attr.Name),
			Subject:  &attr.NameRange,
		})
	}
	for _, block := range remain.C.Blocks {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Extraneous block in mock body",
			Detail:   fmt.Sprintf("Mock body has extraneous block of type %q.", block.Type),
			Subject:  &block.DefRange,
		})
	}
	return content, diags
}

func (b mockBody) PartialContent(schema *hcl.BodySchema) (*hcl.BodyContent, hcl.Body, hcl.Diagnostics) {
	ret := &hcl.BodyContent{
		Attributes:       map[string]*hcl.Attribute{},
		Blocks:           []*hcl.Block{},
		MissingItemRange: b.C.MissingItemRange,
	}
	remain := &hcl.BodyContent{
		Attributes:       map[string]*hcl.Attribute{},
		Blocks:           []*hcl.Block{},
		MissingItemRange: b.C.MissingItemRange,
	}
	var diags hcl.Diagnostics

	if len(schema.Attributes) != 0 {
		for _, attrS := range schema.Attributes {
			name := attrS.Name
			attr, ok := b.C.Attributes[name]
			if !ok {
				if attrS.Required {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Missing required argument",
						Detail:   fmt.Sprintf("Mock body doesn't have argument %q", name),
						Subject:  b.C.MissingItemRange.Ptr(),
					})
				}
				continue
			}
			ret.Attributes[name] = attr
		}
	}

	for attrN, attr := range b.C.Attributes {
		if _, ok := ret.Attributes[attrN]; !ok {
			remain.Attributes[attrN] = attr
		}
	}

	wantedBlocks := map[string]hcl.BlockHeaderSchema{}
	for _, blockS := range schema.Blocks {
		wantedBlocks[blockS.Type] = blockS
	}

	for _, block := range b.C.Blocks {
		if blockS, ok := wantedBlocks[block.Type]; ok {
			if len(block.Labels) != len(blockS.LabelNames) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Wrong number of block labels",
					Detail:   fmt.Sprintf("Block of type %q requires %d labels, but got %d", blockS.Type, len(blockS.LabelNames), len(block.Labels)),
					Subject:  b.C.MissingItemRange.Ptr(),
				})
			}

			ret.Blocks = append(ret.Blocks, block)
		} else {
			remain.Blocks = append(remain.Blocks, block)
		}
	}

	return ret, mockBody{remain}, diags
}

func (b mockBody) JustAttributes() (hcl.Attributes, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if len(b.C.Blocks) != 0 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Mock body has blocks",
			Detail:   "Can't use JustAttributes on a mock body with blocks.",
			Subject:  b.C.MissingItemRange.Ptr(),
		})
	}

	return b.C.Attributes, diags
}

func (b mockBody) MissingItemRange() hcl.Range {
	return b.C.MissingItemRange
}

// MockExprLiteral returns a hcl.Expression that evaluates to the given literal
// value.
func MockExprLiteral(val cty.Value) hcl.Expression {
	return mockExprLiteral{val}
}

type mockExprLiteral struct {
	V cty.Value
}

func (e mockExprLiteral) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	return e.V, nil
}

func (e mockExprLiteral) Variables() []hcl.Traversal {
	return nil
}

func (e mockExprLiteral) Range() hcl.Range {
	return hcl.Range{
		Filename: "MockExprLiteral",
	}
}

func (e mockExprLiteral) StartRange() hcl.Range {
	return e.Range()
}

// Implementation for hcl.ExprList
func (e mockExprLiteral) ExprList() []hcl.Expression {
	v := e.V
	ty := v.Type()
	if v.IsKnown() && !v.IsNull() && (ty.IsListType() || ty.IsTupleType()) {
		ret := make([]hcl.Expression, 0, v.LengthInt())
		for it := v.ElementIterator(); it.Next(); {
			_, v := it.Element()
			ret = append(ret, MockExprLiteral(v))
		}
		return ret
	}
	return nil
}
