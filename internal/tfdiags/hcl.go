package tfdiags

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/hcl/v2"
)

// hclDiagnostic is a Diagnostic implementation that wraps a HCL Diagnostic
type hclDiagnostic struct {
	diag *hcl.Diagnostic
}

var _ Diagnostic = hclDiagnostic{}

func (d hclDiagnostic) Severity() Severity {
	switch d.diag.Severity {
	case hcl.DiagWarning:
		return Warning
	default:
		return Error
	}
}

func (d hclDiagnostic) Description() Description {
	return Description{
		Summary: d.diag.Summary,
		Detail:  d.diag.Detail,
	}
}

func (d hclDiagnostic) Source() Source {
	var ret Source
	if d.diag.Subject != nil {
		rng := SourceRangeFromHCL(*d.diag.Subject)
		ret.Subject = &rng
	}
	if d.diag.Context != nil {
		rng := SourceRangeFromHCL(*d.diag.Context)
		ret.Context = &rng
	}
	return ret
}

func (d hclDiagnostic) FromExpr() *FromExpr {
	if d.diag.Expression == nil || d.diag.EvalContext == nil {
		return nil
	}
	return &FromExpr{
		Expression:  d.diag.Expression,
		EvalContext: d.diag.EvalContext,
	}
}

// SourceRangeFromHCL constructs a SourceRange from the corresponding range
// type within the HCL package.
func SourceRangeFromHCL(hclRange hcl.Range) SourceRange {
	return SourceRange{
		Filename: hclRange.Filename,
		Start: SourcePos{
			Line:   hclRange.Start.Line,
			Column: hclRange.Start.Column,
			Byte:   hclRange.Start.Byte,
		},
		End: SourcePos{
			Line:   hclRange.End.Line,
			Column: hclRange.End.Column,
			Byte:   hclRange.End.Byte,
		},
	}
}
