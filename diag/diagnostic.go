package diag

import (
	"fmt"

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

func (d Diagnostic) Validate() error {
	var validSev bool
	for _, sev := range severities {
		if d.Severity == sev {
			validSev = true
			break
		}
	}
	if !validSev {
		return fmt.Errorf("invalid severity: %v", d.Severity)
	}
	if d.Summary == "" {
		return fmt.Errorf("empty detail")
	}
	return nil
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

var severities = []Severity{Error, Warning}
