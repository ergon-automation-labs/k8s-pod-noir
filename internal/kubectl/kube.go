package kubectl

import (
	"context"
	"time"
)

// Kube is the kubectl surface used by the session REPL and victory checks.
// The concrete [*Runner] implements this interface.
type Kube interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
	ApplyYAML(ctx context.Context, yamlDoc []byte) error
	RolloutStatus(ctx context.Context, namespace, deployment string, timeout time.Duration) error
	KubeContext() string
	KubeconfigPath() string
}
