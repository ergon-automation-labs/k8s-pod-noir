package kubectl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Runner struct {
	Kubeconfig string
	Context    string
}

func (r *Runner) cmd(ctx context.Context, args ...string) *exec.Cmd {
	c := exec.CommandContext(ctx, "kubectl", args...)
	if r.Kubeconfig != "" {
		c.Env = append(os.Environ(), "KUBECONFIG="+r.Kubeconfig)
	}
	return c
}

// Run executes kubectl with args. Context is injected after global kubectl flags when set.
func (r *Runner) Run(ctx context.Context, args ...string) ([]byte, error) {
	full := kubeArgs(r.Context, args)
	var stderr bytes.Buffer
	c := r.cmd(ctx, full...)
	c.Stderr = &stderr
	out, err := c.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return out, fmt.Errorf("%w: %s", err, msg)
		}
		return out, err
	}
	return out, nil
}

// ApplyYAML runs kubectl apply -f - with manifest bytes on stdin.
func (r *Runner) ApplyYAML(ctx context.Context, yamlDoc []byte) error {
	var stderr bytes.Buffer
	args := kubeArgs(r.Context, []string{"apply", "-f", "-"})
	c := r.cmd(ctx, args...)
	c.Stdin = bytes.NewReader(yamlDoc)
	c.Stderr = &stderr
	c.Stdout = io.Discard
	if err := c.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("%w: %s", err, msg)
		}
		return err
	}
	return nil
}

// DeleteNamespace removes the namespace (and all resources inside). Idempotent.
func (r *Runner) DeleteNamespace(ctx context.Context, name string) error {
	_, err := r.Run(ctx, "delete", "namespace", name, "--ignore-not-found", "--wait=false")
	return err
}

// RolloutStatus waits until the deployment rollout completes or times out.
func (r *Runner) RolloutStatus(ctx context.Context, namespace, deployment string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	sec := int(timeout.Seconds())
	if sec < 1 {
		sec = 1
	}
	_, err := r.Run(ctx, "rollout", "status", "deployment/"+deployment, "-n", namespace, fmt.Sprintf("--timeout=%ds", sec))
	return err
}

// KubeContext returns the kubeconfig context name (may be empty).
func (r *Runner) KubeContext() string { return r.Context }

// KubeconfigPath returns the kubeconfig file path (may be empty).
func (r *Runner) KubeconfigPath() string { return r.Kubeconfig }

// Compile-time check that *Runner implements Kube.
var _ Kube = (*Runner)(nil)

func kubeArgs(context string, args []string) []string {
	if context == "" {
		return append([]string{}, args...)
	}
	// Insert --context after any leading global flags (minimal: prepend).
	return append([]string{"--context", context}, args...)
}
