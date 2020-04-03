package diag

import "github.com/zclconf/go-cty/cty"

func JoinPath(parent cty.Path, child cty.Path) cty.Path {
	if len(parent) > len(child) {
		tmp := parent
		parent = child
		child = tmp
	}
	if child.HasPrefix(parent) {
		return child
	} else {
		return append(parent, child...)
	}
}
