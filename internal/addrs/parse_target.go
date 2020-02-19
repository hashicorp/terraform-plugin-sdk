package addrs

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"

	"github.com/hashicorp/terraform-plugin-sdk/internal/tfdiags"
)

// target describes a targeted address with source location information.
type target struct {
	Subject     targetableI
	SourceRange tfdiags.SourceRange
}

// parseTarget attempts to interpret the given traversal as a targetable
// address. The given traversal must be absolute, or this function will
// panic.
//
// If no error diagnostics are returned, the returned target includes the
// address that was extracted and the source range it was extracted from.
//
// If error diagnostics are returned then the target value is invalid and
// must not be used.
func parseTarget(traversal hcl.Traversal) (*target, tfdiags.Diagnostics) {
	path, remain, diags := parseModuleInstancePrefix(traversal)
	if diags.HasErrors() {
		return nil, diags
	}

	rng := tfdiags.SourceRangeFromHCL(traversal.SourceRange())

	if len(remain) == 0 {
		return &target{
			Subject:     path,
			SourceRange: rng,
		}, diags
	}

	mode := ManagedResourceMode
	if remain.RootName() == "data" {
		mode = DataResourceMode
		remain = remain[1:]
	}

	if len(remain) < 2 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid address",
			Detail:   "Resource specification must include a resource type and name.",
			Subject:  remain.SourceRange().Ptr(),
		})
		return nil, diags
	}

	var typeName, name string
	switch tt := remain[0].(type) {
	case hcl.TraverseRoot:
		typeName = tt.Name
	case hcl.TraverseAttr:
		typeName = tt.Name
	default:
		switch mode {
		case ManagedResourceMode:
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid address",
				Detail:   "A resource type name is required.",
				Subject:  remain[0].SourceRange().Ptr(),
			})
		case DataResourceMode:
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid address",
				Detail:   "A data source name is required.",
				Subject:  remain[0].SourceRange().Ptr(),
			})
		default:
			panic("unknown mode")
		}
		return nil, diags
	}

	switch tt := remain[1].(type) {
	case hcl.TraverseAttr:
		name = tt.Name
	default:
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid address",
			Detail:   "A resource name is required.",
			Subject:  remain[1].SourceRange().Ptr(),
		})
		return nil, diags
	}

	var subject targetableI
	remain = remain[2:]
	switch len(remain) {
	case 0:
		subject = path.Resource(mode, typeName, name)
	case 1:
		if tt, ok := remain[0].(hcl.TraverseIndex); ok {
			key, err := parseInstanceKey(tt.Key)
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid address",
					Detail:   fmt.Sprintf("Invalid resource instance key: %s.", err),
					Subject:  remain[0].SourceRange().Ptr(),
				})
				return nil, diags
			}

			subject = path.ResourceInstance(mode, typeName, name, key)
		} else {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid address",
				Detail:   "Resource instance key must be given in square brackets.",
				Subject:  remain[0].SourceRange().Ptr(),
			})
			return nil, diags
		}
	default:
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid address",
			Detail:   "Unexpected extra operators after address.",
			Subject:  remain[1].SourceRange().Ptr(),
		})
		return nil, diags
	}

	return &target{
		Subject:     subject,
		SourceRange: rng,
	}, diags
}
