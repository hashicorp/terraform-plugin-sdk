package terraform

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/terraform-plugin-sdk/internal/addrs"
	"github.com/hashicorp/terraform-plugin-sdk/internal/configs"
	"github.com/hashicorp/terraform-plugin-sdk/internal/configs/configschema"
	"github.com/hashicorp/terraform-plugin-sdk/internal/configs/hcl2shim"
	"github.com/hashicorp/terraform-plugin-sdk/internal/plans"
	"github.com/hashicorp/terraform-plugin-sdk/internal/tfdiags"
	"github.com/mitchellh/copystructure"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

const (
	// StateVersion is the current version for our state file
	StateVersion = 3
)

// rootModulePath is the path of the root module
var rootModulePath = []string{"root"}

// normalizeModulePath transforms a legacy module path (which may or may not
// have a redundant "root" label at the start of it) into an
// addrs.ModuleInstance representing the same module.
//
// For legacy reasons, different parts of Terraform disagree about whether the
// root module has the path []string{} or []string{"root"}, and so this
// function accepts both and trims off the "root". An implication of this is
// that it's not possible to actually have a module call in the root module
// that is itself named "root", since that would be ambiguous.
//
// normalizeModulePath takes a raw module path and returns a path that
// has the rootModulePath prepended to it. If I could go back in time I
// would've never had a rootModulePath (empty path would be root). We can
// still fix this but thats a big refactor that my branch doesn't make sense
// for. Instead, this function normalizes paths.
func normalizeModulePath(p []string) addrs.ModuleInstance {
	// FIXME: Remove this once everyone is using addrs.ModuleInstance.

	if len(p) > 0 && p[0] == "root" {
		p = p[1:]
	}

	ret := make(addrs.ModuleInstance, len(p))
	for i, name := range p {
		// For now we don't actually support modules with multiple instances
		// identified by keys, so we just treat every path element as a
		// step with no key.
		ret[i] = addrs.ModuleInstanceStep{
			Name: name,
		}
	}
	return ret
}

type StateAgeComparison int

const (
	StateAgeEqual         StateAgeComparison = 0
	StateAgeReceiverNewer StateAgeComparison = 1
	StateAgeReceiverOlder StateAgeComparison = -1
)

// BackendState stores the configuration to connect to a remote backend.
type BackendState struct {
	Type      string          `json:"type"`   // Backend type
	ConfigRaw json.RawMessage `json:"config"` // Backend raw config
	Hash      uint64          `json:"hash"`   // Hash of portion of configuration from config files
}

// Empty returns true if BackendState has no state.
func (s *BackendState) Empty() bool {
	return s == nil || s.Type == ""
}

// Config decodes the type-specific configuration object using the provided
// schema and returns the result as a cty.Value.
//
// An error is returned if the stored configuration does not conform to the
// given schema.
func (s *BackendState) Config(schema *configschema.Block) (cty.Value, error) {
	ty := schema.ImpliedType()
	if s == nil {
		return cty.NullVal(ty), nil
	}
	return ctyjson.Unmarshal(s.ConfigRaw, ty)
}

// SetConfig replaces (in-place) the type-specific configuration object using
// the provided value and associated schema.
//
// An error is returned if the given value does not conform to the implied
// type of the schema.
func (s *BackendState) SetConfig(val cty.Value, schema *configschema.Block) error {
	ty := schema.ImpliedType()
	buf, err := ctyjson.Marshal(val, ty)
	if err != nil {
		return err
	}
	s.ConfigRaw = buf
	return nil
}

// ForPlan produces an alternative representation of the reciever that is
// suitable for storing in a plan. The current workspace must additionally
// be provided, to be stored alongside the backend configuration.
//
// The backend configuration schema is required in order to properly
// encode the backend-specific configuration settings.
func (s *BackendState) ForPlan(schema *configschema.Block, workspaceName string) (*plans.Backend, error) {
	if s == nil {
		return nil, nil
	}

	configVal, err := s.Config(schema)
	if err != nil {
		return nil, errwrap.Wrapf("failed to decode backend config: {{err}}", err)
	}
	return plans.NewBackend(s.Type, configVal, schema, workspaceName)
}

