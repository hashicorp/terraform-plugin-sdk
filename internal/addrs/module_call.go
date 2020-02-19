package addrs

import (
	"fmt"
)

// moduleCall is the address of a call from the current module to a child
// module.
//
// There is no "Abs" version of moduleCall because an absolute module path
// is represented by ModuleInstance.
type moduleCall struct {
	referenceable
	Name string
}

func (c moduleCall) String() string {
	return "module." + c.Name
}

// moduleCallInstance is the address of one instance of a module created from
// a module call, which might create multiple instances using "count" or
// "for_each" arguments.
type moduleCallInstance struct {
	referenceable
	Call moduleCall
	Key  instanceKey
}

func (c moduleCallInstance) String() string {
	if c.Key == NoKey {
		return c.Call.String()
	}
	return fmt.Sprintf("module.%s%s", c.Call.Name, c.Key)
}

// moduleCallOutput is the address of a particular named output produced by
// an instance of a module call.
type moduleCallOutput struct {
	referenceable
	Call moduleCallInstance
	Name string
}

func (co moduleCallOutput) String() string {
	return fmt.Sprintf("%s.%s", co.Call.String(), co.Name)
}
