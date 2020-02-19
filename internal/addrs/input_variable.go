package addrs

// inputVariable is the address of an input variable.
type inputVariable struct {
	referenceable
	Name string
}

func (v inputVariable) String() string {
	return "var." + v.Name
}
