package contacts

import (
	"strings"
	"testing"
)

func TestWireRoster(t *testing.T) {
	inv := &InvestigationState{
		SeniorDetectiveUnlocked:        true,
		SeniorHintDelivered:            false,
		SysadminUnlocked:               false,
		NetworkEngineerUnlocked:        true,
		NetworkEngineerHintDelivered:   true,
	}
	out := WireRoster(inv)
	if out == "" {
		t.Fatal("empty roster")
	}
	if !strings.Contains(out, "Senior Detective") || !strings.Contains(out, "open") {
		t.Fatalf("expected senior open: %q", out)
	}
	if !strings.Contains(out, "locked") || !strings.Contains(out, "examine pod") {
		t.Fatalf("expected sysadmin locked: %q", out)
	}
	if !strings.Contains(out, "done") {
		t.Fatalf("expected network done: %q", out)
	}
}
