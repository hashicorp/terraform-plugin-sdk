package hclsyntax

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/hcl/v2"
)

// Walker is an interface used with Walk.
type Walker interface {
	Enter(node Node) hcl.Diagnostics
	Exit(node Node) hcl.Diagnostics
}

// Walk is a more complex way to traverse the AST starting with a particular
// node, which provides information about the tree structure via separate
// Enter and Exit functions.
func Walk(node Node, w Walker) hcl.Diagnostics {
	diags := w.Enter(node)
	node.walkChildNodes(func(node Node) {
		diags = append(diags, Walk(node, w)...)
	})
	moreDiags := w.Exit(node)
	diags = append(diags, moreDiags...)
	return diags
}