// RemoteState is used to track the information about a remote
// state store that we push/pull state to.
type RemoteState struct {
	// Type controls the client we use for the remote state
	Type string `json:"type"`

	// Config is used to store arbitrary configuration that
	// is type specific
	Config map[string]string `json:"config"`

	mu sync.Mutex
}

func (s *RemoteState) Lock()   { s.mu.Lock() }
func (s *RemoteState) Unlock() { s.mu.Unlock() }

func (r *RemoteState) Empty() bool {
	if r == nil {
		return true
	}
	r.Lock()
	defer r.Unlock()

	return r.Type == ""
}

func (r *RemoteState) Equals(other *RemoteState) bool {
	r.Lock()
	defer r.Unlock()

	if r.Type != other.Type {
		return false
	}
	if len(r.Config) != len(other.Config) {
		return false
	}
	for k, v := range r.Config {
		if other.Config[k] != v {
			return false
		}
	}
	return true
}

// OutputState is used to track the state relevant to a single output.
type OutputState struct {
	// Sensitive describes whether the output is considered sensitive,
	// which may lead to masking the value on screen in some cases.
	Sensitive bool `json:"sensitive"`
	// Type describes the structure of Value. Valid values are "string",
	// "map" and "list"
	Type string `json:"type"`
	// Value contains the value of the output, in the structure described
	// by the Type field.
	Value interface{} `json:"value"`

	mu sync.Mutex
}

func (s *OutputState) Lock()   { s.mu.Lock() }
func (s *OutputState) Unlock() { s.mu.Unlock() }

func (s *OutputState) String() string {
	return fmt.Sprintf("%#v", s.Value)
}

// Equal compares two OutputState structures for equality. nil values are
// considered equal.
func (s *OutputState) Equal(other *OutputState) bool {
	if s == nil && other == nil {
		return true
	}

	if s == nil || other == nil {
		return false
	}
	s.Lock()
	defer s.Unlock()

	if s.Type != other.Type {
		return false
	}

	if s.Sensitive != other.Sensitive {
		return false
	}

	if !reflect.DeepEqual(s.Value, other.Value) {
		return false
	}

	return true
}

// ModuleState is used to track all the state relevant to a single
// module. Previous to Terraform 0.3, all state belonged to the "root"
// module.
type ModuleState struct {
	// Path is the import path from the root module. Modules imports are
	// always disjoint, so the path represents amodule tree
	Path []string `json:"path"`

	// Locals are kept only transiently in-memory, because we can always
	// re-compute them.
	Locals map[string]interface{} `json:"-"`

	// Outputs declared by the module and maintained for each module
	// even though only the root module technically needs to be kept.
	// This allows operators to inspect values at the boundaries.
	Outputs map[string]*OutputState `json:"outputs"`

	// Resources is a mapping of the logically named resource to
	// the state of the resource. Each resource may actually have
	// N instances underneath, although a user only needs to think
	// about the 1:1 case.
	Resources map[string]*ResourceState `json:"resources"`

	// Dependencies are a list of things that this module relies on
	// existing to remain intact. For example: an module may depend
	// on a VPC ID given by an aws_vpc resource.
	//
	// Terraform uses this information to build valid destruction
	// orders and to warn the user if they're destroying a module that
	// another resource depends on.
	//
	// Things can be put into this list that may not be managed by
	// Terraform. If Terraform doesn't find a matching ID in the
	// overall state, then it assumes it isn't managed and doesn't
	// worry about it.
	Dependencies []string `json:"depends_on"`

	mu sync.Mutex
}

func (s *ModuleState) Lock()   { s.mu.Lock() }
func (s *ModuleState) Unlock() { s.mu.Unlock() }

