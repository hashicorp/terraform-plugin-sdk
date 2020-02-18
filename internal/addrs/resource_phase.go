package addrs

import "fmt"

// ResourceInstancePhase is a special kind of reference used only internally
// during graph building to represent resource instances that are in a
// non-primary state.
//
// Graph nodes can declare themselves referenceable via an instance phase
// or can declare that they reference an instance phase in order to accomodate
// secondary graph nodes dealing with, for example, destroy actions.
//
// This special reference type cannot be accessed directly by end-users, and
// should never be shown in the UI.
type ResourceInstancePhase struct {
	referenceable
	ResourceInstance ResourceInstance
	Phase            ResourceInstancePhaseType
}

var _ Referenceable = ResourceInstancePhase{}

func (rp ResourceInstancePhase) String() string {
	// We use a different separator here than usual to ensure that we'll
	// never conflict with any non-phased resource instance string. This
	// is intentionally something that would fail parsing with ParseRef,
	// because this special address type should never be exposed in the UI.
	return fmt.Sprintf("%s#%s", rp.ResourceInstance, rp.Phase)
}

// ResourceInstancePhaseType is an enumeration used with ResourceInstancePhase.
type ResourceInstancePhaseType string

func (rpt ResourceInstancePhaseType) String() string {
	return string(rpt)
}

// ResourcePhase is a special kind of reference used only internally
// during graph building to represent resources that are in a
// non-primary state.
//
// Graph nodes can declare themselves referenceable via a resource phase
// or can declare that they reference a resource phase in order to accomodate
// secondary graph nodes dealing with, for example, destroy actions.
//
// Since resources (as opposed to instances) aren't actually phased, this
// address type is used only as an approximation during initial construction
// of the resource-oriented plan graph, under the assumption that resource
// instances with ResourceInstancePhase addresses will be created in dynamic
// subgraphs during the graph walk.
//
// This special reference type cannot be accessed directly by end-users, and
// should never be shown in the UI.
type ResourcePhase struct {
	referenceable
	Resource Resource
	Phase    ResourceInstancePhaseType
}

var _ Referenceable = ResourcePhase{}

func (rp ResourcePhase) String() string {
	// We use a different separator here than usual to ensure that we'll
	// never conflict with any non-phased resource instance string. This
	// is intentionally something that would fail parsing with ParseRef,
	// because this special address type should never be exposed in the UI.
	return fmt.Sprintf("%s#%s", rp.Resource, rp.Phase)
}
