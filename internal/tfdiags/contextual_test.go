// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfdiags

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestGetAttribute(t *testing.T) {
	path := cty.Path{
		cty.GetAttrStep{Name: "foo"},
		cty.IndexStep{Key: cty.NumberIntVal(0)},
		cty.GetAttrStep{Name: "bar"},
	}

	d := AttributeValue(
		Error,
		"foo[0].bar",
		"detail",
		path,
	)

	p := GetAttribute(d)
	if !reflect.DeepEqual(path, p) {
		t.Fatalf("paths don't match:\nexpected: %#v\ngot: %#v", path, p)
	}
}
