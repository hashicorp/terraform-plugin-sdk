package addrs

// terraformAttr is the address of an attribute of the "terraform" object in
// the interpolation scope, like "terraform.workspace".
type terraformAttr struct {
	referenceable
	Name string
}

func (ta terraformAttr) String() string {
	return "terraform." + ta.Name
}
