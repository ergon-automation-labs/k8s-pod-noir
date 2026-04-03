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

	"overnight": Case001, "shift": Case001, "payments": Case001,
	"ghost": Case002, "credential": Case002, "ledger": Case002,
	"harbour": Case003, "harbor": Case003, "dead": Case003, "shipping": Case003,
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
	return "", fmt.Errorf("unknown folder %q — try 1–3, a case id, or quit", raw)
}