// Equal tests whether one module state is equal to another.
func (m *ModuleState) Equal(other *ModuleState) bool {
	m.Lock()
	defer m.Unlock()

	// Paths must be equal
	if !reflect.DeepEqual(m.Path, other.Path) {
		return false
	}

	// Outputs must be equal
	if len(m.Outputs) != len(other.Outputs) {
		return false
	}
	for k, v := range m.Outputs {
		if !other.Outputs[k].Equal(v) {
			return false
		}
	}

	// Dependencies must be equal. This sorts these in place but
	// this shouldn't cause any problems.
	sort.Strings(m.Dependencies)
	sort.Strings(other.Dependencies)
	if len(m.Dependencies) != len(other.Dependencies) {
		return false
	}
	for i, d := range m.Dependencies {
		if other.Dependencies[i] != d {
			return false
		}
	}

	// Resources must be equal
	if len(m.Resources) != len(other.Resources) {
		return false
	}
	for k, r := range m.Resources {
		otherR, ok := other.Resources[k]
		if !ok {
			return false
		}

		if !r.Equal(otherR) {
			return false
		}
	}

	return true
}

// IsRoot says whether or not this module diff is for the root module.
func (m *ModuleState) IsRoot() bool {
	m.Lock()
	defer m.Unlock()
	return reflect.DeepEqual(m.Path, rootModulePath)
}

// IsDescendent returns true if other is a descendent of this module.
func (m *ModuleState) IsDescendent(other *ModuleState) bool {
	m.Lock()
	defer m.Unlock()

	i := len(m.Path)
	return len(other.Path) > i && reflect.DeepEqual(other.Path[:i], m.Path)
}

// Orphans returns a list of keys of resources that are in the State
// but aren't present in the configuration itself. Hence, these keys
// represent the state of resources that are orphans.
func (m *ModuleState) Orphans(c *configs.Module) []addrs.ResourceInstance {
	m.Lock()
	defer m.Unlock()

	inConfig := make(map[string]struct{})
	if c != nil {
		for _, r := range c.ManagedResources {
			inConfig[r.Addr().String()] = struct{}{}
		}
		for _, r := range c.DataResources {
			inConfig[r.Addr().String()] = struct{}{}
		}
	}

	var result []addrs.ResourceInstance
	for k := range m.Resources {
		// Since we've not yet updated state to use our new address format,
		// we need to do some shimming here.
		legacyAddr, err := parseResourceAddressInternal(k)
		if err != nil {
			// Suggests that the user tampered with the state, since we always
			// generate valid internal addresses.
			log.Printf("ModuleState has invalid resource key %q. Ignoring.", k)
			continue
		}

		addr := legacyAddr.AbsResourceInstanceAddr().Resource
		compareKey := addr.Resource.String() // compare by resource address, ignoring instance key
		if _, exists := inConfig[compareKey]; !exists {
			result = append(result, addr)
		}
	}
	return result
}

// RemovedOutputs returns a list of outputs that are in the State but aren't
// present in the configuration itself.
func (s *ModuleState) RemovedOutputs(outputs map[string]*configs.Output) []addrs.OutputValue {
	if outputs == nil {
		// If we got no output map at all then we'll just treat our set of
		// configured outputs as empty, since that suggests that they've all
		// been removed by removing their containing module.
		outputs = make(map[string]*configs.Output)
	}

	s.Lock()
	defer s.Unlock()

	var ret []addrs.OutputValue
	for n := range s.Outputs {
		if _, declared := outputs[n]; !declared {
			ret = append(ret, addrs.OutputValue{
				Name: n,
			})
		}
	}

	return ret
}

// View returns a view with the given resource prefix.
func (m *ModuleState) View(id string) *ModuleState {
	if m == nil {
		return m
	}

	r := m.deepcopy()
	for k, _ := range r.Resources {
		if id == k || strings.HasPrefix(k, id+".") {
			continue
		}

		delete(r.Resources, k)
	}

	return r
}

