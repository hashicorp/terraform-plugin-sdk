package logging

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// NewLoggingHTTPTransport creates a wrapper around an *http.RoundTripper,
// designed to be used for the `Transport` field of http.Client.
//
// This logs each pair of HTTP request/response that it handles.
// The logging is done via `tflog`, that is part of the terraform-plugin-log
// library, included by this SDK.
//
// The request/response is logged via tflog.Debug, using the context.Context
// attached to the http.Request that the transport receives as input
// of http.RoundTripper RoundTrip method.
//
// It's responsibility of the developer using this transport, to ensure that each
// http.Request it handles is configured with the SDK-initialized Provider Root Logger
// context.Context, that it's passed to all resources/data-sources/provider entry-points
// (i.e. schema.Resource fields like `CreateContext`, `ReadContext`, etc.).
//
// This also gives the developer the flexibility to further configure the
// logging behaviour via the above-mentioned context: please see
// https://www.terraform.io/plugin/log/writing.
func NewLoggingHTTPTransport(t http.RoundTripper) *loggingHttpTransport {
	return &loggingHttpTransport{"", false, t}
}

// NewSubsystemLoggingHTTPTransport creates a wrapper around an *http.RoundTripper,
// designed to be used for the `Transport` field of http.Client.
//
// This logs each pair of HTTP request/response that it handles.
// The logging is done via `tflog`, that is part of the terraform-plugin-log
// library, included by this SDK.
//
// The request/response is logged via tflog.SubsystemDebug, using the context.Context
// attached to the http.Request that the transport receives as input
// of http.RoundTripper RoundTrip method, as well as the `subsystem` string
// provided at construction time.
//
// It's responsibility of the developer using this transport, to ensure that each
// http.Request it handles is configured with a Subsystem Logger
// context.Context that was initialized via tflog.NewSubsystem.
//
// This also gives the developer the flexibility to further configure the
// logging behaviour via the above-mentioned context: please see
// https://www.terraform.io/plugin/log/writing.
func NewSubsystemLoggingHTTPTransport(subsystem string, t http.RoundTripper) *loggingHttpTransport {
	return &loggingHttpTransport{subsystem, true, t}
}

const (
	// FieldHttpOperationType is the field key used by NewSubsystemLoggingHTTPTransport when logging the type of operation via tflog.
	FieldHttpOperationType = "tf_http_op_type"

	// OperationHttpRequest is the field value used by NewSubsystemLoggingHTTPTransport when logging a request via tflog.
	OperationHttpRequest = "request"

	// FieldHttpRequestMethod is the field key used by NewSubsystemLoggingHTTPTransport when logging a request method via tflog.
	FieldHttpRequestMethod = "tf_http_req_method"

	// FieldHttpRequestUri is the field key used by NewSubsystemLoggingHTTPTransport when logging a request URI via tflog.
	FieldHttpRequestUri = "tf_http_req_uri"

	// FieldHttpRequestVersion is the field key used by NewSubsystemLoggingHTTPTransport when logging a request HTTP version via tflog.
	FieldHttpRequestVersion = "tf_http_req_version"

	// OperationHttpResponse is the field value used by NewSubsystemLoggingHTTPTransport when logging a response via tflog.
	OperationHttpResponse = "response"

	// FieldHttpResponseVersion is the field key used by NewSubsystemLoggingHTTPTransport when logging a response HTTP version via tflog.
	FieldHttpResponseVersion = "tf_http_res_version"

	// FieldHttpResponseStatusCode is the field key used by NewSubsystemLoggingHTTPTransport when logging a response status code via tflog.
	FieldHttpResponseStatusCode = "tf_http_res_status"

	// FieldHttpResponseReason is the field key used by NewSubsystemLoggingHTTPTransport when logging a response reason phrase via tflog.
	FieldHttpResponseReason = "tf_http_res_reason"
)

type loggingHttpTransport struct {
	subsystem      string
	logToSubsystem bool
	transport      http.RoundTripper
}

