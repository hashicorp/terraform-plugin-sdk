package httpclient

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/meta"
)

func TestUserAgentAppendViaEnvVar(t *testing.T) {
	if oldenv, isSet := os.LookupEnv(uaEnvVar); isSet {
		defer os.Setenv(uaEnvVar, oldenv)
	} else {
		defer os.Unsetenv(uaEnvVar)
	}

	expectedBase := "HashiCorp Terraform/0.0.0 (+https://www.terraform.io) Terraform Plugin SDK/" + meta.SDKVersionString()

	testCases := []struct {
		envVarValue string
		expected    string
	}{
		{"", expectedBase},
		{" ", expectedBase},
		{" \n", expectedBase},
		{"test/1", expectedBase + " test/1"},
		{"test/1 (comment)", expectedBase + " test/1 (comment)"},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			os.Unsetenv(uaEnvVar)
			os.Setenv(uaEnvVar, tc.envVarValue)
			givenUA := TerraformUserAgent("0.0.0")
			if givenUA != tc.expected {
				t.Fatalf("Expected User-Agent '%s' does not match '%s'", tc.expected, givenUA)
			}
		})
	}
}