func (m *ModuleState) deepcopy() *ModuleState {
	if m == nil {
		return nil
	}

	stateCopy, err := copystructure.Config{Lock: true}.Copy(m)
	if err != nil {
		panic(err)
	}

	return stateCopy.(*ModuleState)
}

func (m *ModuleState) String() string {
	m.Lock()
	defer m.Unlock()

	var buf bytes.Buffer

	if len(m.Resources) == 0 {
		buf.WriteString("<no state>")
	}

	names := make([]string, 0, len(m.Resources))
	for name, _ := range m.Resources {
		names = append(names, name)
	}

	sort.Sort(resourceNameSort(names))

	for _, k := range names {
		rs := m.Resources[k]
		var id string
		if rs.Primary != nil {
			id = rs.Primary.ID
		}
		if id == "" {
			id = "<not created>"
		}

		taintStr := ""
		if rs.Primary.Tainted {
			taintStr = " (tainted)"
		}

		deposedStr := ""
		if len(rs.Deposed) > 0 {
			deposedStr = fmt.Sprintf(" (%d deposed)", len(rs.Deposed))
		}

		buf.WriteString(fmt.Sprintf("%s:%s%s\n", k, taintStr, deposedStr))
		buf.WriteString(fmt.Sprintf("  ID = %s\n", id))
		if rs.Provider != "" {
			buf.WriteString(fmt.Sprintf("  provider = %s\n", rs.Provider))
		}

		var attributes map[string]string
		if rs.Primary != nil {
			attributes = rs.Primary.Attributes
		}
		attrKeys := make([]string, 0, len(attributes))
		for ak, _ := range attributes {
			if ak == "id" {
				continue
			}

			attrKeys = append(attrKeys, ak)
		}

		sort.Strings(attrKeys)

		for _, ak := range attrKeys {
			av := attributes[ak]
			buf.WriteString(fmt.Sprintf("  %s = %s\n", ak, av))
		}

		for idx, t := range rs.Deposed {
			taintStr := ""
			if t.Tainted {
				taintStr = " (tainted)"
			}
			buf.WriteString(fmt.Sprintf("  Deposed ID %d = %s%s\n", idx+1, t.ID, taintStr))
		}

		if len(rs.Dependencies) > 0 {
			buf.WriteString(fmt.Sprintf("\n  Dependencies:\n"))
			for _, dep := range rs.Dependencies {
				buf.WriteString(fmt.Sprintf("    %s\n", dep))
			}
		}
	}

	if len(m.Outputs) > 0 {
		buf.WriteString("\nOutputs:\n\n")

		ks := make([]string, 0, len(m.Outputs))
		for k, _ := range m.Outputs {
			ks = append(ks, k)
		}

		sort.Strings(ks)

		for _, k := range ks {
			v := m.Outputs[k]
			switch vTyped := v.Value.(type) {
			case string:
				buf.WriteString(fmt.Sprintf("%s = %s\n", k, vTyped))
			case []interface{}:
				buf.WriteString(fmt.Sprintf("%s = %s\n", k, vTyped))
			case map[string]interface{}:
				var mapKeys []string
				for key, _ := range vTyped {
					mapKeys = append(mapKeys, key)
				}
				sort.Strings(mapKeys)

				var mapBuf bytes.Buffer
				mapBuf.WriteString("{")
				for _, key := range mapKeys {
					mapBuf.WriteString(fmt.Sprintf("%s:%s ", key, vTyped[key]))
				}
				mapBuf.WriteString("}")

				buf.WriteString(fmt.Sprintf("%s = %s\n", k, mapBuf.String()))
			}
		}
	}

	return buf.String()
}

func (m *ModuleState) Empty() bool {
	return len(m.Locals) == 0 && len(m.Outputs) == 0 && len(m.Resources) == 0
}

// ResourceStateKey is a structured representation of the key used for the
// ModuleState.Resources mapping
type ResourceStateKey struct {
	Name  string
	Type  string
	Mode  ResourceMode
	Index int
}

