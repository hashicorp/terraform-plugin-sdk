// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestShimState(t *testing.T) {
	type expectedError struct {
		Prefix string
	}

	testCases := []struct {
		Name          string
		RawState      string
		ExpectedState *terraform.State
		ExpectedErr   *expectedError
	}{
		{
			"empty",
			`{
  "format_version": "0.1"
}`,
			&terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:         []string{"root"},
						Outputs:      map[string]*terraform.OutputState{},
						Resources:    map[string]*terraform.ResourceState{},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"simple outputs",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {
      "bool": {
        "sensitive": false,
        "value": true
      },
      "int": {
        "sensitive": false,
        "value": 42
      },
      "float": {
        "sensitive": false,
        "value": 1.4
      },
      "string": {
        "sensitive": false,
        "value": "test"
      }
    },
    "root_module": {}
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Outputs: map[string]*terraform.OutputState{
							"bool": {
								Type:      "string",
								Value:     "true",
								Sensitive: false,
							},
							"int": {
								Type:      "string",
								Value:     "42",
								Sensitive: false,
							},
							"float": {
								Type:      "string",
								Value:     "1.4",
								Sensitive: false,
							},
							"string": {
								Type:      "string",
								Value:     "test",
								Sensitive: false,
							},
						},
						Resources:    map[string]*terraform.ResourceState{},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"complex outputs",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {
      "empty_list": {
        "sensitive": false,
        "value": []
      },
      "list_of_strings": {
        "sensitive": false,
        "value": ["first", "second", "third"]
      },
      "map_of_strings": {
        "sensitive": false,
        "value": {
        	"hello": "world",
        	"foo": "bar"
        }
      },
      "list_of_int": {
        "sensitive": false,
        "value": [1, 4, 9]
      },
      "list_of_float": {
        "sensitive": false,
        "value": [1.2, 4.2, 9.8]
      },
      "list_of_bool": {
        "sensitive": false,
        "value": [true, false, true]
      }
    },
    "root_module": {}
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Outputs: map[string]*terraform.OutputState{
							"empty_list": {
								Type:      "list",
								Value:     []interface{}{},
								Sensitive: false,
							},
							"list_of_strings": {
								Type: "list",
								Value: []interface{}{
									"first", "second", "third",
								},
								Sensitive: false,
							},
							"map_of_strings": {
								Type: "map",
								Value: map[string]interface{}{
									"hello": "world",
									"foo":   "bar",
								},
								Sensitive: false,
							},
							"list_of_int": {
								Type: "list",
								Value: []interface{}{
									json.Number("1"),
									json.Number("4"),
									json.Number("9"),
								},
								Sensitive: false,
							},
							"list_of_float": {
								Type: "list",
								Value: []interface{}{
									json.Number("1.2"),
									json.Number("4.2"),
									json.Number("9.8"),
								},
								Sensitive: false,
							},
							"list_of_bool": {
								Type: "list",
								Value: []interface{}{
									true, false, true,
								},
								Sensitive: false,
							},
						},
						Resources:    map[string]*terraform.ResourceState{},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"output with slice of slices",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {
      "list_of_lists": {
        "sensitive": false,
        "value": [
          ["one", "two"],
          ["blue", "green", "red"]
        ]
      }
    },
    "root_module": {}
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Outputs: map[string]*terraform.OutputState{
							"list_of_lists": {
								Type: "list",
								Value: []interface{}{
									[]interface{}{"one", "two"},
									[]interface{}{"blue", "green", "red"},
								},
								Sensitive: false,
							},
						},
						Resources:    map[string]*terraform.ResourceState{},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"output with slice of maps",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {
      "list_of_maps": {
        "sensitive": false,
        "value": [
          {
            "rule": "allow",
            "port": 443,
            "allow_bool": true
          },
          {
            "rule": "deny",
            "port": 80,
            "allow_bool": false
          }
        ]
      }
    },
    "root_module": {}
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Outputs: map[string]*terraform.OutputState{
							"list_of_maps": {
								Type: "list",
								Value: []interface{}{
									map[string]interface{}{
										"allow_bool": true,
										"port":       json.Number("443"),
										"rule":       "allow",
									},
									map[string]interface{}{
										"allow_bool": false,
										"port":       json.Number("80"),
										"rule":       "deny",
									},
								},
								Sensitive: false,
							},
						},
						Resources:    map[string]*terraform.ResourceState{},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"output with nested map",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {
      "map_of_maps": {
        "sensitive": false,
        "value": {
        	"hello": {
        	  "whole": "world"
        	},
        	"foo": "bar"
        }
      }
    },
    "root_module": {}
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Outputs: map[string]*terraform.OutputState{
							"map_of_maps": {
								Type: "map",
								Value: map[string]interface{}{
									"hello": map[string]interface{}{
										"whole": "world",
									},
									"foo": "bar",
								},
								Sensitive: false,
							},
						},
						Resources:    map[string]*terraform.ResourceState{},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"invalid address",
			`{
  "format_version": "0.1",
  "values": {
    "root_module": {
      "address": "blah"
    }
  }
}`,
			nil,
			&expectedError{Prefix: "Invalid module instance address"},
		},
		{
			"resource with primitive attributes",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {},
    "root_module": {
      "resources": [
        {
          "address": "cloud_vpc.main",
          "mode": "managed",
          "type": "cloud_vpc",
          "name": "main",
          "provider_name": "cloud",
          "schema_version": 1,
          "values": {
          	"id": "123999",
            "string_field": "hello world",
            "bool_field_1": false,
            "bool_field_2": true,
            "null_field": null,
            "empty_string": "",
            "number_field": 42
          }
        }
      ]
    }
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"cloud_vpc.main": {
								Type:     "cloud_vpc",
								Provider: "cloud",
								Primary: &terraform.InstanceState{
									ID: "123999",
									Meta: map[string]interface{}{
										"schema_version": 1,
									},
									Attributes: map[string]string{
										"%":  "7",
										"id": "123999",

										"bool_field_2": "true",
										"string_field": "hello world",
										"bool_field_1": "false",
										"empty_string": "",
										"number_field": "42",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"resource with nested slice",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {},
    "root_module": {
      "resources": [
        {
          "address": "cloud_vpc.main",
          "mode": "managed",
          "type": "cloud_vpc",
          "name": "main",
          "provider_name": "cloud",
          "schema_version": 1,
          "values": {
          	"id": "123999",
            "list_of_lists": [
              ["one", "two", "three"],
              ["red", "green", "blue"],
              ["black", "white"]
            ],
            "list_of_maps": [
              {
                "action": "allow",
                "port": 443,
                "allow_bool": true
              },
              {
                "action": "deny",
                "port": 80,
                "allow_bool": false
              }
            ]
          }
        }
      ]
    }
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"cloud_vpc.main": {
								Type:     "cloud_vpc",
								Provider: "cloud",
								Primary: &terraform.InstanceState{
									ID: "123999",
									Meta: map[string]interface{}{
										"schema_version": 1,
									},
									Attributes: map[string]string{
										"%":  "3",
										"id": "123999",

										"list_of_lists.#":   "3",
										"list_of_lists.0.#": "3",
										"list_of_lists.0.0": "one",
										"list_of_lists.0.1": "two",
										"list_of_lists.0.2": "three",
										"list_of_lists.1.#": "3",
										"list_of_lists.1.0": "red",
										"list_of_lists.1.1": "green",
										"list_of_lists.1.2": "blue",
										"list_of_lists.2.#": "2",
										"list_of_lists.2.0": "black",
										"list_of_lists.2.1": "white",

										"list_of_maps.#":            "2",
										"list_of_maps.0.%":          "3",
										"list_of_maps.0.action":     "allow",
										"list_of_maps.0.allow_bool": "true",
										"list_of_maps.0.port":       "443",
										"list_of_maps.1.%":          "3",
										"list_of_maps.1.action":     "deny",
										"list_of_maps.1.allow_bool": "false",
										"list_of_maps.1.port":       "80",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"resource with nested map",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {},
    "root_module": {
      "resources": [
        {
          "address": "cloud_vpc.main",
          "mode": "managed",
          "type": "cloud_vpc",
          "name": "main",
          "provider_name": "cloud",
          "schema_version": 1,
          "values": {
          	"id": "123999",
            "map_of_maps": {
              "parent": {
                "inner": "value"
              },
              "second": {
                "inner2": "value2"
              }
            },
            "map_of_lists": {
              "parent": {
                "inner": ["one", "two"]
              },
              "second": {
                "inner2": [1, 4, 9]
              }
            },
            "map_of_list_of_maps": {
              "parent": [
                {
                  "action": "allow",
                  "port": 443,
                  "allow_bool": true
                },
                {
                  "action": "deny",
                  "port": 80,
                  "allow_bool": false
                }
              ]
            }
          }
        }
      ]
    }
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"cloud_vpc.main": {
								Type:     "cloud_vpc",
								Provider: "cloud",
								Primary: &terraform.InstanceState{
									ID: "123999",
									Meta: map[string]interface{}{
										"schema_version": 1,
									},
									Attributes: map[string]string{
										"%":  "4",
										"id": "123999",

										"map_of_maps.%":             "2",
										"map_of_maps.parent.%":      "1",
										"map_of_maps.parent.inner":  "value",
										"map_of_maps.second.%":      "1",
										"map_of_maps.second.inner2": "value2",

										"map_of_lists.%":               "2",
										"map_of_lists.parent.%":        "1",
										"map_of_lists.parent.inner.#":  "2",
										"map_of_lists.parent.inner.0":  "one",
										"map_of_lists.parent.inner.1":  "two",
										"map_of_lists.second.%":        "1",
										"map_of_lists.second.inner2.#": "3",
										"map_of_lists.second.inner2.0": "1",
										"map_of_lists.second.inner2.1": "4",
										"map_of_lists.second.inner2.2": "9",

										"map_of_list_of_maps.%":                   "1",
										"map_of_list_of_maps.parent.#":            "2",
										"map_of_list_of_maps.parent.0.%":          "3",
										"map_of_list_of_maps.parent.0.action":     "allow",
										"map_of_list_of_maps.parent.0.allow_bool": "true",
										"map_of_list_of_maps.parent.0.port":       "443",
										"map_of_list_of_maps.parent.1.%":          "3",
										"map_of_list_of_maps.parent.1.action":     "deny",
										"map_of_list_of_maps.parent.1.allow_bool": "false",
										"map_of_list_of_maps.parent.1.port":       "80",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"data source",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {},
    "root_module": {
      "resources": [
        {
          "address": "data.cloud_vpc.main",
          "mode": "data",
          "type": "cloud_vpc",
          "name": "main",
          "provider_name": "cloud",
          "schema_version": 0,
          "values": {
          	"id": "123999",
            "string_field": "hello world",
            "bool_field_1": false,
            "bool_field_2": true,
            "null_field": null,
            "empty_string": "",
            "number_field": 42
          }
        }
      ]
    }
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"data.cloud_vpc.main": {
								Type:     "cloud_vpc",
								Provider: "cloud",
								Primary: &terraform.InstanceState{
									ID: "123999",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "7",
										"id": "123999",

										"bool_field_2": "true",
										"string_field": "hello world",
										"bool_field_1": "false",
										"empty_string": "",
										"number_field": "42",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"resource with complex attributes",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {},
    "root_module": {
      "resources": [
        {
          "address": "cloud_vpc.main",
          "mode": "managed",
          "type": "cloud_vpc",
          "name": "main",
          "provider_name": "cloud",
          "schema_version": 1,
          "values": {
          	"id": "123999",
            "map_field": {
              "key": "val",
              "foo": "bar"
            },
            "list_of_string": ["first", "second"],
            "list_of_numbers": [1,2,3,4],
            "list_of_bool": [true, false, true]
          }
        }
      ]
    }
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"cloud_vpc.main": {
								Type:     "cloud_vpc",
								Provider: "cloud",
								Primary: &terraform.InstanceState{
									ID: "123999",
									Meta: map[string]interface{}{
										"schema_version": 1,
									},
									Attributes: map[string]string{
										"%":  "5",
										"id": "123999",

										"map_field.%":   "2",
										"map_field.key": "val",
										"map_field.foo": "bar",

										"list_of_string.#": "2",
										"list_of_string.0": "first",
										"list_of_string.1": "second",

										"list_of_numbers.#": "4",
										"list_of_numbers.0": "1",
										"list_of_numbers.1": "2",
										"list_of_numbers.2": "3",
										"list_of_numbers.3": "4",

										"list_of_bool.#": "3",
										"list_of_bool.0": "true",
										"list_of_bool.1": "false",
										"list_of_bool.2": "true",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"indexed resource",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {},
    "root_module": {
      "resources": [
        {
          "address": "cloud_vpc.main",
          "index": 0,
          "mode": "managed",
          "type": "cloud_vpc",
          "name": "main",
          "provider_name": "cloud",
          "schema_version": 1,
          "values": {
          	"id": "11111"
          }
        },
        {
          "address": "cloud_vpc.main",
          "index": 1,
          "mode": "managed",
          "type": "cloud_vpc",
          "name": "main",
          "provider_name": "cloud",
          "schema_version": 1,
          "values": {
          	"id": "22222"
          }
        }
      ]
    }
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"cloud_vpc.main.0": {
								Type:     "cloud_vpc",
								Provider: "cloud",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 1,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "11111",
									},
								},
							},
							"cloud_vpc.main.1": {
								Type:     "cloud_vpc",
								Provider: "cloud",
								Primary: &terraform.InstanceState{
									ID: "22222",
									Meta: map[string]interface{}{
										"schema_version": 1,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "22222",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"indexed data source",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {},
    "root_module": {
      "resources": [
        {
          "address": "data.cloud_vpc.main",
          "index": 0,
          "mode": "data",
          "type": "cloud_vpc",
          "name": "main",
          "provider_name": "cloud",
          "schema_version": 0,
          "values": {
          	"id": "11111"
          }
        },
        {
          "address": "data.cloud_vpc.main",
          "index": 1,
          "mode": "data",
          "type": "cloud_vpc",
          "name": "main",
          "provider_name": "cloud",
          "schema_version": 0,
          "values": {
          	"id": "22222"
          }
        }
      ]
    }
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"data.cloud_vpc.main.0": {
								Type:     "cloud_vpc",
								Provider: "cloud",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "11111",
									},
								},
							},
							"data.cloud_vpc.main.1": {
								Type:     "cloud_vpc",
								Provider: "cloud",
								Primary: &terraform.InstanceState{
									ID: "22222",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "22222",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"for_each",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {},
    "root_module": {
      "resources": [
        {
          "address": "cloud_vpc.main",
          "index": "one",
          "mode": "managed",
          "type": "cloud_vpc",
          "name": "main",
          "provider_name": "cloud",
          "schema_version": 1,
          "values": {
          	"id": "11111"
          }
        },
        {
          "address": "cloud_vpc.main",
          "index": "two",
          "mode": "managed",
          "type": "cloud_vpc",
          "name": "main",
          "provider_name": "cloud",
          "schema_version": 1,
          "values": {
          	"id": "22222"
          }
        }
      ]
    }
  }
}`,
			nil,
			&expectedError{Prefix: "unexpected index type (string)"},
		},
		{
			"depends on",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "outputs": {},
    "root_module": {
      "resources": [
        {
          "address": "cloud_vpc.primary",
          "mode": "managed",
          "type": "cloud_vpc",
          "name": "primary",
          "provider_name": "cloud",
          "values": {
          	"id": "11111"
          }
        },
        {
          "address": "cloud_vpc.secondary",
          "mode": "managed",
          "type": "cloud_vpc",
          "name": "secondary",
          "provider_name": "cloud",
          "values": {
          	"id": "22222"
          },
          "depends_on": [
            "cloud_vpc.primary"
          ]
        }
      ]
    }
  }
}`,
			&terraform.State{
				Version:   3,
				TFVersion: "0.12.18",
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"cloud_vpc.primary": {
								Type:     "cloud_vpc",
								Provider: "cloud",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "11111",
									},
								},
							},
							"cloud_vpc.secondary": {
								Type:     "cloud_vpc",
								Provider: "cloud",
								Primary: &terraform.InstanceState{
									ID: "22222",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "22222",
									},
								},
								Dependencies: []string{
									"cloud_vpc.primary",
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			nil,
		},
		{
			"child modules",
			`{
  "format_version": "0.1",
  "terraform_version": "0.12.18",
  "values": {
    "root_module": {
      "child_modules": [
        {
          "resources": [
            {
              "address": "cloud_vpc.primary",
              "mode": "managed",
              "type": "cloud_vpc",
              "name": "primary",
              "provider_name": "cloud",
              "values": {
                "id": "11111"
              }
            }
          ]
        }
      ]
    }
  }
}`,
			nil,
			&expectedError{Prefix: "Modules are not supported."},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.Name), func(t *testing.T) {
			var rawState tfjson.State
			rawState.UseJSONNumber(true)

			err := unmarshalJSON([]byte(tc.RawState), &rawState)
			if err != nil {
				t.Fatal(err)
			}

			shimmedState, err := shimStateFromJson(&rawState)
			if tc.ExpectedErr != nil {
				if err == nil {
					t.Fatalf("Expected error with prefix: %q\nGot no error.",
						tc.ExpectedErr.Prefix)
				}
				if strings.HasPrefix(err.Error(), tc.ExpectedErr.Prefix) {
					return
				}
				t.Fatalf("Error mismatch.\nExpected prefix: %q\nGot: %q\n",
					tc.ExpectedErr.Prefix, err.Error())
			}
			if err != nil {
				t.Fatal(err)
			}

			// Lineage is randomly generated, so we wipe it to make comparing easier
			shimmedState.Lineage = ""

			if diff := cmp.Diff(tc.ExpectedState, shimmedState); diff != "" {
				t.Fatalf("state mismatch:\n%s", diff)
			}
		})
	}
}
