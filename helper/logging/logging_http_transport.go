package logging

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httputil"
	"net/textproto"
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

	// FieldHttpRequestProtoVersion is the field key used by NewSubsystemLoggingHTTPTransport when logging a request HTTP version via tflog.
	FieldHttpRequestProtoVersion = "tf_http_req_version"

	// OperationHttpResponse is the field value used by NewSubsystemLoggingHTTPTransport when logging a response via tflog.
	OperationHttpResponse = "response"

	// FieldHttpResponseProtoVersion is the field key used by NewSubsystemLoggingHTTPTransport when logging a response HTTP protocol version via tflog.
	FieldHttpResponseProtoVersion = "tf_http_res_version"

	// FieldHttpResponseStatusCode is the field key used by NewSubsystemLoggingHTTPTransport when logging a response status code via tflog.
	FieldHttpResponseStatusCode = "tf_http_res_status_code"

	// FieldHttpResponseStatusReason is the field key used by NewSubsystemLoggingHTTPTransport when logging a response status reason phrase via tflog.
	FieldHttpResponseStatusReason = "tf_http_res_status_reason"
)

type loggingHttpTransport struct {
	subsystem      string
	logToSubsystem bool
	transport      http.RoundTripper
}

func (t *loggingHttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	//// Decompose the request bytes in a message (HTTP body) and fields (HTTP headers), then log it
	msg, fields, err := decomposeRequestForLogging(req)
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

	// Decompose the response bytes in a message (HTTP body) and fields (HTTP headers), then log it
	msg, fields, err = decomposeResponseForLogging(res)
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

func decomposeRequestForLogging(req *http.Request) (string, map[string]interface{}, error) {
	fields := make(map[string]interface{}, len(req.Header)+4)
	fields[FieldHttpOperationType] = OperationHttpRequest

	fields[FieldHttpRequestMethod] = req.Method
	fields[FieldHttpRequestUri] = req.URL.RequestURI()
	fields[FieldHttpRequestProtoVersion] = req.Proto

	// Get the full body of the request, including headers appended by http.Transport:
	// this is necessary because the http.Request at this stage doesn't contain
	// all the headers that will be eventually sent.
	// We rely on `httputil.DumpRequestOut` to obtain the actual bytes that will be sent out.
	reqBytes, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return "", nil, err
	}

	// Create a reader around the request full body
	reqReader := textproto.NewReader(bufio.NewReader(bytes.NewReader(reqBytes)))

	err = readHeadersIntoFields(reqReader, fields)
	if err != nil {
		return "", nil, err
	}

	// Read the rest of the body content,
	// into what will be the log message
	return readRestToString(reqReader), fields, nil
}

func decomposeResponseForLogging(res *http.Response) (string, map[string]interface{}, error) {
	fields := make(map[string]interface{}, len(res.Header)+4)
	fields[FieldHttpOperationType] = OperationHttpResponse

	fields[FieldHttpResponseProtoVersion] = res.Proto
	fields[FieldHttpResponseStatusCode] = res.StatusCode
	fields[FieldHttpResponseStatusReason] = res.Status

	// Get the full body of the response:
	// this is necessary because the http.Response `Body` is a ReadCloser,
	// so the only way to read it's content without affecting the final consumer
	// of this response, is to rely on `httputil.DumpRequestOut`,
	// that handles duplicating the underlying buffers and make the extraction
	// of these bytes transparent.
	resBytes, err := httputil.DumpResponse(res, true)
	if err != nil {
		return "", nil, err
	}

	// Create a reader around the response full body
	resReader := textproto.NewReader(bufio.NewReader(bytes.NewReader(resBytes)))

	err = readHeadersIntoFields(resReader, fields)
	if err != nil {
		return "", nil, err
	}

	// Read the rest of the body content,
	// into what will be the log message
	return readRestToString(resReader), fields, nil
}

func readHeadersIntoFields(reader *textproto.Reader, fields map[string]interface{}) error {
	// Ignore the first line: it contains non-header content
	// that we have already captured.
	// Skipping this step, would cause the following call to `ReadMIMEHeader()`
	// to fail as it cannot parse the first line.
	_, err := reader.ReadLine()
	if err != nil {
		return err
	}

	// Read the MIME-style headers
	mimeHeader, err := reader.ReadMIMEHeader()
	if err != nil {
		return err
	}

	// Set the headers as fields to log
	for k, v := range mimeHeader {
		if len(v) == 1 {
			fields[k] = v[0]
		} else {
			fields[k] = v
		}
	}

	return nil
}

func readRestToString(reader *textproto.Reader) string {
	var builder strings.Builder
	for {
		line, err := reader.ReadContinuedLine()
		if errors.Is(err, io.EOF) {
			break
		}
		builder.WriteString(line)
	}

	return builder.String()
}