// Equal determines whether two ResourceStateKeys are the same
func (rsk *ResourceStateKey) Equal(other *ResourceStateKey) bool {
	if rsk == nil || other == nil {
		return false
	}
	if rsk.Mode != other.Mode {
		return false
	}
	if rsk.Type != other.Type {
		return false
	}
	if rsk.Name != other.Name {
		return false
	}
	if rsk.Index != other.Index {
		return false
	}
	return true
}

func (rsk *ResourceStateKey) String() string {
	if rsk == nil {
		return ""
	}
	var prefix string
	switch rsk.Mode {
	case ManagedResourceMode:
		prefix = ""
	case DataResourceMode:
		prefix = "data."
	default:
		panic(fmt.Errorf("unknown resource mode %s", rsk.Mode))
	}
	if rsk.Index == -1 {
		return fmt.Sprintf("%s%s.%s", prefix, rsk.Type, rsk.Name)
	}
	return fmt.Sprintf("%s%s.%s.%d", prefix, rsk.Type, rsk.Name, rsk.Index)
}

// ParseResourceStateKey accepts a key in the format used by
// ModuleState.Resources and returns a resource name and resource index. In the
// state, a resource has the format "type.name.index" or "type.name". In the
// latter case, the index is returned as -1.
func ParseResourceStateKey(k string) (*ResourceStateKey, error) {
	parts := strings.Split(k, ".")
	mode := ManagedResourceMode
	if len(parts) > 0 && parts[0] == "data" {
		mode = DataResourceMode
		// Don't need the constant "data" prefix for parsing
		// now that we've figured out the mode.
		parts = parts[1:]
	}
	if len(parts) < 2 || len(parts) > 3 {
		return nil, fmt.Errorf("Malformed resource state key: %s", k)
	}
	rsk := &ResourceStateKey{
		Mode:  mode,
		Type:  parts[0],
		Name:  parts[1],
		Index: -1,
	}
	if len(parts) == 3 {
		index, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("Malformed resource state key index: %s", k)
		}
		rsk.Index = index
	}
	return rsk, nil
}

// ResourceState holds the state of a resource that is used so that
// a provider can find and manage an existing resource as well as for
// storing attributes that are used to populate variables of child
// resources.
//
// Attributes has attributes about the created resource that are
// queryable in interpolation: "${type.id.attr}"
//
// Extra is just extra data that a provider can return that we store
// for later, but is not exposed in any way to the user.
//
type ResourceState struct {
	// This is filled in and managed by Terraform, and is the resource
	// type itself such as "mycloud_instance". If a resource provider sets
	// this value, it won't be persisted.
	Type string `json:"type"`

	// Dependencies are a list of things that this resource relies on
	// existing to remain intact. For example: an AWS instance might
	// depend on a subnet (which itself might depend on a VPC, and so
	// on).
	//
	// Terraform uses this information to build valid destruction
	// orders and to warn the user if they're destroying a resource that
	// another resource depends on.
	//
	// Things can be put into this list that may not be managed by
	// Terraform. If Terraform doesn't find a matching ID in the
	// overall state, then it assumes it isn't managed and doesn't
	// worry about it.
	Dependencies []string `json:"depends_on"`

	// Primary is the current active instance for this resource.
	// It can be replaced but only after a successful creation.
	// This is the instances on which providers will act.
	Primary *InstanceState `json:"primary"`

	// Deposed is used in the mechanics of CreateBeforeDestroy: the existing
	// Primary is Deposed to get it out of the way for the replacement Primary to
	// be created by Apply. If the replacement Primary creates successfully, the
	// Deposed instance is cleaned up.
	//
	// If there were problems creating the replacement Primary, the Deposed
	// instance and the (now tainted) replacement Primary will be swapped so the
	// tainted replacement will be cleaned up instead.
	//
	// An instance will remain in the Deposed list until it is successfully
	// destroyed and purged.
	Deposed []*InstanceState `json:"deposed"`

	// Provider is used when a resource is connected to a provider with an alias.
	// If this string is empty, the resource is connected to the default provider,
	// e.g. "aws_instance" goes with the "aws" provider.
	// If the resource block contained a "provider" key, that value will be set here.
	Provider string `json:"provider"`

	mu sync.Mutex
}

