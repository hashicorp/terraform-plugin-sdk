package tfdiags

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestDiagnosticsForRPC(t *testing.T) {
	var diags Diagnostics
	diags = diags.Append(fmt.Errorf("bad"))
	diags = diags.Append(SimpleWarning("less bad"))

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	rpcDiags := diags.ForRPC()
	err := enc.Encode(rpcDiags)
	if err != nil {
		t.Fatalf("error on Encode: %s", err)
	}

	var got Diagnostics
	err = dec.Decode(&got)
	if err != nil {
		t.Fatalf("error on Decode: %s", err)
	}

	want := Diagnostics{
		&rpcFriendlyDiag{
			Severity_: Error,
			Summary_:  "bad",
		},
		&rpcFriendlyDiag{
			Severity_: Warning,
			Summary_:  "less bad",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("wrong result\ngot: %swant: %s", spew.Sdump(got), spew.Sdump(want))
	}
}
