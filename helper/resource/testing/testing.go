// Package testing aims to expose interfaces that mirror go's testing package
// this prevents go's testing package from being imported or exposed to
// importers of the SDK
package testing

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
