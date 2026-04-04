package session

import (
	"errors"
	"strings"
	"testing"

	"podnoir/internal/llm"
	"podnoir/internal/scenario"
)

func TestAccuseHotUnlocksSolveCase001(t *testing.T) {
	t.Parallel()
	def, err := scenario.ByID(scenario.Case001)
	if err != nil {
		t.Fatal(err)
	}
	s, buf := newTestSessionScenario(t, &fakeKube{}, def, def.Namespace)
	if err := s.handle(`accuse missing settings.json under /app/config`); err != nil {
		t.Fatal(err)
	}
	if !s.hotAccusation {
		t.Fatal("expected HOT accusation for case-001 + settings.json signal")
	}
	if !strings.Contains(buf.String(), string(llm.Hot)) {
		t.Fatalf("output should include HOT judgment: %s", buf.String())
	}
	buf.Reset()
	if err := s.handle("solve"); err != nil {
		t.Fatal(err)
	}
	if !s.solveMode {
		t.Fatal("expected solve mode after HOT + solve")
	}
	if !strings.Contains(buf.String(), "Solve mode") {
		t.Fatalf("expected solve banner: %s", buf.String())
	}
}

func TestSolveModePolicyBlocksAllNamespaces(t *testing.T) {
	t.Parallel()
	def, err := scenario.ByID(scenario.Case001)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := newTestSessionScenario(t, &fakeKube{}, def, def.Namespace)
	s.hotAccusation = true
	s.solveMode = true
	herr := s.handle("kubectl get pods -A")
	if herr == nil {
		t.Fatal("expected precinct policy error")
	}
	if !strings.Contains(herr.Error(), "precinct") {
		t.Fatalf("unexpected error: %v", herr)
	}
}

func TestDebriefBlockedWhenVictoryCheckFails(t *testing.T) {
	t.Parallel()
	def, err := scenario.ByID(scenario.Case001)
	if err != nil {
		t.Fatal(err)
	}
	k := &fakeKube{RolloutErr: errors.New("deployment not ready")}
	s, buf := newTestSessionScenario(t, k, def, def.Namespace)
	if err := s.handle("debrief"); err != nil {
		t.Fatalf("debrief should print and return nil when victory fails: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "duty sergeant") && !strings.Contains(out, "stamp") {
		t.Fatalf("expected blocked-debrief copy: %s", out)
	}
}

func TestDebriefSuccessReturnsQuit(t *testing.T) {
	t.Parallel()
	def, err := scenario.ByID(scenario.Case001)
	if err != nil {
		t.Fatal(err)
	}
	s, buf := newTestSessionScenario(t, &fakeKube{}, def, def.Namespace)
	err = s.handle("debrief")
	if err != errQuit {
		t.Fatalf("debrief after victory should return errQuit, got %v", err)
	}
	if !strings.Contains(buf.String(), "CASE #001") {
		t.Fatalf("expected mock debrief banner: %s", buf.String())
	}
}
