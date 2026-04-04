package llm

import (
	"strings"
	"testing"

	"podnoir/internal/scenario"
)

func TestBuildContactWirePrompt(t *testing.T) {
	def, err := scenario.ByID(scenario.Case001)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "STATIC ANCHOR LINE"
	p := buildContactWirePrompt(def, "senior_detective", anchor)
	if !strings.Contains(p, string(def.ID)) || !strings.Contains(p, anchor) {
		t.Fatalf("prompt missing id or anchor: %s", p)
	}
	if !strings.Contains(p, "Senior Detective") {
		t.Fatal("expected persona")
	}
}

func TestPersonaForContactRole(t *testing.T) {
	if !strings.Contains(personaForContactRole("sysadmin"), "basement") {
		t.Fatal(personaForContactRole("sysadmin"))
	}
}
