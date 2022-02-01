package logging

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

// InitContext creates SDK logger contexts.
func InitContext(ctx context.Context) context.Context {
	ctx = tfsdklog.NewSubsystem(ctx, SubsystemHelperSchema, tfsdklog.WithLevelFromEnv(EnvTfLogSdkHelperSchema))

	return ctx
}
