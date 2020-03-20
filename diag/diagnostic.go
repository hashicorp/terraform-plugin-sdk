package diag

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

type Diagnostics []*Diagnostic

func (diags Diagnostics) HasErrors() bool {
	for _, d := range diags {
		if d.Severity == Error {
			return true
		}
	}
	return false
}

func (diags Diagnostics) Errors() []error {
	var errs []error
	for _, d := range diags {
		if d.Severity == Error {
			errs = append(errs, error(d))
		}
	}
	return errs
}

func (diags Diagnostics) Warnings() []string {
	var warns []string
	for _, d := range diags {
		if d.Severity == Warning {
			warns = append(warns, d.Error())
		}
	}
	return warns
}

type Diagnostic struct {
	Severity      Severity
	Summary       string
	Detail        string
	AttributePath cty.Path
}

func (d *Diagnostic) Error() string {
	return fmt.Sprintf("%s: %s", d.Severity, d.Summary)
}

type Severity int

//go:generate go run golang.org/x/tools/cmd/stringer -type=Severity

const (
	Error Severity = iota
	Warning
)
