package diag

import (
	"github.com/zclconf/go-cty/cty"
)

type Diagnostics []Diagnostic

func (diags Diagnostics) HasError() bool {
	for i := range diags {
		if diags[i].Severity == Error {
			return true
		}
	}
	return false
}

type Diagnostic struct {
	Severity      Severity
	Summary       string
	Detail        string
	AttributePath cty.Path
}

func FromErr(err error) Diagnostic {
	return Diagnostic{
		Severity: Error,
		Summary:  err.Error(),
	}
}

type Severity int

const (
	Error Severity = iota
	Warning
)
