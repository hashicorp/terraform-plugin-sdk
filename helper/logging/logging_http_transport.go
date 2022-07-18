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

const (
	// FieldHttpOperationType is the field key used by NewLoggingHTTPTransport when logging the type of operation via tflog.
	FieldHttpOperationType = "__HTTP_OP_TYPE__"

	// OperationHttpRequest is the field value used by NewLoggingHTTPTransport when logging a request via tflog.
	OperationHttpRequest = "HTTP_REQ"

	// FieldHttpRequestMethod is the field key used by NewLoggingHTTPTransport when logging a request method via tflog.
	FieldHttpRequestMethod = "__HTTP_REQ_METHOD__"

	// FieldHttpRequestUri is the field key used by NewLoggingHTTPTransport when logging a request URI via tflog.
	FieldHttpRequestUri = "__HTTP_REQ_URI__"

	// FieldHttpRequestVersion is the field key used by NewLoggingHTTPTransport when logging a request HTTP version via tflog.
	FieldHttpRequestVersion = "__HTTP_REQ_VERSION__"

	// OperationHttpResponse is the field value used by NewLoggingHTTPTransport when logging a response via tflog.
	OperationHttpResponse = "HTTP_RES"

	// FieldHttpResponseVersion is the field key used by NewLoggingHTTPTransport when logging a response HTTP version via tflog.
	FieldHttpResponseVersion = "__HTTP_RES_VERSION__"

	// FieldHttpResponseStatusCode is the field key used by NewLoggingHTTPTransport when logging a response status code via tflog.
	FieldHttpResponseStatusCode = "__HTTP_RES_STATUS__"

	// FieldHttpResponseReason is the field key used by NewLoggingHTTPTransport when logging a response reason phrase via tflog.
	FieldHttpResponseReason = "__HTTP_RES_REASON__"
)

// ConfigureReqCtxFunc is the type of function accepted by loggingHTTPTransport.WithConfigureRequestContext,
// to configure the request subsystem logging context before the actual logging.
type ConfigureReqCtxFunc func(ctx context.Context, subsystem string) context.Context

type loggingHTTPTransport struct {
	subsystem       string
	transport       http.RoundTripper
	configureReqCtx ConfigureReqCtxFunc
}

func (t *loggingHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create a new subsystem logging context from the request context
	ctx := tflog.NewSubsystem(req.Context(), t.subsystem)

	// If set, allow for further configuration of the new subsystem logging context
	if t.configureReqCtx != nil {
		ctx = t.configureReqCtx(ctx, t.subsystem)
	}

	if IsDebugOrHigher() {
		// Grub the outgoing request bytes
		reqDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			tflog.SubsystemError(ctx, t.subsystem, "HTTP Request introspection failed", map[string]interface{}{
				"error": fmt.Sprintf("%#v", err),
			})
		}

		// Decompose the request bytes in a message (HTTP body) and fields (HTTP headers), then log it
		msg, fields := decomposeRequestBytes(reqDump)
		tflog.SubsystemDebug(ctx, t.subsystem, msg, fields)
	}

	// Invoke the wrapped RoundTrip now
	res, err := t.transport.RoundTrip(req)
	if err != nil {
		return res, err
	}

	if IsDebugOrHigher() {
		// Grub the incoming response bytes
		resDump, err := httputil.DumpResponse(res, true)
		if err != nil {
			tflog.SubsystemError(ctx, t.subsystem, "HTTP Response introspection error", map[string]interface{}{
				"error": fmt.Sprintf("%#v", err),
			})
		}

		// Decompose the response bytes in a message (HTTP body) and fields (HTTP headers), then log it
		msg, fields := decomposeResponseBytes(resDump)
		tflog.SubsystemDebug(ctx, t.subsystem, msg, fields)
	}

	return res, nil
}

// NewLoggingHTTPTransport creates a wrapper around a *http.RoundTripper,
// designed to be used for the `Transport` field of http.Client.
//
// This logs each pair of HTTP request/response that it handles.
// The logging is done via `tflog`, that is part of the terraform-plugin-log
// library, included by this SDK.
//
// The request/response is logged via tflog.SubsystemDebug, and the `subsystem`
// is the one provide here.
//
// IMPORTANT: For logging to work, it's mandatory that each http.Request it handles
// is configured with the Context (i.e. `Request.WithContext(ctx)`)
// that the SDK passes into all resources/data-sources/provider entry-points
// (i.e. schema.Resource fields like `CreateContext`, `ReadContext`, etc.).
func NewLoggingHTTPTransport(subsystem string, t http.RoundTripper) *loggingHTTPTransport {
	return &loggingHTTPTransport{subsystem, t, nil}
}

// WithConfigureRequestContext allows to optionally configure a callback ConfigureReqCtxFunc.
// This is used by the underlying structure to allow user to configure the
// http.Request context.Context, before the request is executed and the logger
// tflog.SubsystemDebug invoked.
//
// Log entries will be structured to contain:
//
//   * the HTTP message first line, broken up into "fields" (see `FieldHttp*` constants)
//   * the Headers, each as "fields" of the log
//   * the Request/Response Body as "message" of the log
//
// For example, here is how to add an extra field to each log emitted here,
// as well as masking fields that have a specific key:
//
//   t.WithConfigureRequestContext(func (ctx context.Context, subsystem string) context.Context {
//     ctx = tflog.SetField(ctx, "additional_key", "additional_value")
//     ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "secret", "token", "Authorization")
//   })
//
func (t *loggingHTTPTransport) WithConfigureRequestContext(callback ConfigureReqCtxFunc) *loggingHTTPTransport {
	t.configureReqCtx = callback
	return t
}

func decomposeRequestBytes(b []byte) (string, map[string]interface{}) {
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
		fields[headerParts[0]] = headerParts[1]
	}

	// HTTP Response Body: the last part is always the body (can be empty)
	return parts[len(parts)-1], fields
}

func decomposeResponseBytes(b []byte) (string, map[string]interface{}) {
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
		fields[headerParts[0]] = headerParts[1]
	}

	// HTTP Response Body: the last part is always the body (can be empty)
	return parts[len(parts)-1], fields
}
