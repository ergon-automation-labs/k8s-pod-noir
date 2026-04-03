package llm

import (
	"strings"
	"testing"
)

func TestParseAccuseJSON(t *testing.T) {
	raw := `{"judgment":"hot","reply":"You called it."}`
	res, err := parseAccuseJSON(raw)
	if err != nil || res.Judgment != Hot || res.Reply == "" {
		t.Fatalf("%+v %v", res, err)
	}

	fenced := "```json\n{\"judgment\":\"warm\",\"reply\":\"Close.\"}\n```"
	res, err = parseAccuseJSON(fenced)
	if err != nil || res.Judgment != Warm {
		t.Fatalf("fence: %+v %v", res, err)
	}

	withNoise := "Here you go:\n{\"judgment\":\"stone_cold\",\"reply\":\"Not even close.\"}\nThanks."
	res, err = parseAccuseJSON(withNoise)
	if err != nil || res.Judgment != StoneCold {
		t.Fatalf("noise: %+v %v", res, err)
	}

	_, err = parseAccuseJSON(`{"reply":"no judgment"}`)
	if err == nil {
		t.Fatal("expected error for missing judgment")
	}

	nested := `{"judgment":"cold","reply":"{\"nested\":true}"}`
	res, err = parseAccuseJSON(nested)
	if err != nil {
		t.Fatal(err)
	}
	if res.Judgment != Cold || !strings.Contains(res.Reply, "nested") {
		t.Fatalf("nested string: %+v", res)
	}
}

func TestClampRunes(t *testing.T) {
	s := strings.Repeat("é", 10)
	out := clampRunes(s, 5)
	if len([]rune(out)) < 5 {
		t.Fatalf("expected at least 5 runes, got %q", out)
	}
}
