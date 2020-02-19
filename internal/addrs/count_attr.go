package addrs

// countAttr is the address of an attribute of the "count" object in
// the interpolation scope, like "count.index".
type countAttr struct {
	referenceable
	Name string
}

func (ca countAttr) String() string {
	return "count." + ca.Name
}
