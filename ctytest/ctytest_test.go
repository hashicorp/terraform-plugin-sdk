package ctytest

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestAttributeIsNull(t *testing.T) {

}

// func TestCheckResourceAttr_empty(t *testing.T) {
// 	s := terraform.NewState()
// 	s.AddModuleState(&terraform.ModuleState{
// 		Path: []string{"root"},
// 		Resources: map[string]*terraform.ResourceState{
// 			"test_resource": {
// 				Primary: &terraform.InstanceState{
// 					Attributes: map[string]string{
// 						"empty_list.#": "0",
// 						"empty_map.%":  "0",
// 					},
// 				},
// 			},
// 		},
// 	})

// 	for _, key := range []string{
// 		"empty_list.#",
// 		"empty_map.%",
// 		"missing_list.#",
// 		"missing_map.%",
// 	} {
// 		t.Run(key, func(t *testing.T) {
// 			check := TestCheckResourceAttr("test_resource", key, "0")
// 			if err := check(s); err != nil {
// 				t.Fatal(err)
// 			}
// 		})
// 	}
// }
