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

func TestTestCaseHasProviders(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testCase TestCase
		expected bool
	}{
		"none": {
			testCase: TestCase{},
			expected: false,
		},
		"externalproviders": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {}, // does not need to be real
				},
			},
			expected: true,
		},
		"protov5providerfactories": {
			testCase: TestCase{
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"test": nil, // does not need to be real
				},
			},
			expected: true,
		},
		"protov6providerfactories": {
			testCase: TestCase{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"test": nil, // does not need to be real
				},
			},
			expected: true,
		},
		"providers": {
			testCase: TestCase{
				Providers: map[string]*schema.Provider{
					"test": nil, // does not need to be real
				},
			},
			expected: true,
		},
		"providerfactories": {
			testCase: TestCase{
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

			got := test.testCase.hasProviders(context.Background())

			if got != test.expected {
				t.Errorf("expected %t, got %t", test.expected, got)
			}
		})
	}
}

func TestTestCaseValidate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testCase      TestCase
		expectedError error
	}{
		"valid": {
			testCase: TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"test": nil, // does not need to be real
				},
				Steps: []TestStep{
					{
						Config: "# not empty",
					},
				},
			},
		},
		"externalproviders-overlapping-providers": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {}, // does not need to be real
				},
				Providers: map[string]*schema.Provider{
					"test": nil, // does not need to be real
				},
				Steps: []TestStep{
					{
						Config: "",
					},
				},
			},
			expectedError: fmt.Errorf("TestCase provider \"test\" set in both ExternalProviders and Providers"),
		},
		"externalproviders-overlapping-providerfactories": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {}, // does not need to be real
				},
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"test": nil, // does not need to be real
				},
				Steps: []TestStep{
					{
						Config: "",
					},
				},
			},
			expectedError: fmt.Errorf("TestCase provider \"test\" set in both ExternalProviders and ProviderFactories"),
		},
		"steps-missing": {
			testCase:      TestCase{},
			expectedError: fmt.Errorf("TestCase missing Steps"),
		},
		"steps-validate-error": {
			testCase: TestCase{
				Steps: []TestStep{
					{},
				},
			},
			expectedError: fmt.Errorf("TestStep 1/1 validation error"),
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := test.testCase.validate(context.Background())

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
