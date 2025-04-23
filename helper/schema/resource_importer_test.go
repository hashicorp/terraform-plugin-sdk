// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestInternalValidate(t *testing.T) {
	r := &ResourceImporter{
		State:        ImportStatePassthrough,
		StateContext: ImportStatePassthroughContext,
	}
	if err := r.InternalValidate(); err == nil {
		t.Fatal("ResourceImporter should not allow State and StateContext to be set")
	}
}

func TestImportStatePassthroughWithIdentity(t *testing.T) {
	// shared among all tests, defined once to keep them shorter
	identitySchema := map[string]*Schema{
		"email": {
			Type:              TypeString,
			RequiredForImport: true,
		},
		"region": {
			Type:              TypeString,
			OptionalForImport: true,
		},
	}

	tests := []struct {
		name                 string
		idAttributePath      string
		resourceData         *ResourceData
		expectedResourceData *ResourceData
		expectedError        string
	}{
		{
			name:            "import from id just sets id",
			idAttributePath: "email",
			resourceData: &ResourceData{
				identitySchema: identitySchema,
				state: &terraform.InstanceState{
					ID: "hello@example.internal",
				},
			},
			expectedResourceData: &ResourceData{
				identitySchema: identitySchema,
				state: &terraform.InstanceState{
					ID: "hello@example.internal",
				},
			},
		},
		{
			name:            "import from identity sets id and identity",
			idAttributePath: "email",
			resourceData: &ResourceData{
				identitySchema: identitySchema,
				state: &terraform.InstanceState{
					Identity: map[string]string{
						"email": "hello@example.internal",
					},
				},
			},
			expectedResourceData: &ResourceData{
				identitySchema: identitySchema,
				state: &terraform.InstanceState{
					ID: "hello@example.internal",
				},
				newIdentity: &IdentityData{
					schema: identitySchema,
					raw: map[string]string{
						"email": "hello@example.internal",
					},
				},
			},
		},
		{
			name:            "import from identity sets id and identity (with region set)",
			idAttributePath: "email",
			resourceData: &ResourceData{
				identitySchema: identitySchema,
				state: &terraform.InstanceState{
					Identity: map[string]string{
						"email":  "hello@example.internal",
						"region": "eu-west-1",
					},
				},
			},
			expectedResourceData: &ResourceData{
				identitySchema: identitySchema,
				state: &terraform.InstanceState{
					ID: "hello@example.internal",
				},
				newIdentity: &IdentityData{
					schema: identitySchema,
					raw: map[string]string{
						"email":  "hello@example.internal",
						"region": "eu-west-1",
					},
				},
			},
		},
		{
			name:            "import from identity fails without required field",
			idAttributePath: "email",
			resourceData: &ResourceData{
				identitySchema: identitySchema,
				state: &terraform.InstanceState{
					Identity: map[string]string{
						"region": "eu-west-1",
					},
				},
			},
			expectedError: "expected identity to contain key email",
		},
		{
			name:            "import from identity fails if attribute is not a string",
			idAttributePath: "number",
			resourceData: &ResourceData{
				identitySchema: map[string]*Schema{
					"number": {
						Type:              TypeInt,
						RequiredForImport: true,
					},
				},
				state: &terraform.InstanceState{
					Identity: map[string]string{
						"number": "1",
					},
				},
			},
			expectedError: "expected identity key number to be a string, was: int",
		},
		{
			name:            "import from identity fails without schema",
			idAttributePath: "email",
			resourceData: &ResourceData{
				state: &terraform.InstanceState{
					Identity: map[string]string{
						"email": "hello@example.internal",
					},
				},
			},
			expectedError: "error getting identity: Resource does not have Identity schema. Please set one in order to use Identity(). This is always a problem in the provider code.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results, err := ImportStatePassthroughWithIdentity(test.idAttributePath)(nil, test.resourceData, nil)
			if err != nil {
				if test.expectedError == "" {
					t.Fatalf("unexpected error: %s", err)
				}
				if err.Error() != test.expectedError {
					t.Fatalf("expected error: %s, got: %s", test.expectedError, err)
				}
				return // we don't expect any results if there is an error
			}
			if len(results) != 1 {
				t.Fatalf("expected 1 result, got: %d", len(results))
			}
			// compare id and identity in resource data
			if results[0].Id() != test.expectedResourceData.Id() {
				t.Fatalf("expected id: %s, got: %s", test.expectedResourceData.Id(), results[0].Id())
			}
			// compare identity
			expectedIdentity, err := test.expectedResourceData.Identity()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			resultIdentity, err := results[0].Identity()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// check whether all result identity attributes exist as expected
			for key := range expectedIdentity.schema {
				expected := expectedIdentity.getRaw(key)
				if expected.Exists {
					result := resultIdentity.getRaw(key)
					if !result.Exists {
						t.Fatalf("expected identity attribute %s to exist", key)
					}
					if expected.Value != result.Value {
						t.Fatalf("expected identity attribute %s to be %s, got: %s", key, expected.Value, result.Value)
					}
				}
			}
			// check whether there are no additional attributes in the result identity
			for key := range resultIdentity.schema {
				if _, ok := expectedIdentity.schema[key]; !ok {
					t.Fatalf("unexpected identity attribute %s", key)
				}
			}
		})
	}
}