func (s *ResourceState) Lock()   { s.mu.Lock() }
func (s *ResourceState) Unlock() { s.mu.Unlock() }

// Equal tests whether two ResourceStates are equal.
func (s *ResourceState) Equal(other *ResourceState) bool {
	s.Lock()
	defer s.Unlock()

	if s.Type != other.Type {
		return false
	}

	if s.Provider != other.Provider {
		return false
	}

	// Dependencies must be equal
	sort.Strings(s.Dependencies)
	sort.Strings(other.Dependencies)
	if len(s.Dependencies) != len(other.Dependencies) {
		return false
	}
	for i, d := range s.Dependencies {
		if other.Dependencies[i] != d {
			return false
		}
	}

	// States must be equal
	if !s.Primary.Equal(other.Primary) {
		return false
	}

	return true
}

// Taint marks a resource as tainted.
func (s *ResourceState) Taint() {
	s.Lock()
	defer s.Unlock()

	if s.Primary != nil {
		s.Primary.Tainted = true
	}
}

// Untaint unmarks a resource as tainted.
func (s *ResourceState) Untaint() {
	s.Lock()
	defer s.Unlock()

	if s.Primary != nil {
		s.Primary.Tainted = false
	}
}

// ProviderAddr returns the provider address for the receiver, by parsing the
// string representation saved in state. An error can be returned if the
// value in state is corrupt.
func (s *ResourceState) ProviderAddr() (addrs.AbsProviderConfig, error) {
	var diags tfdiags.Diagnostics

	str := s.Provider
	traversal, travDiags := hclsyntax.ParseTraversalAbs([]byte(str), "", hcl.Pos{Line: 1, Column: 1})
	diags = diags.Append(travDiags)
	if travDiags.HasErrors() {
		return addrs.AbsProviderConfig{}, diags.Err()
	}

	addr, addrDiags := addrs.ParseAbsProviderConfig(traversal)
	diags = diags.Append(addrDiags)
	return addr, diags.Err()
}

func (s *ResourceState) String() string {
	s.Lock()
	defer s.Unlock()

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Type = %s", s.Type))
	return buf.String()
}

// InstanceState is used to track the unique state information belonging
// to a given instance.
type InstanceState struct {
	// A unique ID for this resource. This is opaque to Terraform
	// and is only meant as a lookup mechanism for the providers.
	ID string `json:"id"`

	// Attributes are basic information about the resource. Any keys here
	// are accessible in variable format within Terraform configurations:
	// ${resourcetype.name.attribute}.
	Attributes map[string]string `json:"attributes"`

	// Ephemeral is used to store any state associated with this instance
	// that is necessary for the Terraform run to complete, but is not
	// persisted to a state file.
	Ephemeral EphemeralState `json:"-"`

	// Meta is a simple K/V map that is persisted to the State but otherwise
	// ignored by Terraform core. It's meant to be used for accounting by
	// external client code. The value here must only contain Go primitives
	// and collections.
	Meta map[string]interface{} `json:"meta"`

	// Tainted is used to mark a resource for recreation.
	Tainted bool `json:"tainted"`

	mu sync.Mutex
}

func (s *InstanceState) Lock()   { s.mu.Lock() }
func (s *InstanceState) Unlock() { s.mu.Unlock() }

func (s *InstanceState) init() {
	s.Lock()
	defer s.Unlock()

	if s.Attributes == nil {
		s.Attributes = make(map[string]string)
	}
	if s.Meta == nil {
		s.Meta = make(map[string]interface{})
	}
	s.Ephemeral.init()
}

