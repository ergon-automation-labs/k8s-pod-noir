package scenario

import (
	"errors"
	"testing"
)

func TestMatchCaseChoice(t *testing.T) {
	id, err := MatchCaseChoice("1")
	if err != nil || id != Case001 {
		t.Fatalf("1: got %v %v", id, err)
	}
	id, err = MatchCaseChoice("ledger")
	if err != nil || id != Case002 {
		t.Fatalf("ledger alias: got %v %v", id, err)
	}
	_, err = MatchCaseChoice("quit")
	if !errors.Is(err, ErrMenuQuit) {
		t.Fatalf("quit: %v", err)
	}
	id, err = MatchCaseChoice(string(Case003))
	if err != nil || id != Case003 {
		t.Fatalf("full id: %v %v", id, err)
	}
	id, err = MatchCaseChoice("4")
	if err != nil || id != Case004 {
		t.Fatalf("4: %v %v", id, err)
	}
	id, err = MatchCaseChoice("7")
	if err != nil || id != Case007 {
		t.Fatalf("7: %v %v", id, err)
	}
	id, err = MatchCaseChoice("10")
	if err != nil || id != Case010 {
		t.Fatalf("10: %v %v", id, err)
	}
	_, err = MatchCaseChoice("nope")
	if err == nil {
		t.Fatal("expected error")
	}
}
