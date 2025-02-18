package schema

// TODO: implement IdentityUpgrader struct
type IdentityUpgrader interface{}

type ResourceIdentity struct {
	// Version is the identity schema version.
	Version int64

	// Schema is the structure and type information for the identity.
	// The types allowed in this Schema will be more restricted than
	// previous resource schemas.
	Schema map[string]*Schema

	// New struct, will be similar to (Resource).StateUpgraders
	IdentityUpgraders []IdentityUpgrader
}
