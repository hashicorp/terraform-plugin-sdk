package terraform

//go:generate go run golang.org/x/tools/cmd/stringer -type=resourceMode -output=resource_mode_string.go resource_mode.go

// resourceMode is deprecated, use addrs.ResourceMode instead.
// It has been preserved for backwards compatibility.
type resourceMode int

const (
	managedResourceMode resourceMode = iota
	dataResourceMode
)