// NewInstanceStateShimmedFromValue is a shim method to lower a new-style
// object value representing the attributes of an instance object into the
// legacy InstanceState representation.
//
// This is for shimming to old components only and should not be used in new code.
func NewInstanceStateShimmedFromValue(state cty.Value, schemaVersion int) *InstanceState {
	attrs := hcl2shim.FlatmapValueFromHCL2(state)
	return &InstanceState{
		ID:         attrs["id"],
		Attributes: attrs,
		Meta: map[string]interface{}{
			"schema_version": schemaVersion,
		},
	}
}

// AttrsAsObjectValue shims from the legacy InstanceState representation to
// a new-style cty object value representation of the state attributes, using
// the given type for guidance.
//
// The given type must be the implied type of the schema of the resource type
// of the object whose state is being converted, or the result is undefined.
//
// This is for shimming from old components only and should not be used in
// new code.
func (s *InstanceState) AttrsAsObjectValue(ty cty.Type) (cty.Value, error) {
	if s == nil {
		// if the state is nil, we need to construct a complete cty.Value with
		// null attributes, rather than a single cty.NullVal(ty)
		s = &InstanceState{}
	}

	if s.Attributes == nil {
		s.Attributes = map[string]string{}
	}

	// make sure ID is included in the attributes. The InstanceState.ID value
	// takes precedence.
	if s.ID != "" {
		s.Attributes["id"] = s.ID
	}

	return hcl2shim.HCL2ValueFromFlatmap(s.Attributes, ty)
}

// Copy all the Fields from another InstanceState
func (s *InstanceState) Set(from *InstanceState) {
	s.Lock()
	defer s.Unlock()

	from.Lock()
	defer from.Unlock()

	s.ID = from.ID
	s.Attributes = from.Attributes
	s.Ephemeral = from.Ephemeral
	s.Meta = from.Meta
	s.Tainted = from.Tainted
}

func (s *InstanceState) DeepCopy() *InstanceState {
	copy, err := copystructure.Config{Lock: true}.Copy(s)
	if err != nil {
		panic(err)
	}

	return copy.(*InstanceState)
}

func (s *InstanceState) Empty() bool {
	if s == nil {
		return true
	}
	s.Lock()
	defer s.Unlock()

	return s.ID == ""
}

func (s *InstanceState) Equal(other *InstanceState) bool {
	// Short circuit some nil checks
	if s == nil || other == nil {
		return s == other
	}
	s.Lock()
	defer s.Unlock()

	// IDs must be equal
	if s.ID != other.ID {
		return false
	}

	// Attributes must be equal
	if len(s.Attributes) != len(other.Attributes) {
		return false
	}
	for k, v := range s.Attributes {
		otherV, ok := other.Attributes[k]
		if !ok {
			return false
		}

		if v != otherV {
			return false
		}
	}

	// Meta must be equal
	if len(s.Meta) != len(other.Meta) {
		return false
	}
	if s.Meta != nil && other.Meta != nil {
		// We only do the deep check if both are non-nil. If one is nil
		// we treat it as equal since their lengths are both zero (check
		// above).
		//
		// Since this can contain numeric values that may change types during
		// serialization, let's compare the serialized values.
		sMeta, err := json.Marshal(s.Meta)
		if err != nil {
			// marshaling primitives shouldn't ever error out
			panic(err)
		}
		otherMeta, err := json.Marshal(other.Meta)
		if err != nil {
			panic(err)
		}

		if !bytes.Equal(sMeta, otherMeta) {
			return false
		}
	}

	if s.Tainted != other.Tainted {
		return false
	}

	return true
}

