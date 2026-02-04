package schema

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/configschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/helper/hashcode"
)

// setComputedOnlyNullValues takes a cty.Value, and compares it to the schema setting any non-null
// values that are ComputedOnly to null.
func setComputedOnlyNullValues(val cty.Value, schema *configschema.Block) cty.Value {
	if !val.IsKnown() || val.IsNull() {
		return val
	}

	valMap := val.AsValueMap()
	newVals := make(map[string]cty.Value)

	for name, attr := range schema.Attributes {
		v := valMap[name]

		if attr.Computed && !attr.Optional && !v.IsNull() {
			newVals[name] = cty.NullVal(attr.Type)
			continue
		}

		newVals[name] = v
	}

	for name, blockS := range schema.BlockTypes {
		blockVal := valMap[name]
		if blockVal.IsNull() || !blockVal.IsKnown() {
			newVals[name] = blockVal
			continue
		}

		blockValType := blockVal.Type()
		blockElementType := blockS.Block.ImpliedType()

		// This switches on the value type here, so we can correctly switch
		// between Tuples/Lists and Maps/Objects.
		switch {
		case blockS.Nesting == configschema.NestingSingle || blockS.Nesting == configschema.NestingGroup:
			// NestingSingle is the only exception here, where we treat the
			// block directly as an object
			newVals[name] = setComputedOnlyNullValues(blockVal, &blockS.Block)

		case blockValType.IsSetType(), blockValType.IsListType(), blockValType.IsTupleType():
			listVals := blockVal.AsValueSlice()
			newListVals := make([]cty.Value, 0, len(listVals))

			for _, v := range listVals {
				newListVals = append(newListVals, setComputedOnlyNullValues(v, &blockS.Block))
			}

			switch {
			case blockValType.IsSetType():
				switch len(newListVals) {
				case 0:
					newVals[name] = cty.SetValEmpty(blockElementType)
				default:
					newVals[name] = cty.SetVal(newListVals)
				}
			case blockValType.IsListType():
				switch len(newListVals) {
				case 0:
					newVals[name] = cty.ListValEmpty(blockElementType)
				default:
					newVals[name] = cty.ListVal(newListVals)
				}
			case blockValType.IsTupleType():
				newVals[name] = cty.TupleVal(newListVals)
			}

		case blockValType.IsMapType(), blockValType.IsObjectType():
			mapVals := blockVal.AsValueMap()
			newMapVals := make(map[string]cty.Value)

			for k, v := range mapVals {
				newMapVals[k] = setComputedOnlyNullValues(v, &blockS.Block)
			}

			switch {
			case blockValType.IsMapType():
				switch len(newMapVals) {
				case 0:
					newVals[name] = cty.MapValEmpty(blockElementType)
				default:
					newVals[name] = cty.MapVal(newMapVals)
				}
			case blockValType.IsObjectType():
				if len(newMapVals) == 0 {
					// We need to populate empty values to make a valid object.
					for attr, ty := range blockElementType.AttributeTypes() {
						newMapVals[attr] = cty.NullVal(ty)
					}
				}
				newVals[name] = cty.ObjectVal(newMapVals)
			}

		default:
			panic(fmt.Sprintf("failed to set null values for nested block %q:%#v", name, blockValType))
		}
	}

	return cty.ObjectVal(newVals)
}

func generateConflictsWith(schemaMap map[string]*Schema, path string) map[string][]string {
	result := make(map[string][]string)

	curPath := path

	//for subK, s := range schema {
	//	key := subK
	//	if k != "" {
	//		key = fmt.Sprintf("%s.%s", k, subK)
	//	}
	//	diags = append(diags, m.validate(key, s, c, append(path, cty.GetAttrStep{Name: subK}))...)
	//}

	for m, v := range schemaMap {
		if curPath == "" {
			curPath = m
		} else {
			curPath = curPath + m
		}

		if len(v.ConflictsWith) != 0 {
			conflictsWith := v.ConflictsWith

			conflictsWith = append(conflictsWith, curPath)
			sort.Strings(conflictsWith)

			key := hashcode.Strings(conflictsWith)
			_, ok := result[key]
			if !ok {
				result[key] = conflictsWith
			}
		}

		if v.Elem != nil {
			switch t := v.Elem.(type) {
			case *Resource:
				blockResults := generateConflictsWith(t.SchemaMap(), curPath+".0.")
				for blockKey, blockVal := range blockResults {
					result[blockKey] = blockVal
				}
			case *Schema:
				// Todo: check if we need to implement this
			}
		}
	}

	return result
}

func configGenerationSchemaMap(schemaMap map[string]*Schema, configGenSchemaMap map[string]*configGenSchema, parentPath string) {

	for attrName, schema := range schemaMap {
		var curPath string
		if parentPath == "" {
			curPath = attrName
		} else {
			curPath = parentPath + attrName
		}

		genSchema := &configGenSchema{
			Required:      schema.Required,
			Optional:      schema.Optional,
			Computed:      schema.Computed,
			Sensitive:     schema.Sensitive,
			Deprecated:    schema.Deprecated,
			ConflictsWith: schema.ConflictsWith,
			ExactlyOneOf:  schema.ExactlyOneOf,
			AtLeastOneOf:  schema.AtLeastOneOf,
			RequiredWith:  schema.RequiredWith,
		}
		configGenSchemaMap[curPath] = genSchema

		if schema.Elem != nil {
			switch t := schema.Elem.(type) {
			case *Resource:
				configGenerationSchemaMap(t.SchemaMap(), configGenSchemaMap, curPath+".0.")

			case *Schema:
				// Todo: check if we need to implement this
			}
		}
	}
}

