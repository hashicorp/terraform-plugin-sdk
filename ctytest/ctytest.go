package ctytest

// TODO should this just be acctest package?

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-sdk/internal/addrs"
	"github.com/hashicorp/terraform-plugin-sdk/internal/configs/configschema"
	"github.com/hashicorp/terraform-plugin-sdk/internal/states"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zclconf/go-cty/cty"
)

type CtyCheckFunc func(*terraform.State, map[string]terraform.ResourceProvider) error

type CtyValueCheckFunc func(cty.Value) error

var GetAttrPath = cty.GetAttrPath

func shimJsonSchemaAttribute(jsonAttribute *tfjson.SchemaAttribute) *configschema.Attribute {
	attribute := configschema.Attribute{
		Type:            jsonAttribute.AttributeType,
		Description:     jsonAttribute.Description,
		DescriptionKind: configschema.StringPlain,
		Required:        jsonAttribute.Required,
		Optional:        jsonAttribute.Optional,
		Computed:        jsonAttribute.Computed,
		Sensitive:       jsonAttribute.Sensitive,
	}
	return &attribute
}

func shimJsonSchemaBlock(jsonBlock tfjson.SchemaBlock) *configschema.Block {
	block := configschema.Block{}

	for attributeName, attributeValue := range jsonBlock.Attributes {
		block.Attributes[attributeName] = shimJsonSchemaAttribute(attributeValue)
	}

	return &block

	// TODO KEM BLOCK RECURSION
}

func shimJsonProviderSchema(jsonProviderSchema *tfjson.ProviderSchema) *terraform.ProviderSchema {
	resourceTypes := map[string]*configschema.Block{}

	for resourceName, resourceSchema := range jsonProviderSchema.ResourceSchemas {
		shimmedResourceSchema := shimJsonSchemaBlock(*resourceSchema.Block)
		resourceTypes[resourceName] = shimmedResourceSchema
	}

	providerSchema := terraform.ProviderSchema{
		ResourceTypes: resourceTypes,
	}

	return &providerSchema
}

func shimSchemasFromJson(jsonSchema *tfjson.ProviderSchemas) *terraform.Schemas {
	schemas := new(terraform.Schemas)

	for providerName, providerSchema := range jsonSchema.Schemas {

		schemas.Providers[providerName] = shimJsonProviderSchema(providerSchema)
	}

	return schemas
}

// ComposeAggregateCheckFunc lets you compose multiple CtyValueCheckFuncs into
// a single CtyValueCheckFunc.
//
// Unlike ComposeCheckFunc, ComposeAggergateCheckFunc runs _all_ of the
// CtyValueCheckFuncs and aggregates failures.
func ComposeAggregateCheckFunc(fs ...CtyValueCheckFunc) CtyValueCheckFunc {
	return func(v cty.Value) error {
		var result *multierror.Error

		for i, f := range fs {
			if err := f(v); err != nil {
				result = multierror.Append(result, fmt.Errorf("Check %d/%d error: %s", i+1, len(fs), err))
			}
		}

		return result.ErrorOrNil()
	}

}

// ComposeCheckFunc lets you compose multiple CtyValueCheckFuncs into
// a single CtyValueCheckFunc.
func ComposeCheckFunc(fs ...CtyValueCheckFunc) CtyValueCheckFunc {
	return func(v cty.Value) error {
		for i, f := range fs {
			if err := f(v); err != nil {
				return fmt.Errorf("Check %d/%d error: %s", i+1, len(fs), err)
			}
		}

		return nil
	}
}

func getAbsResourceInstanceAddr(resourceAddr string) (*addrs.AbsResourceInstance, error) {
	r, tDiags := addrs.ParseAbsResourceInstanceStr(resourceAddr)
	if tDiags.HasErrors() {
		return nil, fmt.Errorf("error parsing resource address %s: %s", resourceAddr, tDiags.Err())
	}

	return &r, nil
}

func getResourceInstance(state *states.State, resourceAddr string) (*states.ResourceInstance, error) {
	absResourceInstanceAddr, tDiags := addrs.ParseAbsResourceInstanceStr(resourceAddr)
	if tDiags.HasErrors() {
		return nil, fmt.Errorf("error parsing resource address %s: %s", resourceAddr, tDiags.Err())
	}

	// get the addrs.AbsResourceInstance
	r := state.ResourceInstance(absResourceInstanceAddr)

	return r, nil
}

