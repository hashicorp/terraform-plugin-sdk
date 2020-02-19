package addrs

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/hashicorp/terraform-plugin-sdk/internal/tfdiags"
)

// ProviderConfig is the address of a provider configuration.
type ProviderConfig struct {
	Type string

	// If not empty, Alias identifies which non-default (aliased) provider
	// configuration this address refers to.
	Alias string
}

// parseProviderConfigCompact parses the given absolute traversal as a relative
// provider address in compact form. The following are examples of traversals
// that can be successfully parsed as compact relative provider configuration
// addresses:
//
//     aws
//     aws.foo
//
// This function will panic if given a relative traversal.
//
// If the returned diagnostics contains errors then the result value is invalid
// and must not be used.
func parseProviderConfigCompact(traversal hcl.Traversal) (ProviderConfig, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics
	ret := ProviderConfig{
		Type: traversal.RootName(),
	}

	if len(traversal) < 2 {
		// Just a type name, then.
		return ret, diags
	}

	aliasStep := traversal[1]
	switch ts := aliasStep.(type) {
	case hcl.TraverseAttr:
		ret.Alias = ts.Name
		return ret, diags
	default:
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid provider configuration address",
			Detail:   "The provider type name must either stand alone or be followed by an alias name separated with a dot.",
			Subject:  aliasStep.SourceRange().Ptr(),
		})
	}

	if len(traversal) > 2 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid provider configuration address",
			Detail:   "Extraneous extra operators after provider configuration address.",
			Subject:  traversal[2:].SourceRange().Ptr(),
		})
	}

	return ret, diags
}

func (pc ProviderConfig) String() string {
	if pc.Type == "" {
		// Should never happen; always indicates a bug
		return "provider.<invalid>"
	}

	if pc.Alias != "" {
		return fmt.Sprintf("provider.%s.%s", pc.Type, pc.Alias)
	}

	return "provider." + pc.Type
}

// absProviderConfig is the absolute address of a provider configuration
// within a particular module instance.
type absProviderConfig struct {
	Module         ModuleInstance
	ProviderConfig ProviderConfig
}

// parseAbsProviderConfig parses the given traversal as an absolute provider
// address. The following are examples of traversals that can be successfully
// parsed as absolute provider configuration addresses:
//
//     provider.aws
//     provider.aws.foo
//     module.bar.provider.aws
//     module.bar.module.baz.provider.aws.foo
//     module.foo[1].provider.aws.foo
//
// This type of address is used, for example, to record the relationships
// between resources and provider configurations in the state structure.
// This type of address is not generally used in the UI, except in error
// messages that refer to provider configurations.
func parseAbsProviderConfig(traversal hcl.Traversal) (absProviderConfig, tfdiags.Diagnostics) {
	modInst, remain, diags := parseModuleInstancePrefix(traversal)
	ret := absProviderConfig{
		Module: modInst,
	}
	if len(remain) < 2 || remain.RootName() != "provider" {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid provider configuration address",
			Detail:   "Provider address must begin with \"provider.\", followed by a provider type name.",
			Subject:  remain.SourceRange().Ptr(),
		})
		return ret, diags
	}
	if len(remain) > 3 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid provider configuration address",
			Detail:   "Extraneous operators after provider configuration alias.",
			Subject:  hcl.Traversal(remain[3:]).SourceRange().Ptr(),
		})
		return ret, diags
	}

	if tt, ok := remain[1].(hcl.TraverseAttr); ok {
		ret.ProviderConfig.Type = tt.Name
	} else {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid provider configuration address",
			Detail:   "The prefix \"provider.\" must be followed by a provider type name.",
			Subject:  remain[1].SourceRange().Ptr(),
		})
		return ret, diags
	}

	if len(remain) == 3 {
		if tt, ok := remain[2].(hcl.TraverseAttr); ok {
			ret.ProviderConfig.Alias = tt.Name
		} else {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid provider configuration address",
				Detail:   "Provider type name must be followed by a configuration alias name.",
				Subject:  remain[2].SourceRange().Ptr(),
			})
			return ret, diags
		}
	}

	return ret, diags
}

// ParseAbsProviderConfigStr is a helper wrapper around ParseAbsProviderConfig
// that takes a string and parses it with the HCL native syntax traversal parser
// before interpreting it.
//
// This should be used only in specialized situations since it will cause the
// created references to not have any meaningful source location information.
// If a reference string is coming from a source that should be identified in
// error messages then the caller should instead parse it directly using a
// suitable function from the HCL API and pass the traversal itself to
// ParseAbsProviderConfig.
//
// Error diagnostics are returned if either the parsing fails or the analysis
// of the traversal fails. There is no way for the caller to distinguish the
// two kinds of diagnostics programmatically. If error diagnostics are returned
// the returned address is invalid.
func ParseAbsProviderConfigStr(str string) (absProviderConfig, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics

	traversal, parseDiags := hclsyntax.ParseTraversalAbs([]byte(str), "", hcl.Pos{Line: 1, Column: 1})
	diags = diags.Append(parseDiags)
	if parseDiags.HasErrors() {
		return absProviderConfig{}, diags
	}

	addr, addrDiags := parseAbsProviderConfig(traversal)
	diags = diags.Append(addrDiags)
	return addr, diags
}

func (pc absProviderConfig) String() string {
	if len(pc.Module) == 0 {
		return pc.ProviderConfig.String()
	}
	return fmt.Sprintf("%s.%s", pc.Module.String(), pc.ProviderConfig.String())
}
