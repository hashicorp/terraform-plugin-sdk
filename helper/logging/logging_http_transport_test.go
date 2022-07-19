package logging_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

func TestNewLoggingHTTPTransport(t *testing.T) {
	ctx, loggerOutput := setupRootLogger()

	transport := logging.NewLoggingHTTPTransport(http.DefaultTransport)
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
		"@level":              "debug",
		"@message":            "",
		"@module":             "provider",
		"tf_http_op_type":     "request",
		"tf_http_req_method":  "GET",
		"tf_http_req_uri":     "/",
		"tf_http_req_version": "HTTP/1.1",
		"Accept-Encoding":     "gzip",
		"Host":                "www.terraform.io",
		"User-Agent":          "Go-http-client/1.1",
	}); diff != "" {
		t.Fatalf("unexpected difference for logging of the request:\n%s", diff)
	}

	resEntry := entries[1]
	expectedResEntryFields := map[string]interface{}{
		"@level":              "debug",
		"@module":             "provider",
		"Content-Type":        "text/html",
		"tf_http_op_type":     "response",
		"tf_http_res_status":  float64(200),
		"tf_http_res_version": "HTTP/2.0",
		"tf_http_res_reason":  "OK",
	}
	for ek, ev := range expectedResEntryFields {
		if resEntry[ek] != ev {
			t.Fatalf("Unexpected value for field %q; expected %q, got %q", ek, ev, resEntry[ek])
		}
	}
}

func TestNewSubsystemLoggingHTTPTransport(t *testing.T) {
	subsys := "test-subsystem"

	ctx, loggerOutput := setupRootLogger()
	ctx = tflog.NewSubsystem(ctx, subsys)

	transport := logging.NewSubsystemLoggingHTTPTransport(subsys, http.DefaultTransport)
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
		"@level":              "debug",
		"@message":            "",
		"@module":             "provider.test-subsystem",
		"tf_http_op_type":     "request",
		"tf_http_req_method":  "GET",
		"tf_http_req_uri":     "/",
		"tf_http_req_version": "HTTP/1.1",
		"Accept-Encoding":     "gzip",
		"Host":                "www.terraform.io",
		"User-Agent":          "Go-http-client/1.1",
	}); diff != "" {
		t.Fatalf("unexpected difference for logging of the request:\n%s", diff)
	}

	resEntry := entries[1]
	expectedResEntryFields := map[string]interface{}{
		"@level":              "debug",
		"@module":             "provider.test-subsystem",
		"Content-Type":        "text/html",
		"tf_http_op_type":     "response",
		"tf_http_res_status":  float64(200),
		"tf_http_res_version": "HTTP/2.0",
		"tf_http_res_reason":  "OK",
	}
	for ek, ev := range expectedResEntryFields {
		if resEntry[ek] != ev {
			t.Fatalf("Unexpected value for field %q; expected %q, got %q", ek, ev, resEntry[ek])
		}
	}
}

//func TestNewLoggingHTTPTransport_LogMasking(t *testing.T) {
//	ctx, loggerOutput := setupRootLogger(t)
//
//	transport := logging.NewSubsystemLoggingHTTPTransport("test", http.DefaultTransport)
//	transport.WithConfigureRequestContext(func(ctx context.Context, subsystem string) context.Context {
//		ctx = tflog.SubsystemMaskFieldValuesWithFieldKeys(ctx, subsystem, "tf_http_op_type")
//		ctx = tflog.SubsystemMaskMessageRegexes(ctx, subsystem, regexp.MustCompile(`<html>.*</html>`))
//		return ctx
//	})
//	client := http.Client{
//		Transport: transport,
//		Timeout:   10 * time.Second,
//	}
//
//	req, _ := http.NewRequest("GET", "https://www.terraform.io", nil)
//	res, err := client.Do(req.WithContext(ctx))
//	if err != nil {
//		t.Fatalf("request failed: %v", err)
//	}
//	defer res.Body.Close()
//
//	entries, err := tflogtest.MultilineJSONDecode(loggerOutput)
//	if err != nil {
//		t.Fatalf("log outtput parsing failed: %v", err)
//	}
//
//	if len(entries) != 2 {
//		t.Fatalf("unexpected amount of logs produced; expected 2, got %d", len(entries))
//	}
//
//	expectedMaskedEntryFields := map[string]interface{}{
//		"tf_http_op_type": "***",
//		"@message":        "<!DOCTYPE html>***",
//	}
//	for _, entry := range entries {
//		for ek, ev := range expectedMaskedEntryFields {
//			if entry[ek] != "" && entry[ek] != ev {
//				t.Fatalf("Unexpected value for field %q; expected %q, got %q", ek, ev, entry[ek])
//			}
//		}
//	}
//}

func setupRootLogger() (context.Context, *bytes.Buffer) {
	var output bytes.Buffer
	return tflogtest.RootLogger(context.Background(), &output), &output
}