func ctyCheck(legacyState *terraform.State, providers map[string]terraform.ResourceProvider, name string, path cty.Path, checkFunc CtyValueCheckFunc) error {
	// shim to the new state format
	state, err := terraform.ShimLegacyState(legacyState)
	if err != nil {
		return err
	}

	absResourceInstanceAddr, err := getAbsResourceInstanceAddr(name)
	if err != nil {
		return err
	}

	r := absResourceInstanceAddr.ContainingResource().Resource

	providerName := state.RootModule().Resource(r).ProviderConfig.ProviderConfig.Type

	provider := providers[providerName]

	var schemaRequest terraform.ProviderSchemaRequest

	if r.Mode == addrs.ManagedResourceMode {
		schemaRequest = terraform.ProviderSchemaRequest{
			ResourceTypes: []string{r.Type},
		}
	} else { // assume addrs.DataResourceMode
		schemaRequest = terraform.ProviderSchemaRequest{
			DataSources: []string{r.Type},
		}
	}

	schema, err := provider.GetSchema(&schemaRequest)
	if err != nil {
		return err
	}

	resSchema := schema.ResourceTypes[r.Type]

	ms := state.RootModule()
	res := ms.Resources[name]

	for _, is := range res.Instances {
		if is.HasCurrent() {
			resInstObjSrc := is.Current

			obj, err := resInstObjSrc.Decode(resSchema.ImpliedType())
			if err != nil {
				return err
			}

			val, err := path.Apply(obj.Value)
			if err != nil {
				return err
			}
			fmt.Print(val)

			err = checkFunc(val)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// func attributeIsNotNull() CtyValueCheckFunc {
// 	return func(v cty.Value) error {
// 		ok := v.IsNull()
// 		if !ok {
// 			return fmt.Errorf("expected %s to be null, but it was non-null", v)
// 		}
// 		return nil
// 	}
// }

// func AttributeIsNotNull(address string, path cty.Path) CtyCheckFunc {
// 	return func(state *terraform.State, provider terraform.ResourceProvider) error {
// 		return ctyCheck(state, provider, address, path, attributeIsNotNull())

// 	}
// }

// func attributeEquals(other cty.Value) CtyValueCheckFunc {
// 	return func(v cty.Value) error {
// 		ok := v.RawEquals(other)
// 		if !ok {
// 			return fmt.Errorf("expected %s to equal %s", v, other)
// 		}
// 		return nil
// 	}
// }

// func AttributeEquals(address string, path cty.Path, value cty.Value) CtyCheckFunc {
// 	return func(state *terraform.State, provider terraform.ResourceProvider) error {
// 		return ctyCheck(state, provider, address, path, attributeEquals(value))
// 	}
// }

// TODO KEM
// func attributeEqualsToPtr(v cty.Value, other *ctyValue) bool {
// }
// func AttributeEqualsToPtr(address string, path cty.Path, value *cty.Value) {}

// func AttributesEqual(leftAddress string, leftPath cty.Path,
// 	rightAddress string, rightPath cty.Path) {
// }

func stringEquals(value string) CtyValueCheckFunc {
	return func(v cty.Value) error {
		s := cty.StringVal(value)
		ok := v.RawEquals(s)
		if !ok {
			return fmt.Errorf("expected %s to equal %s", v, s)
		}
		return nil
	}
}

func StringEquals(address string, path cty.Path, value string) CtyCheckFunc {
	return func(state *terraform.State, providers map[string]terraform.ResourceProvider) error {
		return ctyCheck(state, providers, address, path, stringEquals(value))
	}
}

// func stringMatchFunc(matchFunc func(string) error) CtyValueCheckFunc {
// 	return func(v cty.Value) {
// 		s := v.String()
// 		return matchFunc(s)
// 	}
// }

// func StringMatchFunc(address string, path cty.Path, matchFunc func(string) error) CtyCheckFunc {
// 	return func(state *terraform.State, provider *terraform.ResourceProvider) error {
// 		return ctyCheck(state, provider, address, path, stringMatchFunc())
// 	}
// }

// func intEquals(value int) CtyValueCheckFunc {
// 	i := cty.NumberIntVal(i)
// 	return func(v cty.Value) error {
// 		return v.Equals(i)
// 	}
// }

// func IntEquals(address string, path cty.Path, value int) CtyCheckFunc {
// 	return func(state *terraform.State, provider *terraform.ResourceProvider) error {
// 		return ctyCheck(state, provider, address, path, intEquals(value))
// 	}
// }

// func floatEquals(value float64) CtyValueCheckFunc {
// 	f := cty.NumberFloatVal(value)
// 	return func(v cty.Value) error {
// 		return v.Equals(f)
// 	}
// }

// func FloatEquals(address string, path cty.Path, value float64) CtyCheckFunc {
// 	return func(state *terraform.State, provider *terraform.ResourceProvider) error {
// 		return ctyCheck(state, provider, address, path, floatEquals(value))
// 	}
// }

// func isTrue() CtyValueCheckFunc {
// 	return func(v cty.Value) error {
// 		return v.Equals(cty.True)
// 	}
// }

// func IsTrue(address string, path cty.Path) CtyCheckFunc {
// 	return func(state *terraform.State, provider *terraform.ResourceProvider) error {
// 		return ctyCheck(state, provider, address, path, isTrue())
// 	}
// }

// func isFalse() CtyValueCheckFunc {
// 	return func(v cty.Value) error {
// 		return v.Equals(cty.False)
// 	}
// }

// func IsFalse(address string, path cty.Path) CtyCheckFunc {
// 	return func(state *terraform.State, provider *terraform.ResourceProvider) error {
// 		return ctyCheck(state, provider, address, path, isFalse())
// 	}
// }

// TODO KEM
// func setLengthEquals(length int) CtyValueCheckFunc {
// 	return func(v cty.Value) error {

// 	}
// }

// func SetLengthEquals(address string, path cty.Path, length int) {}

// func SetEquals(address string, path cty.Path, values []cty.Value) {}

// func ListLengthEquals(address string, path cty.Path, length int) {}

// func ListEquals(address string, path cty.Path, values []cty.Value) {}

// func MapLengthEquals(address string, path cty.Path, length int) {}

// func MapEquals(address string, path cty.Path, m map[string]cty.Value) {}
