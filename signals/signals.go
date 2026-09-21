package signals

import (
	"context"
	"os/signal"
	"syscall"
)

// CreateSoftKillSignalContext helps to create signal Context to listen syscall.SIGINT and syscall.SIGTERM
// syscall.SIGINT  - user press Ctrl + C
// syscall.SIGTERM - start to rollout, kubernetes send sigterm
// syscall.SIGKILL - grace period timeout, process is killed
func CreateSoftKillSignalContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
}
