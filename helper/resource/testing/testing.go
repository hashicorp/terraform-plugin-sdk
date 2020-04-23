// Package testing aims to expose interfaces that mirror go's testing package
// this prevents go's testing package from being imported or exposed to
// importers of the SDK
package testing

import "testing"

// T is the interface used to handle the test lifecycle of a test.
//
// Users should just use a *testing.T object, which implements this.
type T interface {
	Error(args ...interface{})
	FailNow()
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
	Log(args ...interface{})
	Name() string
	Parallel()
	Skip(args ...interface{})
	SkipNow()
}

type M interface {
	Run() int
}

// Verbose just wraps the official testing package's helper of the same name.
// This is the final reference to the testing package in non *_test.go files
// in the SDK.
func Verbose() bool {
	return testing.Verbose()
}
