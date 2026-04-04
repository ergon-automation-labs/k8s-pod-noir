package session

import (
	"context"
	"fmt"
	"time"

	"podnoir/internal/kubectl"
)

// fakeKube is a test double for [kubectl.Kube]. If [fakeKube.run] is nil, [Run] errors.
// If [fakeKube.RolloutErr] is set, [RolloutStatus] returns it (for debrief / victory tests).
type fakeKube struct {
	run        func(ctx context.Context, args []string) ([]byte, error)
	RolloutErr error
}

func (f *fakeKube) Run(ctx context.Context, args ...string) ([]byte, error) {
	if f != nil && f.run != nil {
		return f.run(ctx, args)
	}
	return nil, fmt.Errorf("unexpected kubectl: %v", args)
}

func (f *fakeKube) ApplyYAML(ctx context.Context, yamlDoc []byte) error { return nil }

func (f *fakeKube) RolloutStatus(ctx context.Context, namespace, deployment string, timeout time.Duration) error {
	if f != nil && f.RolloutErr != nil {
		return f.RolloutErr
	}
	return nil
}

func (f *fakeKube) KubeContext() string { return "" }

func (f *fakeKube) KubeconfigPath() string { return "" }

var _ kubectl.Kube = (*fakeKube)(nil)
