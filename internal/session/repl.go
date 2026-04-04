package session

import (
	"fmt"
	"strings"
)

// rememberSolveLine stores the last successful kubectl invocation for r/again in solve mode.
func (s *Session) rememberSolveLine(user string) {
	ln := strings.TrimSpace(user)
	if ln != "kubectl" && !strings.HasPrefix(ln, "kubectl ") {
		ln = "kubectl " + ln
	}
	if s.Kube != nil && s.Kube.KubeContext() != "" {
		ln = strings.Replace(ln, "kubectl ", "kubectl --context="+s.Kube.KubeContext()+" ", 1)
	}
	s.lastSolveKubectl = ln
}

const maxReplHistory = 60

func (s *Session) expandReplShortcuts(orig string, solveMode bool) (string, error) {
	low := strings.ToLower(strings.TrimSpace(orig))
	switch {
	case low == "r" || low == "again":
		if solveMode {
			if strings.TrimSpace(s.lastSolveKubectl) == "" {
				return "", fmt.Errorf("no previous kubectl to repeat")
			}
			return s.lastSolveKubectl, nil
		}
		if strings.TrimSpace(s.lastExpandedCmd) == "" {
			return "", fmt.Errorf("no previous command to repeat (try observe or check logs first)")
		}
		return s.lastExpandedCmd, nil
	}
	if solveMode {
		return strings.TrimSpace(orig), nil
	}

	switch {
	case low == "o" || low == "obs":
		return "observe", nil
	case low == "l" || low == "logs":
		if strings.TrimSpace(s.lastLogsPod) == "" {
			return "", fmt.Errorf(`"l" needs a pod after you've run "check logs <pod>" once, or use: l <pod>`)
		}
		return "check logs " + s.lastLogsPod, nil
	case strings.HasPrefix(low, "l ") && !strings.HasPrefix(low, "logs "):
		return "check logs " + strings.TrimSpace(orig[2:]), nil
	case strings.HasPrefix(low, "logs "):
		return "check logs " + strings.TrimSpace(orig[5:]), nil
	case strings.HasPrefix(low, "t "):
		return "trace " + strings.TrimSpace(orig[2:]), nil
	case strings.HasPrefix(low, "x "):
		return "examine pod " + strings.TrimSpace(orig[2:]), nil
	default:
		return strings.TrimSpace(orig), nil
	}
}

func (s *Session) recordReplSuccess(orig, expanded string, solveMode bool) {
	low := strings.ToLower(strings.TrimSpace(orig))
	if low != "r" && low != "again" {
		if solveMode {
			// lastSolveKubectl set after successful execKubectl
		} else {
			s.lastExpandedCmd = expanded
		}
	}
	line := orig
	if orig != expanded {
		line = fmt.Sprintf("%s → %s", orig, expanded)
	}
	s.replHistory = append(s.replHistory, line)
	if len(s.replHistory) > maxReplHistory {
		s.replHistory = s.replHistory[len(s.replHistory)-maxReplHistory:]
	}
}

func (s *Session) showHistory() {
	n := len(s.replHistory)
	if n == 0 {
		fmt.Fprintln(s.Out, "No commands in this session yet.")
		return
	}
	start := 0
	if n > 12 {
		start = n - 12
	}
	fmt.Fprintln(s.Out, "Recent commands (this session):")
	for i := start; i < n; i++ {
		fmt.Fprintf(s.Out, "  %d  %s\n", i+1, s.replHistory[i])
	}
}
