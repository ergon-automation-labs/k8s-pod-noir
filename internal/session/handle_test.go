package session

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"podnoir/internal/events"
	"podnoir/internal/kubectl"
	"podnoir/internal/llm"
	"podnoir/internal/scenario"
	"podnoir/internal/store"
)

func newTestSession(t *testing.T) (*Session, *bytes.Buffer) {
	t.Helper()
	return newTestSessionKube(t, &fakeKube{})
}

func newTestSessionKube(t *testing.T, kube kubectl.Kube) (*Session, *bytes.Buffer) {
	t.Helper()
	return newTestSessionScenario(t, kube, &scenario.Definition{ID: "custom-unknown", Title: "Repl Test", Namespace: fakeTestNS}, fakeTestNS)
}

// newTestSessionScenario builds a session with a real scenario definition and namespace (e.g. Case001 + pod-noir).
func newTestSessionScenario(t *testing.T, kube kubectl.Kube, def *scenario.Definition, ns string) (*Session, *bytes.Buffer) {
	t.Helper()
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	id, err := st.StartSession(ctx, string(def.ID), "detective")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	s := &Session{
		Out:       &buf,
		ctx:       ctx,
		cancel:    cancel,
		Kube:      kube,
		Store:     st,
		SessID:    id,
		Def:       def,
		NS:        ns,
		Detective: "D",
		LLM:       llm.Mock{},
		Emitter:   events.NopEmitter{},
	}
	return s, &buf
}

func TestHandleSolveLocked(t *testing.T) {
	t.Parallel()
	s, _ := newTestSession(t)
	err := s.handle("solve")
	if err == nil {
		t.Fatal("expected error when not HOT")
	}
	if !strings.Contains(err.Error(), "HOT") {
		t.Fatalf("err=%v", err)
	}
}

func TestHandleSolveEntersModeWhenHot(t *testing.T) {
	t.Parallel()
	s, buf := newTestSession(t)
	s.hotAccusation = true
	if err := s.handle("solve"); err != nil {
		t.Fatal(err)
	}
	if !s.solveMode {
		t.Fatal("solveMode should be true")
	}
	out := buf.String()
	if !strings.Contains(out, "Solve mode") {
		t.Fatalf("out=%q", out)
	}
}

func TestHandleSolveExit(t *testing.T) {
	t.Parallel()
	s, buf := newTestSession(t)
	s.solveMode = true
	if err := s.handle("exit"); err != nil {
		t.Fatal(err)
	}
	if s.solveMode {
		t.Fatal("solveMode should be false after exit")
	}
	if !strings.Contains(buf.String(), "Left solve mode") {
		t.Fatalf("out=%q", buf.String())
	}
}

func TestHandleHelp(t *testing.T) {
	t.Parallel()
	s, buf := newTestSession(t)
	if err := s.handle("help"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Commands:") {
		t.Fatalf("out=%q", buf.String())
	}
}

func TestHandleUnknown(t *testing.T) {
	t.Parallel()
	s, buf := newTestSession(t)
	if err := s.handle("not-a-real-command"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Unknown command") {
		t.Fatalf("out=%q", buf.String())
	}
}

func TestHandleHistEmpty(t *testing.T) {
	t.Parallel()
	s, buf := newTestSession(t)
	if err := s.handle("hist"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No commands in this session yet") {
		t.Fatalf("out=%q", buf.String())
	}
}

func TestHandleHintRoster(t *testing.T) {
	t.Parallel()
	s, buf := newTestSession(t)
	if err := s.handle("hint"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "WIRE ROOM") {
		t.Fatalf("out=%q", buf.String())
	}
}

func TestRunREPLQuit(t *testing.T) {
	t.Parallel()
	s, buf := newTestSession(t)
	s.In = strings.NewReader("quit\n")
	if err := s.RunREPL(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "wire room") {
		t.Fatalf("expected briefing banner; out=%q", out)
	}
}
