package diag

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/zclconf/go-cty/cty"
)

type Diagnostics []*Diagnostic

func (diags Diagnostics) Append(i interface{}) Diagnostics {
	switch v := i.(type) {
	case Diagnostics:
		if len(v) != 0 {
			diags = append(diags, v...)
		}
	case *Diagnostic:
		diags = append(diags, v)
	case error:
		if v != nil {
			diags = append(diags, &Diagnostic{
				Severity: Error,
				Summary:  v.Error(),
			})
		}
	}
	return diags
}

func (diags Diagnostics) Err() error {
	return errors.New(multierror.ListFormatFunc(diags.Errors()))
}

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
	s := fmt.Sprintf("%s: %s", d.Severity, d.Summary)
	if d.Detail != "" {
		s = fmt.Sprintf("%s: %s", s, d.Detail)
	}
	return s
}

type Severity int

//go:generate go run golang.org/x/tools/cmd/stringer -type=Severity

const (
	Error Severity = iota
	Warning
)
