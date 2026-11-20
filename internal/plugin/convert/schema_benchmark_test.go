// Copyright IBM Corp. 2019, 2026
// SPDX-License-Identifier: MPL-2.0

package convert

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/configschema"
)

var benchmarkSortedKeysResult []string
var benchmarkConfigSchemaToProtoResult *tfprotov5.SchemaBlock
var benchmarkConfigIdentitySchemaToProtoResult []*tfprotov5.ResourceIdentitySchemaAttribute

func BenchmarkSortedKeys(b *testing.B) {
	for _, size := range []int{0, 3, 10, 40} {
		attributes := make(map[string]*configschema.Attribute, size)
		for i := range size {
			attributes[fmt.Sprintf("attribute_%02d", i)] = &configschema.Attribute{}
		}

		b.Run(fmt.Sprintf("elements=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			var keys []string
			for b.Loop() {
				keys = sortedKeys(attributes)
			}
			benchmarkSortedKeysResult = keys
		})
	}
}

func BenchmarkConfigSchemaToProto(b *testing.B) {
	const nestedBlockAttributes = 2

	for _, tc := range []struct {
		name       string
		attributes int
		blocks     int
	}{
		{name: "attributes=3", attributes: 3},
		{name: "attributes=10", attributes: 10},
		{name: "attributes=40", attributes: 40},
		{name: "blocks=3", blocks: 3},
		{name: "blocks=10", blocks: 10},
		{name: "blocks=40", blocks: 40},
		{name: "attributes=10_blocks=3", attributes: 10, blocks: 3},
	} {
		schema := &configschema.Block{
			Attributes: make(map[string]*configschema.Attribute, tc.attributes),
			BlockTypes: make(map[string]*configschema.NestedBlock, tc.blocks),
		}
		for i := range tc.attributes {
			schema.Attributes[fmt.Sprintf("attribute_%02d", i)] = &configschema.Attribute{Type: cty.String}
		}
		for i := range tc.blocks {
			nested := &configschema.NestedBlock{
				Block:   configschema.Block{Attributes: make(map[string]*configschema.Attribute, nestedBlockAttributes)},
				Nesting: configschema.NestingSingle,
			}
			for j := range nestedBlockAttributes {
				nested.Attributes[fmt.Sprintf("attribute_%02d", j)] = &configschema.Attribute{Type: cty.String}
			}
			schema.BlockTypes[fmt.Sprintf("block_%02d", i)] = nested
		}

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			var result *tfprotov5.SchemaBlock
			for b.Loop() {
				result = ConfigSchemaToProto(context.Background(), schema)
			}
			benchmarkConfigSchemaToProtoResult = result
		})
	}
}

func BenchmarkConfigIdentitySchemaToProto(b *testing.B) {
	for _, size := range []int{1, 2, 3, 4, 5, 6} {
		schema := &configschema.Block{
			Attributes: make(map[string]*configschema.Attribute, size),
		}
		for i := range size {
			schema.Attributes[fmt.Sprintf("attribute_%02d", i)] = &configschema.Attribute{Type: cty.String}
		}

		b.Run(fmt.Sprintf("attributes=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			var result []*tfprotov5.ResourceIdentitySchemaAttribute
			for b.Loop() {
				result = ConfigIdentitySchemaToProto(context.Background(), schema)
			}
			benchmarkConfigIdentitySchemaToProtoResult = result
		})
	}
}
