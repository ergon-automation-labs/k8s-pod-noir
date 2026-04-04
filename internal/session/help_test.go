package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplHelpGolden(t *testing.T) {
	goldenPath := filepath.Join("testdata", "repl_help.golden")
	if os.Getenv("POD_NOIR_UPDATE_REPL_HELP_GOLDEN") == "1" {
		s, buf := newTestSession(t)
		if err := s.handle("help"); err != nil {
			t.Fatal(err)
		}
		body := strings.TrimSuffix(buf.String(), "\n")
		if err := os.WriteFile(goldenPath, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", goldenPath)
		return
	}

	t.Parallel()

	wantRaw, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSpace(string(wantRaw)) + "\n"

	s, buf := newTestSession(t)
	if err := s.handle("help"); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != want {
		t.Fatalf("help output differs from %s (set POD_NOIR_UPDATE_REPL_HELP_GOLDEN=1 to refresh)\n--- got (%d bytes):\n%s\n--- want (%d bytes):\n%s",
			goldenPath, len(got), got, len(want), want)
	}
}

func TestReplHelpEmbeddedMatchesFile(t *testing.T) {
	t.Parallel()
	goldenPath := filepath.Join("testdata", "repl_help.golden")
	fileBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if replHelpText() != strings.TrimSpace(string(fileBytes)) {
		t.Fatal("replHelpText() must match testdata/repl_help.golden (source of truth for embed)")
	}
}
