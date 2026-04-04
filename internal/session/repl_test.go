package session

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"podnoir/internal/kubectl"
)

func TestExpandReplShortcuts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		s         *Session
		orig      string
		solveMode bool
		want      string
		wantErr   string
	}{
		{
			name: "o to observe", s: &Session{}, orig: "o", want: "observe",
		},
		{
			name: "obs to observe", s: &Session{}, orig: "obs", want: "observe",
		},
		{
			name: "t name", s: &Session{}, orig: "t  pod-a", want: "trace pod-a",
		},
		{
			name: "x pod", s: &Session{}, orig: "x  bedside", want: "examine pod bedside",
		},
		{
			name: "l with pod suffix", s: &Session{}, orig: "l witness", want: "check logs witness",
		},
		{
			name: "logs prefix", s: &Session{}, orig: "logs  z", want: "check logs z",
		},
		{
			name: "l bare needs lastLogsPod",
			s:    &Session{lastLogsPod: "cached"},
			orig: "l",
			want: "check logs cached",
		},
		{
			name:    "l bare missing lastLogsPod",
			s:       &Session{},
			orig:    "l",
			wantErr: `"l" needs a pod`,
		},
		{
			name: "passthrough trim", s: &Session{}, orig: "  observe  ", want: "observe",
		},
		{
			name:      "solve mode raw kubectl",
			s:         &Session{},
			orig:      "kubectl get pods",
			solveMode: true,
			want:      "kubectl get pods",
		},
		{
			name:      "solve mode trim",
			s:         &Session{},
			orig:      "  rollout status  ",
			solveMode: true,
			want:      "rollout status",
		},
		{
			name: "r repeats last expanded",
			s: &Session{
				lastExpandedCmd: "observe",
			},
			orig: "r",
			want: "observe",
		},
		{
			name: "again repeats last expanded",
			s: &Session{
				lastExpandedCmd: "check logs foo",
			},
			orig: "again",
			want: "check logs foo",
		},
		{
			name:    "r without history",
			s:       &Session{},
			orig:    "r",
			wantErr: "no previous command",
		},
		{
			name: "solve r repeats kubectl",
			s: &Session{
				lastSolveKubectl: "kubectl get pods -n pod-noir",
			},
			orig:      "r",
			solveMode: true,
			want:      "kubectl get pods -n pod-noir",
		},
		{
			name:      "solve r without kubectl history",
			s:         &Session{},
			orig:      "r",
			solveMode: true,
			wantErr:   "no previous kubectl",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.s.expandReplShortcuts(tt.orig, tt.solveMode)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err=%v want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestRecordReplSuccess(t *testing.T) {
	t.Parallel()

	t.Run("stores expanded command for repeat", func(t *testing.T) {
		t.Parallel()
		s := &Session{}
		s.recordReplSuccess("o", "observe", false)
		if s.lastExpandedCmd != "observe" {
			t.Fatalf("lastExpandedCmd=%q", s.lastExpandedCmd)
		}
		if len(s.replHistory) != 1 || s.replHistory[0] != "o → observe" {
			t.Fatalf("history=%v", s.replHistory)
		}
	})

	t.Run("r does not clobber lastExpandedCmd", func(t *testing.T) {
		t.Parallel()
		s := &Session{lastExpandedCmd: "observe"}
		s.recordReplSuccess("r", "observe", false)
		if s.lastExpandedCmd != "observe" {
			t.Fatalf("lastExpandedCmd=%q", s.lastExpandedCmd)
		}
		if len(s.replHistory) != 1 || s.replHistory[0] != "r → observe" {
			t.Fatalf("history=%v", s.replHistory)
		}
	})

	t.Run("caps history length", func(t *testing.T) {
		t.Parallel()
		s := &Session{}
		total := maxReplHistory + 5
		for i := 0; i < total; i++ {
			line := fmt.Sprintf("c%d", i)
			s.recordReplSuccess(line, line, false)
		}
		if len(s.replHistory) != maxReplHistory {
			t.Fatalf("len=%d want %d", len(s.replHistory), maxReplHistory)
		}
		wantFirst := fmt.Sprintf("c%d", total-maxReplHistory)
		if s.replHistory[0] != wantFirst {
			t.Fatalf("first=%q want %q", s.replHistory[0], wantFirst)
		}
	})
}

func TestRememberSolveLine(t *testing.T) {
	t.Parallel()

	t.Run("adds kubectl prefix", func(t *testing.T) {
		t.Parallel()
		s := &Session{}
		s.rememberSolveLine("get pods")
		if want := "kubectl get pods"; s.lastSolveKubectl != want {
			t.Fatalf("got %q want %q", s.lastSolveKubectl, want)
		}
	})

	t.Run("injects context flag", func(t *testing.T) {
		t.Parallel()
		s := &Session{
			Kube: &kubectl.Runner{Context: "kind-kind"},
		}
		s.rememberSolveLine("get pods -n ns")
		want := "kubectl --context=kind-kind get pods -n ns"
		if s.lastSolveKubectl != want {
			t.Fatalf("got %q want %q", s.lastSolveKubectl, want)
		}
	})
}

func TestShowHistory(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		s := &Session{Out: &buf}
		s.showHistory()
		out := buf.String()
		if !strings.Contains(out, "No commands in this session yet") {
			t.Fatalf("out=%q", out)
		}
	})

	t.Run("shows last 12 when more than 12", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		s := &Session{Out: &buf}
		for i := 1; i <= 13; i++ {
			// Use zero-padded ids so "line-01" is not a substring of "line-10".
			s.replHistory = append(s.replHistory, fmt.Sprintf("line-%02d", i))
		}
		s.showHistory()
		out := buf.String()
		if strings.Contains(out, "line-01") {
			t.Fatalf("should not include oldest line: %q", out)
		}
		if !strings.Contains(out, "line-02") || !strings.Contains(out, "line-13") {
			t.Fatalf("expected window of last 12: %q", out)
		}
	})
}
