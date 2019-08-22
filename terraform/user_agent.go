package terraform

import (
	"github.com/hashicorp/terraform-plugin-sdk/httpclient"
)

// Generate a UserAgent string
//
// Deprecated: Use httpclient.UserAgent(version) instead
func UserAgentString() string {
	return httpclient.UserAgentString()
}
