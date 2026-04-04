package contacts

import (
	"fmt"
	"strings"
)

// ParseHintTarget maps REPL tokens to contact IDs (hint sysadmin, hint network, …).
func ParseHintTarget(s string) (ID, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "senior", "detective", "wire-two", "sd":
		return SeniorDetective, nil
	case "sysadmin", "sys", "basement", "ops":
		return Sysadmin, nil
	case "network", "net", "routes", "trunk":
		return NetworkEngineer, nil
	case "archivist", "archive", "clerk", "records":
		return Archivist, nil
	default:
		return "", fmt.Errorf("unknown contact %q — try: senior, sysadmin, network, archivist", s)
	}
}
