package configs

import "github.com/hashicorp/hcl2/hcl"

// Connection represents a "connection" block when used within a "resource" block in a module or file.
type Connection struct {
	Config hcl.Body

	DeclRange hcl.Range
}
