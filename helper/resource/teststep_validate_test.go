// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestTestStepHasProviders(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testStep TestStep
		expected bool
	}{
		"none": {
			testStep: TestStep{},
			expected: false,
		},
		"externalproviders": {
			testStep: TestStep{
				ExternalProviders: map[string]ExternalProvider{
					"test": {}, // does not need to be real
				},
			},
			expected: true,
		},
		"protov5providerfactories": {
			testStep: TestStep{
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"test": nil, // does not need to be real
				},
			},
			expected: true,
		},
		"protov6providerfactories": {
			testStep: TestStep{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"test": nil, // does not need to be real
				},
			},
			expected: true,
		},
		"providerfactories": {
			testStep: TestStep{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"test": nil, // does not need to be real
				},
			},
			expected: true,
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.testStep.hasProviders(context.Background())

			if got != test.expected {
				t.Errorf("expected %t, got %t", test.expected, got)
			}
		})
	}
}

func TestTestStepValidate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testStep                TestStep
		testStepValidateRequest testStepValidateRequest
		expectedError           error
	}{
		"config-and-importstate-and-refreshstate-missing": {
			testStep:                TestStep{},
			testStepValidateRequest: testStepValidateRequest{},
			expectedError:           fmt.Errorf("TestStep missing Config or ImportState or RefreshState"),
		},
		"config-and-refreshstate-both-set": {
			testStep: TestStep{
				Config:       "# not empty",
				RefreshState: true,
			},
			expectedError: fmt.Errorf("TestStep cannot have Config and RefreshState"),
		},
		"refreshstate-first-step": {
			testStep: TestStep{
				RefreshState: true,
			},
			testStepValidateRequest: testStepValidateRequest{
				StepNumber: 1,
			},
			expectedError: fmt.Errorf("TestStep cannot have RefreshState as first step"),
		},
		"importstate-and-refreshstate-both-true": {
			testStep: TestStep{
				ImportState:  true,
				RefreshState: true,
			},
			testStepValidateRequest: testStepValidateRequest{},
			expectedError:           fmt.Errorf("TestStep cannot have ImportState and RefreshState in same step"),
		},
		"destroy-and-refreshstate-both-true": {
			testStep: TestStep{
				Destroy:      true,
				RefreshState: true,
			},
			testStepValidateRequest: testStepValidateRequest{},
			expectedError:           fmt.Errorf("TestStep cannot have RefreshState and Destroy"),
		},
		"externalproviders-overlapping-providerfactories": {
			testStep: TestStep{
				Config: "# not empty",
				ExternalProviders: map[string]ExternalProvider{
					"test": {}, // does not need to be real
				},
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"test": nil, // does not need to be real
				},
			},
			testStepValidateRequest: testStepValidateRequest{},
			expectedError:           fmt.Errorf("TestStep provider \"test\" set in both ExternalProviders and ProviderFactories"),
		},
		"externalproviders-testcase-providers": {
			testStep: TestStep{
				Config: "# not empty",
				ExternalProviders: map[string]ExternalProvider{
					"test": {}, // does not need to be real
				},
			},
			testStepValidateRequest: testStepValidateRequest{
				TestCaseHasProviders: true,
			},
			expectedError: fmt.Errorf("Providers must only be specified either at the TestCase or TestStep level"),
		},
		"importstate-missing-resourcename": {
			testStep: TestStep{
				ImportState: true,
			},
			testStepValidateRequest: testStepValidateRequest{
				TestCaseHasProviders: true,
			},
			expectedError: fmt.Errorf("TestStep ImportState must be specified with ImportStateId, ImportStateIdFunc, or ResourceName"),
		},
		"protov5providerfactories-testcase-providers": {
			testStep: TestStep{
				Config: "# not empty",
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"test": nil, // does not need to be real
				},
			},
			testStepValidateRequest: testStepValidateRequest{
				TestCaseHasProviders: true,
			},
			expectedError: fmt.Errorf("Providers must only be specified either at the TestCase or TestStep level"),
		},
		"protov6providerfactories-testcase-providers": {
			testStep: TestStep{
				Config: "# not empty",
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"test": nil, // does not need to be real
				},
			},
			testStepValidateRequest: testStepValidateRequest{
				TestCaseHasProviders: true,
			},
			expectedError: fmt.Errorf("Providers must only be specified either at the TestCase or TestStep level"),
		},
		"providerfactories-testcase-providers": {
			testStep: TestStep{
				Config: "# not empty",
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"test": nil, // does not need to be real
				},
			},
			testStepValidateRequest: testStepValidateRequest{
				TestCaseHasProviders: true,
			},
			expectedError: fmt.Errorf("Providers must only be specified either at the TestCase or TestStep level"),
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := test.testStep.validate(context.Background(), test.testStepValidateRequest)

			if err != nil {
				if test.expectedError == nil {
					t.Fatalf("unexpected error: %s", err)
				}

				if !strings.Contains(err.Error(), test.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", test.expectedError, err)
				}
			}

			if err == nil && test.expectedError != nil {
				t.Errorf("expected error: %s", test.expectedError)
			}
		})
	}
}