func (t *loggingHttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// Grub the outgoing request bytes
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		t.Error(ctx, "HTTP Request introspection failed", map[string]interface{}{
			"error": fmt.Sprintf("%#v", err),
		})
	}

	// Decompose the request bytes in a message (HTTP body) and fields (HTTP headers), then log it
	msg, fields, err := parseRequestBytes(reqDump)
	if err != nil {
		t.Error(ctx, "Failed to parse request bytes for logging", map[string]interface{}{
			"error": err,
		})
	} else {
		t.Debug(ctx, msg, fields)
	}

	// Invoke the wrapped RoundTrip now
	res, err := t.transport.RoundTrip(req)
	if err != nil {
		return res, err
	}

	// Grub the incoming response bytes
	resDump, err := httputil.DumpResponse(res, true)
	if err != nil {
		t.Error(ctx, "HTTP Response introspection error", map[string]interface{}{
			"error": fmt.Sprintf("%#v", err),
		})
	}

	// Decompose the response bytes in a message (HTTP body) and fields (HTTP headers), then log it
	msg, fields, err = parseResponseBytes(resDump)
	if err != nil {
		t.Error(ctx, "Failed to parse response bytes for logging", map[string]interface{}{
			"error": err,
		})
	} else {
		t.Debug(ctx, msg, fields)
	}

	return res, nil
}

func (t *loggingHttpTransport) Debug(ctx context.Context, msg string, fields ...map[string]interface{}) {
	if t.logToSubsystem {
		tflog.SubsystemDebug(ctx, t.subsystem, msg, fields...)
	} else {
		tflog.Debug(ctx, msg, fields...)
	}
}

func (t *loggingHttpTransport) Error(ctx context.Context, msg string, fields ...map[string]interface{}) {
	if t.logToSubsystem {
		tflog.SubsystemError(ctx, t.subsystem, msg, fields...)
	} else {
		tflog.Error(ctx, msg, fields...)
	}
}

func parseRequestBytes(b []byte) (string, map[string]interface{}, error) {
	parts := strings.Split(string(b), "\r\n")

	// We will end up with a number of fields equivalent to the number of parts + 1:
	// - the first will be split into 3 parts
	// - the last 2 are "end of headers" and "body"
	// - one extra for the operation type
	fields := make(map[string]interface{}, len(parts)+1)
	fields[FieldHttpOperationType] = OperationHttpRequest

	// HTTP Request Line
	reqLineParts := strings.Split(parts[0], " ")
	fields[FieldHttpRequestMethod] = reqLineParts[0]
	fields[FieldHttpRequestUri] = reqLineParts[1]
	fields[FieldHttpRequestVersion] = reqLineParts[2]

	// HTTP Request Headers
	var i int
	for i = 1; i < len(parts)-2; i++ {
		// Check if we reached the end of the headers
		if parts[i] == "" {
			break
		}

		headerParts := strings.Split(parts[i], ": ")
		if len(headerParts) != 2 {
			return "", nil, fmt.Errorf("failed to parse header line %q", parts[i])
		}
		fields[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
	}

	// HTTP Response Body: the last part is always the body (can be empty)
	return parts[len(parts)-1], fields, nil
}

func parseResponseBytes(b []byte) (string, map[string]interface{}, error) {
	parts := strings.Split(string(b), "\r\n")

	// We will end up with a number of fields equivalent to the number of parts:
	// - the first will be split into 3 parts
	// - the last 2 are "end of headers" and "body"
	// - one extra for the operation type
	fields := make(map[string]interface{}, len(parts))
	fields[FieldHttpOperationType] = OperationHttpResponse

	// HTTP Message Status Line
	reqLineParts := strings.Split(parts[0], " ")
	fields[FieldHttpResponseVersion] = reqLineParts[0]
	// NOTE: Unlikely, but if we can't parse the status code,
	// we set its field value to string
	statusCode, err := strconv.Atoi(reqLineParts[1])
	if err != nil {
		fields[FieldHttpResponseStatusCode] = reqLineParts[1]
	} else {
		fields[FieldHttpResponseStatusCode] = statusCode
	}
	fields[FieldHttpResponseReason] = reqLineParts[2]

	// HTTP Response Headers
	var i int
	for i = 1; i < len(parts)-2; i++ {
		// Check if we reached the end of the headers
		if parts[i] == "" {
			break
		}

		headerParts := strings.Split(parts[i], ": ")
		if len(headerParts) != 2 {
			return "", nil, fmt.Errorf("failed to parse header line %q", parts[i])
		}
		fields[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
	}

	// HTTP Response Body: the last part is always the body (can be empty)
	return parts[len(parts)-1], fields, nil
}
