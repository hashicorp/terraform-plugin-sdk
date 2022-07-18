package logging_test

import (
	"bytes"
	"context"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

func TestNewLoggingHTTPTransport(t *testing.T) {
	ctx, loggerOutput := setupRootLogger(t)

	transport := logging.NewLoggingHTTPTransport("test", http.DefaultTransport)
	client := http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	req, _ := http.NewRequest("GET", "https://www.terraform.io", nil)
	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	entries, err := tflogtest.MultilineJSONDecode(loggerOutput)
	if err != nil {
		t.Fatalf("log outtput parsing failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("unexpected amount of logs produced; expected 2, got %d", len(entries))
	}

	reqEntry := entries[0]
	if diff := cmp.Diff(reqEntry, map[string]interface{}{
		"@level":               "debug",
		"@message":             "",
		"@module":              "provider.test",
		"__HTTP_OP_TYPE__":     "HTTP_REQ",
		"__HTTP_REQ_METHOD__":  "GET",
		"__HTTP_REQ_URI__":     "/",
		"__HTTP_REQ_VERSION__": "HTTP/1.1",
		"Accept-Encoding":      "gzip",
		"Host":                 "www.terraform.io",
		"User-Agent":           "Go-http-client/1.1",
	}); diff != "" {
		t.Fatalf("unexpected difference for logging of the request:\n%s", diff)
	}

	resEntry := entries[1]
	expectedResEntryFields := map[string]interface{}{
		"@level":               "debug",
		"@module":              "provider.test",
		"Content-Type":         "text/html",
		"__HTTP_OP_TYPE__":     "HTTP_RES",
		"__HTTP_RES_STATUS__":  float64(200),
		"__HTTP_RES_VERSION__": "HTTP/2.0",
		"__HTTP_RES_REASON__":  "OK",
	}
	for ek, ev := range expectedResEntryFields {
		if resEntry[ek] != ev {
			t.Fatalf("Unexpected value for field %q; expected %q, got %q", ek, ev, resEntry[ek])
		}
	}
}

func TestNewLoggingHTTPTransport_LogMasking(t *testing.T) {
	ctx, loggerOutput := setupRootLogger(t)

	transport := logging.NewLoggingHTTPTransport("test", http.DefaultTransport)
	transport.WithConfigureRequestContext(func(ctx context.Context, subsystem string) context.Context {
		ctx = tflog.SubsystemMaskFieldValuesWithFieldKeys(ctx, subsystem, "__HTTP_OP_TYPE__")
		ctx = tflog.SubsystemMaskMessageRegexes(ctx, subsystem, regexp.MustCompile(`<html>.*</html>`))
		return ctx
	})
	client := http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	req, _ := http.NewRequest("GET", "https://www.terraform.io", nil)
	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	entries, err := tflogtest.MultilineJSONDecode(loggerOutput)
	if err != nil {
		t.Fatalf("log outtput parsing failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("unexpected amount of logs produced; expected 2, got %d", len(entries))
	}

	expectedMaskedEntryFields := map[string]interface{}{
		"__HTTP_OP_TYPE__": "***",
		"@message":         "<!DOCTYPE html>***",
	}
	for _, entry := range entries {
		for ek, ev := range expectedMaskedEntryFields {
			if entry[ek] != "" && entry[ek] != ev {
				t.Fatalf("Unexpected value for field %q; expected %q, got %q", ek, ev, entry[ek])
			}
		}
	}
}

func setupRootLogger(t *testing.T) (context.Context, *bytes.Buffer) {
	t.Setenv(logging.EnvLog, "TRACE")
	var output bytes.Buffer
	return tflogtest.RootLogger(context.Background(), &output), &output
}
