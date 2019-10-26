package terraform

import (
	"fmt"
)

// StateFilterResult is a single result from a filter operation. Filter
// can match multiple things within a state (module, resource, instance, etc.)
// and this unifies that.
type StateFilterResult struct {
	// Module path of the result
	Path []string

	// Address is the address that can be used to reference this exact result.
	Address string

	// Parent, if non-nil, is a parent of this result. For instances, the
	// parent would be a resource. For resources, the parent would be
	// a module. For modules, this is currently nil.
	Parent *StateFilterResult

	// Value is the actual value. This must be type switched on. It can be
	// any data structures that `State` can hold: `ModuleState`,
	// `ResourceState`, `InstanceState`.
	Value interface{}
}

func (r *StateFilterResult) String() string {
	return fmt.Sprintf("%T: %s", r.Value, r.Address)
}

func (r *StateFilterResult) sortedType() int {
	switch r.Value.(type) {
	case *ModuleState:
		return 0
	case *ResourceState:
		return 1
	case *InstanceState:
		return 2
	default:
		return 50
	}
}

// StateFilterResultSlice is a slice of results that implements
// sort.Interface. The sorting goal is what is most appealing to
// human output.
type StateFilterResultSlice []*StateFilterResult

func (s StateFilterResultSlice) Len() int      { return len(s) }
func (s StateFilterResultSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s StateFilterResultSlice) Less(i, j int) bool {
	a, b := s[i], s[j]

	// if these address contain an index, we want to sort by index rather than name
	addrA, errA := ParseResourceAddress(a.Address)
	addrB, errB := ParseResourceAddress(b.Address)
	if errA == nil && errB == nil && addrA.Name == addrB.Name && addrA.Index != addrB.Index {
		return addrA.Index < addrB.Index
	}

	// If the addresses are different it is just lexographic sorting
	if a.Address != b.Address {
		return a.Address < b.Address
	}

	// Addresses are the same, which means it matters on the type
	return a.sortedType() < b.sortedType()
}
