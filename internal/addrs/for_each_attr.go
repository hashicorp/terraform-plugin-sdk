package addrs

// forEachAttr is the address of an attribute referencing the current "for_each" object in
// the interpolation scope, addressed using the "each" keyword, ex. "each.key" and "each.value"
type forEachAttr struct {
	referenceable
	Name string
}

func (f forEachAttr) String() string {
	return "each." + f.Name
}