// MergeDiff takes a ResourceDiff and merges the attributes into
// this resource state in order to generate a new state. This new
// state can be used to provide updated attribute lookups for
// variable interpolation.
//
// If the diff attribute requires computing the value, and hence
// won't be available until apply, the value is replaced with the
// computeID.
func (s *InstanceState) MergeDiff(d *InstanceDiff) *InstanceState {
	result := s.DeepCopy()
	if result == nil {
		result = new(InstanceState)
	}
	result.init()

	if s != nil {
		s.Lock()
		defer s.Unlock()
		for k, v := range s.Attributes {
			result.Attributes[k] = v
		}
	}
	if d != nil {
		for k, diff := range d.CopyAttributes() {
			if diff.NewRemoved {
				delete(result.Attributes, k)
				continue
			}
			if diff.NewComputed {
				result.Attributes[k] = hcl2shim.UnknownVariableValue
				continue
			}

			result.Attributes[k] = diff.New
		}
	}

	return result
}

func (s *InstanceState) String() string {
	notCreated := "<not created>"

	if s == nil {
		return notCreated
	}

	s.Lock()
	defer s.Unlock()

	var buf bytes.Buffer

	if s.ID == "" {
		return notCreated
	}

	buf.WriteString(fmt.Sprintf("ID = %s\n", s.ID))

	attributes := s.Attributes
	attrKeys := make([]string, 0, len(attributes))
	for ak, _ := range attributes {
		if ak == "id" {
			continue
		}

		attrKeys = append(attrKeys, ak)
	}
	sort.Strings(attrKeys)

	for _, ak := range attrKeys {
		av := attributes[ak]
		buf.WriteString(fmt.Sprintf("%s = %s\n", ak, av))
	}

	buf.WriteString(fmt.Sprintf("Tainted = %t\n", s.Tainted))

	return buf.String()
}

// EphemeralState is used for transient state that is only kept in-memory
type EphemeralState struct {
	// ConnInfo is used for the providers to export information which is
	// used to connect to the resource for provisioning. For example,
	// this could contain SSH or WinRM credentials.
	ConnInfo map[string]string `json:"-"`

	// Type is used to specify the resource type for this instance. This is only
	// required for import operations (as documented). If the documentation
	// doesn't state that you need to set this, then don't worry about
	// setting it.
	Type string `json:"-"`
}

func (e *EphemeralState) init() {
	if e.ConnInfo == nil {
		e.ConnInfo = make(map[string]string)
	}
}

func (e *EphemeralState) DeepCopy() *EphemeralState {
	copy, err := copystructure.Config{Lock: true}.Copy(e)
	if err != nil {
		panic(err)
	}

	return copy.(*EphemeralState)
}

// ErrNoState is returned by ReadState when the io.Reader contains no data
var ErrNoState = errors.New("no state")

func ReadStateV1(jsonBytes []byte) (*stateV1, error) {
	v1State := &stateV1{}
	if err := json.Unmarshal(jsonBytes, v1State); err != nil {
		return nil, fmt.Errorf("Decoding state file failed: %v", err)
	}

	if v1State.Version != 1 {
		return nil, fmt.Errorf("Decoded state version did not match the decoder selection: "+
			"read %d, expected 1", v1State.Version)
	}

	return v1State, nil
}

// resourceNameSort implements the sort.Interface to sort name parts lexically for
// strings and numerically for integer indexes.
type resourceNameSort []string

func (r resourceNameSort) Len() int      { return len(r) }
func (r resourceNameSort) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

func (r resourceNameSort) Less(i, j int) bool {
	iParts := strings.Split(r[i], ".")
	jParts := strings.Split(r[j], ".")

	end := len(iParts)
	if len(jParts) < end {
		end = len(jParts)
	}

	for idx := 0; idx < end; idx++ {
		if iParts[idx] == jParts[idx] {
			continue
		}

		// sort on the first non-matching part
		iInt, iIntErr := strconv.Atoi(iParts[idx])
		jInt, jIntErr := strconv.Atoi(jParts[idx])

		switch {
		case iIntErr == nil && jIntErr == nil:
			// sort numerically if both parts are integers
			return iInt < jInt
		case iIntErr == nil:
			// numbers sort before strings
			return true
		case jIntErr == nil:
			return false
		default:
			return iParts[idx] < jParts[idx]
		}
	}

	return r[i] < r[j]
}