func newResourceConfigShimmedComputedKeys(val cty.Value, path string) []string {
	var ret []string
	ty := val.Type()

	if val.IsNull() {
		return ret
	}

	if !val.IsKnown() {
		// we shouldn't have an entirely unknown resource, but prevent empty
		// strings just in case
		if len(path) > 0 {
			ret = append(ret, path)
		}
		return ret
	}

	if path != "" {
		path += "."
	}
	switch {
	case ty.IsListType(), ty.IsTupleType(), ty.IsSetType():
		i := 0
		for it := val.ElementIterator(); it.Next(); i++ {
			_, subVal := it.Element()
			keys := newResourceConfigShimmedComputedKeys(subVal, fmt.Sprintf("%s%d", path, i))
			ret = append(ret, keys...)
		}

	case ty.IsMapType(), ty.IsObjectType():
		for it := val.ElementIterator(); it.Next(); {
			subK, subVal := it.Element()
			keys := newResourceConfigShimmedComputedKeys(subVal, fmt.Sprintf("%s%s", path, subK.AsString()))
			ret = append(ret, keys...)
		}
	case ty.IsPrimitiveType():
		ret = append(ret, path)
	}

	return ret
}

func ctyPathToFlatmapPath(p cty.Path) string {
	pathLen := len(p)
	flatMapPath := ""
	for i := range p {
		pv := p[i]
		switch pv := pv.(type) {
		case cty.GetAttrStep:
			flatMapPath += pv.Name
		case cty.IndexStep:
			if i == pathLen-1 {
				break
			}
			pv.GoString()
			flatMapPath += ".0."
		default:
			return flatMapPath
		}
	}

	return flatMapPath
}

func flatmapPathToCtyPath(fp string) cty.Path {
	result := cty.Path{}
	if fp == "" {
		return result
	}

	steps := strings.Split(fp, ".")
	for i := range steps {
		curStep := steps[i]

		var index int
		index, err := strconv.Atoi(curStep)
		if err != nil {
			result = result.GetAttr(curStep)
			continue
		}
		result = result.IndexInt(index)

	}

	return result
}

func processConflictsWith(configVal cty.Value, schema map[string]*Schema) (cty.Value, []string, error) {
	genSchema := make(map[string]*configGenSchema)
	markedForNullification := make([]string, 0)

	configGenerationSchemaMap(schema, genSchema, "")

	// Handle conflicts with
	configVal, err := cty.Transform(configVal, func(path cty.Path, val cty.Value) (cty.Value, error) {
		if len(path) == 0 {
			return val, nil
		}
		curVal := val

		curValMapPath := ctyPathToFlatmapPath(path)
		sch, ok := genSchema[curValMapPath]
		if !ok {
			return curVal, fmt.Errorf("Cannot retrieve config schema at key %s", curValMapPath)
		}

		conflictsWith := make(map[string]cty.Value)
		conflictsWithKeys := make([]string, 0)
		if len(sch.ConflictsWith) > 0 {

			if !curVal.IsNull() {
				conflictsWith[curValMapPath] = curVal
				conflictsWithKeys = append(conflictsWithKeys, curValMapPath)
			}

			for i := range sch.ConflictsWith {
				key := sch.ConflictsWith[i]
				ctyPath := flatmapPathToCtyPath(key)

				// get cty.Value at that path
				val, err := ctyPath.Apply(configVal)
				if err != nil {
					return cty.Value{}, err
				}

				if !val.IsNull() {
					conflictsWith[key] = val
					conflictsWithKeys = append(conflictsWithKeys, key)
				}
			}

			// only process conflictsWith values if there are more
			// than one non-nil values, otherwise just leave alone.
			if len(conflictsWithKeys) > 1 {
				// sort the keys and find the first non-null value to keep
				sort.Strings(conflictsWithKeys)
				firstKey := conflictsWithKeys[0]
				delete(conflictsWith, firstKey)

				// if the current value is not the first key,
				// nullify it immediately.
				if firstKey != curValMapPath {
					curVal = cty.NullVal(curVal.Type())
					delete(conflictsWith, curValMapPath)
				}
			}
		}

		for k := range conflictsWith {
			markedForNullification = append(markedForNullification, k)
		}

		return curVal, nil
	})
	if err != nil {
		return configVal, markedForNullification, err
	}

	return configVal, markedForNullification, nil
}

type configGenSchema struct {
	Required      bool
	Optional      bool
	Computed      bool
	Sensitive     bool
	Deprecated    string
	ConflictsWith []string
	ExactlyOneOf  []string
	AtLeastOneOf  []string
	RequiredWith  []string
}
