package addrs

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// instanceKey represents the key of an instance within an object that
// contains multiple instances due to using "count" or "for_each" arguments
// in configuration.
//
// intKey and stringKey are the two implementations of this type. No other
// implementations are allowed. The single instance of an object that _isn't_
// using "count" or "for_each" is represented by NoKey, which is a nil
// InstanceKey.
type instanceKey interface {
	instanceKeySigil()
	String() string
}

// parseInstanceKey returns the instance key corresponding to the given value,
// which must be known and non-null.
//
// If an unknown or null value is provided then this function will panic. This
// function is intended to deal with the values that would naturally be found
// in a hcl.TraverseIndex, which (when parsed from source, at least) can never
// contain unknown or null values.
func parseInstanceKey(key cty.Value) (instanceKey, error) {
	switch key.Type() {
	case cty.String:
		return stringKey(key.AsString()), nil
	case cty.Number:
		var idx int
		err := gocty.FromCtyValue(key, &idx)
		return intKey(idx), err
	default:
		return NoKey, fmt.Errorf("either a string or an integer is required")
	}
}

// NoKey represents the absense of an instanceKey, for the single instance
// of a configuration object that does not use "count" or "for_each" at all.
var NoKey instanceKey

// intKey is the InstanceKey representation representing integer indices, as
// used when the "count" argument is specified or if for_each is used with
// a sequence type.
type intKey int

func (k intKey) instanceKeySigil() {
}

func (k intKey) String() string {
	return fmt.Sprintf("[%d]", int(k))
}

// stringKey is the InstanceKey representation representing string indices, as
// used when the "for_each" argument is specified with a map or object type.
type stringKey string

func (k stringKey) instanceKeySigil() {
}

func (k stringKey) String() string {
	// FIXME: This isn't _quite_ right because Go's quoted string syntax is
	// slightly different than HCL's, but we'll accept it for now.
	return fmt.Sprintf("[%q]", string(k))
}
