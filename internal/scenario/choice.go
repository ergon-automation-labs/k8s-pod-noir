package scenario

import (
	"errors"
	"fmt"
	"strings"
)

// ErrMenuQuit means the player left from the case file menu without opening a folder.
var ErrMenuQuit = errors.New("menu quit")

var choiceAliases = map[string]ID{
	"1": Case001, "01": Case001, "one": Case001,
	"2": Case002, "02": Case002, "two": Case002,
	"3": Case003, "03": Case003, "three": Case003,
	"4": Case004, "04": Case004, "four": Case004,
	"5": Case005, "05": Case005, "five": Case005,
	"6": Case006, "06": Case006, "six": Case006,
	"7": Case007, "07": Case007, "seven": Case007,
	"8": Case008, "08": Case008, "eight": Case008,
	"9": Case009, "09": Case009, "nine": Case009,
	"10": Case010, "ten": Case010,

	"overnight": Case001, "shift": Case001, "payments": Case001,
	"ghost": Case002, "credential": Case002, "ledger": Case002,
	"harbour": Case003, "harbor": Case003, "dead": Case003, "shipping": Case003,
	"probe": Case004, "bedside": Case004, "liveness": Case004,
	"oom": Case005, "memory": Case005, "margin": Case005,
	"wire": Case006, "gateway": Case006, "selector": Case006,
	"witness": Case007, "init": Case007, "gate": Case007,
	"quota": Case008, "redtape": Case008, "bureaucracy": Case008,
	"locker": Case009, "pvc": Case009, "evidence": Case009, "storage": Case009,
	"corridor": Case010, "policy": Case010, "networkpolicy": Case010, "silent": Case010,
	"tape": Case010, "deck": Case010,
}

// MatchCaseChoice maps menu input to a scenario ID.
// Returns ErrMenuQuit for quit synonyms.
func MatchCaseChoice(raw string) (ID, error) {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return "", fmt.Errorf("empty choice")
	}
	switch s {
	case "quit", "q", "exit", "bye", "leave", "walk":
		return "", ErrMenuQuit
	}
	if id, ok := choiceAliases[s]; ok {
		return id, nil
	}
	// Exact or partial id match
	id := ID(strings.TrimSpace(raw))
	if _, err := ByID(id); err == nil {
		return id, nil
	}
	for _, known := range List() {
		if strings.EqualFold(string(known), raw) {
			return known, nil
		}
	}
	return "", fmt.Errorf("unknown folder %q — try 1–10, a case id, or quit", raw)
}
