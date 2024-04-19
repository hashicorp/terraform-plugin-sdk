package schema

// MAINTAINER NOTE: Only PROVIDER_CONFIG_UNKNOWN (enum value 2 in the plugin-protocol) is relevant
// for SDKv2. Since (DeferralResponse).DeferralReason is mapped directly to the plugin-protocol,
// the other enum values are intentionally omitted here.
const (
	// DeferralReasonUnknown represents an undefined deferral reason.
	DeferralReasonUnknown DeferralReason = 0

	// DeferralReasonProviderConfigUnknown represents a deferral reason caused
	// by unknown provider configuration.
	DeferralReasonProviderConfigUnknown DeferralReason = 2
)

// DeferralResponse is used to indicate to Terraform that a resource or data source is not able
// to be applied yet and should be skipped (deferred). After completing an apply that has deferred actions,
// the practitioner can then execute additional plan and apply “rounds” to eventually reach convergence
// where there are no remaining deferred actions.
type DeferralResponse struct {
	// Reason represents the deferral reason.
	Reason DeferralReason
}

// TODO: doc
type DeferralReason int32

// TODO: doc
func (d DeferralReason) String() string {
	switch d {
	case 0:
		return "Unknown"
	case 2:
		return "Provider Config Unknown"
	}
	return "Unknown"
}
