package addrs

// localValue is the address of a local value.
type localValue struct {
	referenceable
	Name string
}

func (v localValue) String() string {
	return "local." + v.Name
}
