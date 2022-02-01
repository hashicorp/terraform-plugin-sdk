package logging

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

const (
	// SubsystemHelperSchema is the tfsdklog subsystem name for helper/schema.
	SubsystemHelperSchema = "helper_schema"
)

// HelperSchemaDebug emits a helper/schema subsystem log at DEBUG level.
func HelperSchemaDebug(ctx context.Context, msg string, args ...interface{}) {
	tfsdklog.SubsystemDebug(ctx, SubsystemHelperSchema, msg, args)
}

// HelperSchemaTrace emits a helper/schema subsystem log at TRACE level.
func HelperSchemaTrace(ctx context.Context, msg string, args ...interface{}) {
	tfsdklog.SubsystemTrace(ctx, SubsystemHelperSchema, msg, args)
}

// HelperSchemaWarn emits a helper/schema subsystem log at WARN level.
func HelperSchemaWarn(ctx context.Context, msg string, args ...interface{}) {
	tfsdklog.SubsystemWarn(ctx, SubsystemHelperSchema, msg, args)
}
