package kubectl

import (
	"fmt"
	"regexp"
	"strings"
)

var mutatingPattern = regexp.MustCompile(`(?i)\b(apply|create|delete|patch|replace|rollout|scale|set|annotate|label)\b`)

// EnsureMutatingUsesGameNamespace returns an error if the line looks like a mutating kubectl
// command but does not target the game namespace. Invocations using only -f / --filename are allowed.
func EnsureMutatingUsesGameNamespace(line, allowedNS string) error {
	s := strings.TrimSpace(line)
	if s == "" {
		return nil
	}
	if strings.HasPrefix(s, "kubectl ") {
		s = strings.TrimSpace(strings.TrimPrefix(s, "kubectl"))
	}
	if !mutatingPattern.MatchString(s) {
		return nil
	}
	if strings.Contains(s, " -f ") || strings.HasPrefix(s, "-f ") ||
		strings.Contains(s, " --filename=") || strings.HasPrefix(s, "--filename=") ||
		strings.Contains(s, " --filename ") {
		return nil
	}
	if allowedNS == "" {
		return nil
	}
	if namespaceSpecified(s, allowedNS) {
		return nil
	}
	return fmt.Errorf("mutating kubectl must target namespace %q (e.g. -n %s) or use -f with YAML in that namespace", allowedNS, allowedNS)
}

func namespaceSpecified(s, ns string) bool {
	prefixes := []string{
		"-n " + ns,
		"--namespace=" + ns,
		"--namespace " + ns,
	}
	for _, p := range prefixes {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
