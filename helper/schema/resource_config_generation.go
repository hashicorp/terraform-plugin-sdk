package schema

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
)

func processConflictsWith(conflictsWith []string, configVal cty.Value, curPath cty.Path) (cty.PathSet, error) {
	markedForNullification := cty.NewPathSet()
	nonNullKeys := make([]string, 0)
	curValMapPath := ctyPathToFlatmapPath(curPath)

	if len(conflictsWith) > 0 {

		nonNullKeys = append(nonNullKeys, curValMapPath)

		for i := range conflictsWith {
			key := conflictsWith[i]
			ctyPath := flatmapPathToCtyPath(key)

			// get cty.Value at that path
			val, err := ctyPath.Apply(configVal)
			if err != nil {
				// There's a possibility that the path does not
				// exist in the value because the parent attribute/block
				// value is null.
				continue
			}

			if !val.IsNull() {
				nonNullKeys = append(nonNullKeys, key)
			}
		}

		// only process conflictsWith values if there are more
		// than one non-nil values, otherwise just leave alone.
		if len(nonNullKeys) < 1 {
			return markedForNullification, nil
		}

		// sort the keys and find the first non-null value to keep
		sort.Strings(nonNullKeys)
		for k := range nonNullKeys {
			// don't null out the first val
			if k == 0 {
				continue
			}
			// there's a possibility that the current value path
			// indexes into a set, so use the cty.Path from
			// the calling function to avoid lossy conversion
			// from/to flatmap path.
			if nonNullKeys[k] == curValMapPath {
				markedForNullification.Add(curPath)
				continue
			}
			markedForNullification.Add(flatmapPathToCtyPath(nonNullKeys[k]))
		}
	}

	return markedForNullification, nil
}

func processExactlyOneOf(exactlyOneOf []string, configVal cty.Value, curPath cty.Path) (cty.PathSet, error) {
	markedForNullification := cty.NewPathSet()
	nonNullKeys := make([]string, 0)
	curValMapPath := ctyPathToFlatmapPath(curPath)

	if len(exactlyOneOf) > 0 {

		// Self-references are allowed in the ExactlyOneOf slice
		// so we need to avoid duplication for the current value path.
		selfReference := false

		for i := range exactlyOneOf {
			key := exactlyOneOf[i]
			if key == curValMapPath {
				selfReference = true
			}

			ctyPath := flatmapPathToCtyPath(key)

			// get cty.Value at that path
			val, err := ctyPath.Apply(configVal)
			if err != nil {
				// There's a possibility that the path does not
				// exist in the value because the parent attribute/block
				// value is null.
				continue
			}

			if !val.IsNull() {
				nonNullKeys = append(nonNullKeys, key)
			}
		}

		if !selfReference {
			nonNullKeys = append(nonNullKeys, curValMapPath)
		}

		// only process exactlyOneOf values if there are more
		// than one non-nil values, otherwise just leave alone.
		if len(nonNullKeys) < 1 {
			return markedForNullification, nil
		}

		// sort the keys and find the first non-null value to keep
		sort.Strings(nonNullKeys)
		for k := range nonNullKeys {
			// don't null out the first val
			if k == 0 {
				continue
			}
			// there's a possibility that the current value path
			// indexes into a set, so use the cty.Path from
			// the calling function to avoid lossy conversion
			// from/to flatmap path.
			if nonNullKeys[k] == curValMapPath {
				markedForNullification.Add(curPath)
				continue
			}
			markedForNullification.Add(flatmapPathToCtyPath(nonNullKeys[k]))
		}
	}

	return markedForNullification, nil
}

func processRequiredWith(requiredWith []string, configVal cty.Value, curPath cty.Path) (cty.PathSet, error) {
	markedForNullification := cty.NewPathSet()
	nonNullKeys := make([]string, 0)
	curValMapPath := ctyPathToFlatmapPath(curPath)

	if len(requiredWith) > 0 {

		// Self-references are allowed in the requiredWith slice
		// so we need to avoid duplication for the current value path.
		selfReference := false

		for i := range requiredWith {
			key := requiredWith[i]
			if key == curValMapPath {
				selfReference = true
			}

			ctyPath := flatmapPathToCtyPath(key)

			// get cty.Value at that path
			val, err := ctyPath.Apply(configVal)
			if err != nil {
				// There's a possibility that the path does not
				// exist in the value because the parent attribute/block
				// value is null.
				continue
			}

			if !val.IsNull() {
				nonNullKeys = append(nonNullKeys, key)
			}
		}

		if !selfReference {
			nonNullKeys = append(nonNullKeys, curValMapPath)
			requiredWith = append(requiredWith, curValMapPath)
		}

		if len(requiredWith) == len(nonNullKeys) {
			return markedForNullification, nil
		}

		for k := range nonNullKeys {
			// there's a possibility that the current value path
			// indexes into a set, so use the cty.Path from
			// the calling function to avoid lossy conversion
			// from/to flatmap path.
			if nonNullKeys[k] == curValMapPath {
				markedForNullification.Add(curPath)
				continue
			}
			markedForNullification.Add(flatmapPathToCtyPath(nonNullKeys[k]))
		}

	}

	return markedForNullification, nil
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
			ty := pv.Key.Type()
			switch ty {
			case cty.Number:
				i, _ := pv.Key.AsBigFloat().Int64()
				flatMapPath += fmt.Sprintf(".%d.", i)
			case cty.String:
				flatMapPath += fmt.Sprintf(".%s.", pv.Key.AsString())
			default:
				flatMapPath += fmt.Sprintf(".%d.", i)
			}
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
