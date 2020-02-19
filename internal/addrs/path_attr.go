package addrs

// pathAttr is the address of an attribute of the "path" object in
// the interpolation scope, like "path.module".
type pathAttr struct {
	referenceable
	Name string
}

func (pa pathAttr) String() string {
	return "path." + pa.Name
}
