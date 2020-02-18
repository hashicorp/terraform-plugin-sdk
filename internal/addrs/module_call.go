package addrs

import (
	"fmt"
)

// ModuleCall is the address of a call from the current module to a child
// module.
//
// There is no "Abs" version of ModuleCall because an absolute module path
// is represented by ModuleInstance.
type ModuleCall struct {
	referenceable
	Name string
}

func (c ModuleCall) String() string {
	return "module." + c.Name
}

// ModuleCallInstance is the address of one instance of a module created from
// a module call, which might create multiple instances using "count" or
// "for_each" arguments.
type ModuleCallInstance struct {
	referenceable
	Call ModuleCall
	Key  InstanceKey
}

func (c ModuleCallInstance) String() string {
	if c.Key == NoKey {
		return c.Call.String()
	}
	return fmt.Sprintf("module.%s%s", c.Call.Name, c.Key)
}

// ModuleCallOutput is the address of a particular named output produced by
// an instance of a module call.
type ModuleCallOutput struct {
	referenceable
	Call ModuleCallInstance
	Name string
}

func (co ModuleCallOutput) String() string {
	return fmt.Sprintf("%s.%s", co.Call.String(), co.Name)
}
