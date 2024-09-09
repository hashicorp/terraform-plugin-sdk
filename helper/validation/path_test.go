// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestPathEquals(t *testing.T) {
	tests := map[string]struct {
		p     cty.Path
		other cty.Path
		want  bool
	}{
		"null paths returns true": {
			p:     nil,
			other: nil,
			want:  true,
		},
		"empty paths returns true": {
			p:     cty.Path{},
			other: cty.Path{},
			want:  true,
		},
		"exact same path returns true": {
			p:     cty.GetAttrPath("attribute").Index(cty.NumberIntVal(1)).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.NumberIntVal(1)).GetAttr("nestedAttribute"),
			want:  true,
		},
		"path with unknown number index returns true": {
			p:     cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.Number)).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.NumberIntVal(1)).GetAttr("nestedAttribute"),
			want:  true,
		},
		"other path with unknown number index returns true": {
			p:     cty.GetAttrPath("attribute").Index(cty.NumberIntVal(1)).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.Number)).GetAttr("nestedAttribute"),
			want:  true,
		},
		"both paths with unknown number index returns true": {
			p:     cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.Number)).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.Number)).GetAttr("nestedAttribute"),
			want:  true,
		},
		"path with unknown string index returns true": {
			p:     cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.String)).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.StringVal("key")).GetAttr("nestedAttribute"),
			want:  true,
		},
		"other path with unknown string index returns true": {
			p:     cty.GetAttrPath("attribute").Index(cty.StringVal("key")).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.String)).GetAttr("nestedAttribute"),
			want:  true,
		},
		"both paths with unknown string index returns true": {
			p:     cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.String)).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.String)).GetAttr("nestedAttribute"),
			want:  true,
		},
		"path with unknown object index returns true": {
			p: cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.Object(nil))).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.ObjectVal(
				map[string]cty.Value{
					"oldAttribute":       cty.StringVal("old"),
					"writeOnlyAttribute": cty.StringVal("writeOnly"),
				},
			)).GetAttr("nestedAttribute"),
			want: true,
		},
		"other path with unknown object index returns true": {
			p: cty.GetAttrPath("attribute").Index(cty.ObjectVal(
				map[string]cty.Value{
					"oldAttribute":       cty.StringVal("old"),
					"writeOnlyAttribute": cty.StringVal("writeOnly"),
				},
			)).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.Object(nil))).GetAttr("nestedAttribute"),
			want:  true,
		},
		"both paths with unknown object index returns true": {
			p:     cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.Object(nil))).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.Object(nil))).GetAttr("nestedAttribute"),
			want:  true,
		},
		"paths with unequal steps return false": {
			p:     cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.Number)),
			other: cty.GetAttrPath("attribute").Index(cty.NumberIntVal(1)).GetAttr("nestedAttribute"),
			want:  false,
		},
		"paths with mismatched attribute names return false": {
			p:     cty.GetAttrPath("attribute").Index(cty.NumberIntVal(1)).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("incorrect").Index(cty.NumberIntVal(1)).GetAttr("nestedAttribute"),
			want:  false,
		},
		"paths with mismatched unknown index types return false": {
			p:     cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.Number)).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.String)).GetAttr("nestedAttribute"),
			want:  false,
		},
		"other path with unknown index, different type return false": {
			p:     cty.GetAttrPath("attribute").Index(cty.NumberIntVal(1)).GetAttr("nestedAttribute"),
			other: cty.GetAttrPath("attribute").Index(cty.UnknownVal(cty.String)).GetAttr("nestedAttribute"),
			want:  false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := PathEquals(tc.p, tc.other); got != tc.want {
				t.Errorf("PathEquals() = %v, want %v", got, tc.want)
			}
		})
	}
}
