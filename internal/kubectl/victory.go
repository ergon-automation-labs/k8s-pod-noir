package kubectl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"podnoir/internal/scenario"
)

// DefaultVictoryTimeout is how long debrief waits for the cluster to look healthy.
const DefaultVictoryTimeout = 90 * time.Second

// VictoryForDefinition checks rollout or endpoints depending on scenario.VictoryMode.
func VictoryForDefinition(ctx context.Context, r *Runner, ns string, d *scenario.Definition, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = DefaultVictoryTimeout
	}
	mode := strings.TrimSpace(d.VictoryMode)
	if mode == "" {
		mode = "rollout"
	}
	switch mode {
	case "rollout":
		dep := d.SolveDeployment
		if dep == "" {
			return fmt.Errorf("scenario %s has no SolveDeployment for victory check", d.ID)
		}
		if err := r.RolloutStatus(ctx, ns, dep, timeout); err != nil {
			return fmt.Errorf("deployment/%s not healthy yet in %s: %w", dep, ns, err)
		}
		return nil
	case "endpoints":
		svc := strings.TrimSpace(d.VictoryService)
		if svc == "" {
			return fmt.Errorf("scenario %s has VictoryMode=endpoints but no VictoryService", d.ID)
		}
		return waitEndpointsReady(ctx, r, ns, svc, timeout)
	default:
		return fmt.Errorf("unknown VictoryMode %q", d.VictoryMode)
	}
}

func waitEndpointsReady(ctx context.Context, r *Runner, ns, svc string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		out, err := r.Run(ctx, "get", "endpoints", svc, "-n", ns, "-o", "jsonpath={.subsets[*].addresses[*].ip}")
		if err == nil && strings.TrimSpace(string(out)) != "" {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("service %s/%s has no endpoint addresses yet — fix the Service selector", ns, svc)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}
