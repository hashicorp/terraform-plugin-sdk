// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-uuid"
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

	reqBody := `An example
		multiline
		request body`
	req, _ := http.NewRequest("GET", "https://www.terraform.io", bytes.NewBufferString(reqBody))
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

	if transId, ok := entries[0]["tf_http_trans_id"]; !ok || transId != entries[1]["tf_http_trans_id"] {
		t.Fatalf("expected to find the same 'tf_http_trans_id' in both req/res entries, got %q", transId)
	}

	transId, ok := entries[0]["tf_http_trans_id"].(string)
	if !ok {
		t.Fatalf("expected 'tf_http_trans_id' to be a string, got %T", transId)
	}

	if _, err := uuid.ParseUUID(transId); err != nil {
		t.Fatalf("expected 'tf_http_trans_id' to be contain a valid UUID, but got an error: %v", err)
	}

	reqEntry := entries[0]
	if diff := cmp.Diff(reqEntry, map[string]interface{}{
		"@level":              "debug",
		"@message":            "Sending HTTP Request",
		"@module":             "provider",
		"tf_http_op_type":     "request",
		"tf_http_req_method":  "GET",
		"tf_http_req_uri":     "/",
		"tf_http_req_version": "HTTP/1.1",
		"tf_http_req_body":    "An example multiline request body",
		"tf_http_trans_id":    transId,
		"Accept-Encoding":     "gzip",
		"Host":                "www.terraform.io",
		"User-Agent":          "Go-http-client/1.1",
		"Content-Length":      "37",
	}); diff != "" {
		t.Fatalf("unexpected difference for logging of the request:\n%s", diff)
	}

	resEntry := entries[1]
	expectedResEntryFields := map[string]interface{}{
		"@level":                    "debug",
		"@module":                   "provider",
		"@message":                  "Received HTTP Response",
		"Content-Type":              "text/html; charset=utf-8",
		"tf_http_op_type":           "response",
		"tf_http_res_status_code":   float64(200),
		"tf_http_res_version":       "HTTP/2.0",
		"tf_http_res_status_reason": "200 OK",
		"tf_http_trans_id":          transId,
	}
	for ek, ev := range expectedResEntryFields {
		if resEntry[ek] != ev {
			t.Fatalf("Unexpected value for field %q; expected %q, got %q", ek, ev, resEntry[ek])
		}
	}

	expectedNonEmptyEntryFields := []string{
		"tf_http_res_body", "Etag", "Date", "X-Frame-Options", "Server",
	}
	for _, ek := range expectedNonEmptyEntryFields {
		if ev, ok := resEntry[ek]; !ok || ev == "" {
			t.Fatalf("Expected field %q to contain a non-null value", ek)
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

	reqBody := `An example
		multiline
		request body`
	req, _ := http.NewRequest("GET", "https://www.terraform.io", bytes.NewBufferString(reqBody))
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

	if transId, ok := entries[0]["tf_http_trans_id"]; !ok || transId != entries[1]["tf_http_trans_id"] {
		t.Fatalf("expected to find the same 'tf_http_trans_id' in both req/res entries, got %q", transId)
	}

	transId, ok := entries[0]["tf_http_trans_id"].(string)
	if !ok {
		t.Fatalf("expected 'tf_http_trans_id' to be a string, got %T", transId)
	}

	if _, err := uuid.ParseUUID(transId); err != nil {
		t.Fatalf("expected 'tf_http_trans_id' to be contain a valid UUID, but got an error: %v", err)
	}

	reqEntry := entries[0]
	if diff := cmp.Diff(reqEntry, map[string]interface{}{
		"@level":              "debug",
		"@message":            "Sending HTTP Request",
		"@module":             "provider.test-subsystem",
		"tf_http_op_type":     "request",
		"tf_http_req_method":  "GET",
		"tf_http_req_uri":     "/",
		"tf_http_req_version": "HTTP/1.1",
		"tf_http_req_body":    "An example multiline request body",
		"tf_http_trans_id":    transId,
		"Accept-Encoding":     "gzip",
		"Host":                "www.terraform.io",
		"User-Agent":          "Go-http-client/1.1",
		"Content-Length":      "37",
	}); diff != "" {
		t.Fatalf("unexpected difference for logging of the request:\n%s", diff)
	}

	resEntry := entries[1]
	expectedResEntryFields := map[string]interface{}{
		"@level":                    "debug",
		"@module":                   "provider.test-subsystem",
		"@message":                  "Received HTTP Response",
		"Content-Type":              "text/html; charset=utf-8",
		"tf_http_op_type":           "response",
		"tf_http_res_status_code":   float64(200),
		"tf_http_res_version":       "HTTP/2.0",
		"tf_http_res_status_reason": "200 OK",
		"tf_http_trans_id":          transId,
	}
	for ek, ev := range expectedResEntryFields {
		if resEntry[ek] != ev {
			t.Fatalf("Unexpected value for field %q; expected %q, got %q", ek, ev, resEntry[ek])
		}
	}

	expectedNonEmptyEntryFields := []string{
		"tf_http_res_body", "Etag", "Date", "X-Frame-Options", "Server",
	}
	for _, ek := range expectedNonEmptyEntryFields {
		if ev, ok := resEntry[ek]; !ok || ev == "" {
			t.Fatalf("Expected field %q to contain a non-null value", ek)
		}
	}
}

func TestNewLoggingHTTPTransport_LogMasking(t *testing.T) {
	ctx, loggerOutput := setupRootLogger()
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "tf_http_op_type")
	ctx = tflog.MaskAllFieldValuesRegexes(ctx, regexp.MustCompile(`<html>.*</html>`))
	ctx = tflog.MaskMessageStrings(ctx, "Request", "Response")

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
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	entries, err := tflogtest.MultilineJSONDecode(loggerOutput)
	if err != nil {
		t.Fatalf("log outtput parsing failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("unexpected amount of logs produced; expected 2, got %d", len(entries))
	}

	if diff := cmp.Diff(entries[0]["@message"], "Sending HTTP ***"); diff != "" {
		t.Fatalf("unexpected difference for logging message of request:\n%s", diff)
	}

	if diff := cmp.Diff(entries[1]["@message"], "Received HTTP ***"); diff != "" {
		t.Fatalf("unexpected difference for logging message of response:\n%s", diff)
	}

	expectedMaskedEntryFields := map[string]interface{}{
		"tf_http_op_type":  "***",
		"tf_http_res_body": "<!DOCTYPE html>***",
	}
	for _, entry := range entries {
		for expectedK, expectedV := range expectedMaskedEntryFields {
			if entryV, ok := entry[expectedK]; ok && entryV != expectedV {
				t.Fatalf("Unexpected value for field %q; expected %q, got %q", expectedK, expectedV, entry[expectedK])
			}
		}
	}

	if diff := cmp.Diff(entries[1]["tf_http_res_body"], string(resBody)); diff == "" {
		t.Fatalf("expected HTTP response body and content of field 'tf_http_res_body' to differ, but they do not")
	}
}

func TestNewLoggingHTTPTransport_LogOmitting(t *testing.T) {
	ctx, loggerOutput := setupRootLogger()
	ctx = tflog.OmitLogWithMessageRegexes(ctx, regexp.MustCompile("(?i)rEsPoNsE"))
	ctx = tflog.OmitLogWithFieldKeys(ctx, "tf_http_req_method")

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

	if len(entries) != 0 {
		t.Fatalf("unexpected amount of logs produced; expected 0 (because they should have been omitted), got %d", len(entries))
	}
}

func setupRootLogger() (context.Context, *bytes.Buffer) {
	var output bytes.Buffer
	return tflogtest.RootLogger(context.Background(), &output), &output
}
