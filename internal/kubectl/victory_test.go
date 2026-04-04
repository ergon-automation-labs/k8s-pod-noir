package kubectl

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"podnoir/internal/scenario"
)

// stubKube implements [Kube] with controllable [RolloutStatus] / [Run] for victory tests.
type stubKube struct {
	RolloutErr error
	RunOut     []byte
	RunErr     error
}

func (s *stubKube) Run(ctx context.Context, args ...string) ([]byte, error) {
	if s.RunErr != nil {
		return s.RunOut, s.RunErr
	}
	return s.RunOut, nil
}

func (s *stubKube) ApplyYAML(ctx context.Context, yamlDoc []byte) error { return nil }

func (s *stubKube) RolloutStatus(ctx context.Context, namespace, deployment string, timeout time.Duration) error {
	if s != nil && s.RolloutErr != nil {
		return s.RolloutErr
	}
	return nil
}

func (s *stubKube) KubeContext() string    { return "" }
func (s *stubKube) KubeconfigPath() string { return "" }

var _ Kube = (*stubKube)(nil)

func TestVictoryForDefinitionRolloutFailsPropagates(t *testing.T) {
	t.Parallel()
	def := &scenario.Definition{
		ID:              scenario.Case001,
		SolveDeployment: "payments-worker",
		VictoryMode:     "rollout",
	}
	k := &stubKube{RolloutErr: errors.New("not ready")}
	err := VictoryForDefinition(context.Background(), k, "pod-noir", def, time.Second)
	if err == nil || !strings.Contains(err.Error(), "payments-worker") {
		t.Fatalf("expected rollout error mentioning deployment, got %v", err)
	}
}

func TestVictoryForDefinitionRolloutOK(t *testing.T) {
	t.Parallel()
	def := &scenario.Definition{
		ID:              scenario.Case001,
		SolveDeployment: "payments-worker",
	}
	k := &stubKube{}
	if err := VictoryForDefinition(context.Background(), k, "pod-noir", def, time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestVictoryForDefinitionEndpointsWaitsForAddresses(t *testing.T) {
	t.Parallel()
	def, err := scenario.ByID(scenario.Case006)
	if err != nil {
		t.Fatal(err)
	}
	k := &stubKube{
		RunOut: []byte("10.0.0.1"),
	}
	if err := VictoryForDefinition(context.Background(), k, def.Namespace, def, 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
