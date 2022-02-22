package logging

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

const (
	// SubsystemHelperResource is the tfsdklog subsystem name for helper/resource.
	SubsystemHelperResource = "helper_resource"
)

// HelperResourceTrace emits a helper/resource subsystem log at TRACE level.
func HelperResourceTrace(ctx context.Context, msg string, args ...interface{}) {
	tfsdklog.SubsystemTrace(ctx, SubsystemHelperResource, msg, args)
}

// HelperResourceDebug emits a helper/resource subsystem log at DEBUG level.
func HelperResourceDebug(ctx context.Context, msg string, args ...interface{}) {
	tfsdklog.SubsystemDebug(ctx, SubsystemHelperResource, msg, args)
}

// HelperResourceWarn emits a helper/resource subsystem log at WARN level.
func HelperResourceWarn(ctx context.Context, msg string, args ...interface{}) {
	tfsdklog.SubsystemWarn(ctx, SubsystemHelperResource, msg, args)
}

// HelperResourceError emits a helper/resource subsystem log at ERROR level.
func HelperResourceError(ctx context.Context, msg string, args ...interface{}) {
	tfsdklog.SubsystemError(ctx, SubsystemHelperResource, msg, args)
}
