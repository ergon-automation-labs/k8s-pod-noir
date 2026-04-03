package kubectl

import (
	"fmt"
	"regexp"
	"strings"
)

// EnsureSolvePolicy applies namespace rules and blocks cluster-wide or node-level kubectl in solve mode.
func EnsureSolvePolicy(line, allowedNS string) error {
	if err := EnsureMutatingUsesGameNamespace(line, allowedNS); err != nil {
		return err
	}
	s := kubectlCommandBody(line)
	if err := ensureSolveMutatingFilenameHasNamespace(s, allowedNS); err != nil {
		return err
	}
	low := strings.ToLower(s)

	if strings.Contains(low, "--all-namespaces") {
		return fmt.Errorf("precinct policy: --all-namespaces is blocked in solve mode")
	}
	// Match only kubectl's uppercase -A (--all-namespaces). Do not lowercase first: -a can mean other flags.
	if allNamespacesShortFlag.MatchString(s) {
		return fmt.Errorf("precinct policy: -A (all namespaces) is blocked in solve mode")
	}
	if low == "adm" || strings.HasPrefix(low, "adm ") {
		return fmt.Errorf("precinct policy: kubectl adm is blocked in solve mode")
	}
	if kustomizeDirFlag.MatchString(s) {
		return fmt.Errorf("precinct policy: -k (kustomize directory) is blocked in solve mode")
	}
	if low == "kustomize" || strings.HasPrefix(low, "kustomize ") {
		return fmt.Errorf("precinct policy: kubectl kustomize is blocked in solve mode")
	}
	if strings.HasPrefix(low, "certificate ") {
		return fmt.Errorf("precinct policy: certificate operations are blocked in solve mode")
	}

	for _, d := range solveBlockedSubstrings {
		if strings.Contains(low, d) {
			return fmt.Errorf("precinct policy: blocked in solve mode (%s)", strings.TrimSpace(d))
		}
	}
	if strings.HasPrefix(low, "taint ") {
		return fmt.Errorf("precinct policy: node taint operations are blocked in solve mode")
	}

	if strings.Contains(low, "delete") && strings.Contains(low, "namespace") {
		return fmt.Errorf("precinct policy: namespace delete is blocked in solve mode")
	}

	return nil
}

func ensureSolveMutatingFilenameHasNamespace(s, allowedNS string) error {
	if allowedNS == "" {
		return nil
	}
	if !mutatingPattern.MatchString(s) {
		return nil
	}
	if !filenameOptionPresent(s) {
		return nil
	}
	if namespaceSpecified(s, allowedNS) {
		return nil
	}
	return fmt.Errorf("precinct policy: mutating -f/--filename must include -n %q (or --namespace=...) in solve mode", allowedNS)
}

// Uppercase A only: kubectl uses -A for --all-namespaces; lowercase -a is a different flag on some subcommands.
var allNamespacesShortFlag = regexp.MustCompile(`(^|\s)-A(\s|$)`)

// -k points at a kustomization directory; manifests are not scanned the same way as -f.
var kustomizeDirFlag = regexp.MustCompile(`(^|\s)-k(\s|=|$)`)

var solveBlockedSubstrings = []string{
	"delete clusterrole", "delete clusterrolebinding",
	"delete validatingwebhookconfiguration", "delete mutatingwebhookconfiguration",
	"delete customresourcedefinition", " delete node", "delete nodes",
	" drain ", " cordon ", " uncordon ",
	" label node", " annotate node",
	"create clusterrole", "create clusterrolebinding",
}
