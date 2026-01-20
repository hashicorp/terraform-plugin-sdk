package schema

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
)

func transformAttributes(in cty.Value, s Schema) (*cty.Value, error) {
	ty := in.Type()
	out := cty.NullVal(ty)

	if s.Required {
		// This will be set with a valid value, so return it
		return &in, nil
	}

	if s.Computed && !s.Optional {
		// Computed only should always be null
		return &out, nil
	}

	if s.Optional {
		switch {
		case ty.Equals(cty.String):
			if in.AsString() != "" {
				// There is a non-empty value here, return it
				return &in, nil
			}

			d, err := s.DefaultValue()
			if err != nil {
				return nil, fmt.Errorf("getting default value from schema: %+v", err)
			}
			if d != nil {
				if v, ok := d.(string); ok {
					out = cty.StringVal(v)
				}
			}
		case ty.Equals(cty.Bool):
			// Bools should be passed along as they are
			out = in
			//if in.RawEquals(cty.True) {
			//	return &in, nil
			//}
			//
			//d, err := s.DefaultValue()
			//if err != nil {
			//	return nil, fmt.Errorf("getting default value from schema: %+v", err)
			//}
			//if d != nil {
			//	if v, ok := d.(bool); ok {
			//		out = cty.BoolVal(v)
			//	}
			//}
		case ty.Equals(cty.Number):
			if !in.RawEquals(cty.NumberIntVal(0)) {
				return &in, nil
			}
			d, err := s.DefaultValue()
			if err != nil {
				return nil, fmt.Errorf("getting default value from schema: %+v", err)
			}
			if d != nil {
				if v, ok := d.(int); ok {
					out = cty.NumberIntVal(int64(v))
				}
			}
		default:
			return nil, fmt.Errorf("unsupported primitive type %s", ty)
		}
	}
	return &out, nil
}

func transformMaps(in cty.Value, s Schema) (*cty.Value, error) {
	attrs := in.AsValueMap()

	if s.Computed && !s.Optional || len(attrs) == 0 {
		nullVal := cty.NullVal(in.Type())
		return &nullVal, nil
	}

	return &in, nil
}

func transformObjects(in cty.Value, s Schema) (*cty.Value, error) {
	attrs := in.AsValueMap()
	out := make(map[string]cty.Value)

	if s.Computed && !s.Optional {
		nullVal := cty.NullVal(in.Type())
		return &nullVal, nil
	}

	r, ok := s.Elem.(*Resource)
	if !ok {
		return nil, fmt.Errorf("expected schema Elem to be of type *Resource, got %T", s.Elem)
	}

	for k, v := range attrs {
		rs, ok := r.Schema[k]
		if !ok {
			return nil, fmt.Errorf("no schema found for %q", k)
		}
		if v.Type() == cty.String || v.Type() == cty.Number || v.Type() == cty.Bool {
			newVal, err := transformAttributes(v, *rs)
			if err != nil {
				return nil, fmt.Errorf("transforming attribute %q: %+v", k, err)
			}
			out[k] = *newVal
			continue
		}
		if v.Type().IsListType() || v.Type().IsSetType() {
			newVal, err := transformListsOrSets(v, *rs)
			if err != nil {
				return nil, fmt.Errorf("transforming list/set attribute %q: %+v", k, err)
			}
			out[k] = *newVal
			continue
		}
		if v.Type().IsObjectType() {
			newVal, err := transformObjects(v, *rs)
			if err != nil {
				return nil, fmt.Errorf("transforming nested object attribute %q: %+v", k, err)
			}
			out[k] = *newVal
			continue
		}
		if v.Type().IsMapType() {
			newVal, err := transformMaps(v, *rs)
			if err != nil {
				return nil, fmt.Errorf("transforming nested object attribute %q: %+v", k, err)
			}
			out[k] = *newVal
			continue
		}
	}

	outVal := cty.ObjectVal(out)
	return &outVal, nil
}

func transformListsOrSets(in cty.Value, s Schema) (*cty.Value, error) {
	vals := in.AsValueSlice()
	out := make([]cty.Value, 0)

	if len(vals) == 0 || s.Computed && !s.Optional {
		nullVal := cty.NullVal(in.Type())
		return &nullVal, nil
	}

	//if s.Required {
	//	// These should be set so return them
	//	return &in, nil
	//}

	for _, v := range in.AsValueSlice() {
		if v.Type() == cty.String || v.Type() == cty.Number || v.Type() == cty.Bool {
			newVal, err := transformAttributes(v, s)
			if err != nil {
				return nil, fmt.Errorf("transforming list/set element: %+v", err)
			}
			out = append(out, *newVal)
			continue
		}
		if v.Type().IsListType() || v.Type().IsSetType() {
			newVal, err := transformListsOrSets(v, s)
			if err != nil {
				return nil, fmt.Errorf("transforming nested list/set element: %+v", err)
			}
			out = append(out, *newVal)
			continue
		}
		if v.Type().IsObjectType() {
			newVal, err := transformObjects(v, s)
			if err != nil {
				return nil, fmt.Errorf("transforming nested object element: %+v", err)
			}
			out = append(out, *newVal)
			continue
		}
	}

	outVal := cty.ListVal(out)
	return &outVal, nil
}
