// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// PreferWriteOnlyAttribute is a ValidateResourceConfigFunc that returns a warning
// if the Terraform client supports write-only attributes and the old attribute
// has a value instead of the write-only attribute.
func PreferWriteOnlyAttribute(oldAttribute cty.Path, writeOnlyAttribute cty.Path) schema.ValidateResourceConfigFunc {
	return func(ctx context.Context, req schema.ValidateResourceConfigRequest, resp *schema.ValidateResourceConfigResponse) {
		if !req.WriteOnlyAttributesAllowed {
			return
		}

		// Apply all but the last step to retrieve the attribute name
		// for any diags that we return.
		oldLastStepVal, oldLastStep, err := oldAttribute.LastStep(req.RawConfig)
		if err != nil {
			resp.Diagnostics = diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Invalid oldAttributePath",
					Detail: fmt.Sprintf("Encountered an error when applying the specified oldAttribute path, "+
						"original error: %s", err),
					AttributePath: oldAttribute,
				},
			}
			return
		}

		// Only attribute steps have a Name field
		oldAttributeStep, ok := oldLastStep.(cty.GetAttrStep)
		if !ok {
			resp.Diagnostics = diag.Diagnostics{
				{
					Severity:      diag.Error,
					Summary:       "Invalid oldAttributePath",
					Detail:        "The specified oldAttribute path must point to an attribute",
					AttributePath: oldAttribute,
				},
			}
			return
		}

		oldAttributeConfigVal, err := oldAttributeStep.Apply(oldLastStepVal)
		if err != nil {
			resp.Diagnostics = diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Invalid oldAttributePath",
					Detail: fmt.Sprintf("Encountered an error when applying the specified oldAttribute path, "+
						"original error: %s", err),
					AttributePath: oldAttribute,
				},
			}
			return
		}

		writeOnlyLastStepVal, writeOnlyLastStep, err := writeOnlyAttribute.LastStep(req.RawConfig)
		if err != nil {
			resp.Diagnostics = diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Invalid writeOnlyAttributePath",
					Detail: fmt.Sprintf("Encountered an error when applying the specified writeOnlyAttribute path, "+
						"original error: %s", err),
					AttributePath: writeOnlyAttribute,
				},
			}
			return
		}

		// Only attribute steps have a Name field
		writeOnlyAttributeStep, ok := writeOnlyLastStep.(cty.GetAttrStep)
		if !ok {
			resp.Diagnostics = diag.Diagnostics{
				{
					Severity:      diag.Error,
					Summary:       "Invalid writeOnlyAttributePath",
					Detail:        "The specified writeOnlyAttribute path must point to an attribute",
					AttributePath: writeOnlyAttribute,
				},
			}
			return
		}

		writeOnlyAttributeConfigVal, err := writeOnlyAttributeStep.Apply(writeOnlyLastStepVal)
		if err != nil {
			resp.Diagnostics = diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Invalid writeOnlyAttributePath",
					Detail: fmt.Sprintf("Encountered an error when applying the specified writeOnlyAttribute path, "+
						"original error: %s", err),
					AttributePath: writeOnlyAttribute,
				},
			}
			return
		}

		//oldAttributeConfigVal, err := cty.Transform(req.RawConfig, func(path cty.Path, val cty.Value) (cty.Value, error) {
		//	if path.Equals(oldAttribute) {
		//		oldAttributeConfig := req.RawConfig.GetAttr(oldAttributeName)
		//		println(oldAttributeConfig.IsKnown())
		//		return val, nil
		//	}
		//
		//	// nothing to do if we already have a value
		//	if !val.IsNull() {
		//		return val, nil
		//	}
		//
		//	return val, nil
		//})
		//// We shouldn't encounter any errors here, but handling them just in case.
		//if err != nil {
		//	resp.Diagnostics = diag.FromErr(err)
		//	return
		//}

		if !oldAttributeConfigVal.IsNull() && writeOnlyAttributeConfigVal.IsNull() {
			resp.Diagnostics = diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: fmt.Sprintf("The attribute %s has a WriteOnly version %s available. "+
						"Use the WriteOnly version of the attribute when possible.", oldAttributeStep.Name, writeOnlyAttributeStep.Name),
					AttributePath: oldAttribute,
				},
			}
		}
	}
}
