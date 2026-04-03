package contacts

import (
	"testing"

	"podnoir/internal/llm"
)

func TestShouldUnlockSeniorFromAccusation(t *testing.T) {
	if !ShouldUnlockSeniorFromAccusation(llm.Cold) {
		t.Fatal("cold should unlock")
	}
	if !ShouldUnlockSeniorFromAccusation(llm.Warm) {
		t.Fatal("warm should unlock")
	}
	if ShouldUnlockSeniorFromAccusation(llm.Hot) {
		t.Fatal("hot should not unlock via accusation path")
	}
}

func TestShouldUnlockSeniorFromEvidence(t *testing.T) {
	st := &InvestigationState{}
	if ShouldUnlockSeniorFromEvidence(st) {
		t.Fatal("need both")
	}
	st.SeenLogs = true
	if ShouldUnlockSeniorFromEvidence(st) {
		t.Fatal("need trace")
	}
	st.SeenTrace = true
	if !ShouldUnlockSeniorFromEvidence(st) {
		t.Fatal("both logs and trace should unlock evidence path")
	}
}
