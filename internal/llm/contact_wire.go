package llm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"podnoir/internal/scenario"
)

// ErrUseStaticWire signals the session should print authored wire copy (mock, disabled, or HTTP fallback).
var ErrUseStaticWire = errors.New("llm: use static contact wire copy")

// ContactWirer is implemented by HTTP; session falls back to contacts.StaticWireMessage when not implemented or on ErrUseStaticWire.
type ContactWirer interface {
	ContactWire(ctx context.Context, def *scenario.Definition, role string, staticAnchor string) (string, error)
}

const contactWireSystemPrompt = `You write short "incoming wire" messages for a terminal Kubernetes noir training game.
Output plain text only — no JSON, no markdown code fences. You may use simple ASCII box lines (─ │ ┌ ┐ └ ┘ ├ ┤) like a teletype.
Keep under 32 short lines. Voice: dry, atmospheric, operational — one concrete kubectl-angled hint without solving the whole case.
Do not invent cluster state; scenario facts below are the only ground truth.`

func buildContactWirePrompt(def *scenario.Definition, role string, staticAnchor string) string {
	role = strings.TrimSpace(role)
	persona := personaForContactRole(role)
	anchor := strings.TrimSpace(staticAnchor)
	anchor = clampRunes(anchor, 1800)
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", persona)
	fmt.Fprintf(&b, "Scenario ID: %s\nTitle: %s\nNamespace: %s\n", def.ID, def.Title, def.Namespace)
	fmt.Fprintf(&b, "Main workload: deployment/%s\n", def.SolveDeployment)
	if strings.TrimSpace(def.VictoryMode) == "endpoints" && strings.TrimSpace(def.VictoryService) != "" {
		fmt.Fprintf(&b, "Victory involves Service/endpoints: %s\n", def.VictoryService)
	}
	fmt.Fprintf(&b, "Teaching keywords (do not quote verbatim): hot %v; warm %v\n\n", def.HotHints, def.WarmHints)
	fmt.Fprintf(&b, "Reference tone from the training script (paraphrase; do not copy verbatim):\n%s\n\n", anchor)
	fmt.Fprintf(&b, "Write ONE incoming wire message in character. No preamble explaining the task.\n")
	return b.String()
}

func personaForContactRole(role string) string {
	switch strings.ToLower(role) {
	case "senior_detective", "senior":
		return "You are the Senior Detective on wire two — blunt, experienced, impatient with weak theories."
	case "sysadmin":
		return "You are the Sysadmin on the basement line — nervous, practical, owes the agency a favor; kubelet and node truth."
	case "network_engineer", "network":
		return "You are the Network Engineer on the trunk line — speaks in metaphors (wires, switchboards, egress); precise about Services, endpoints, policies."
	case "archivist":
		return "You are the Archivist in the stacks — dry, precise, references case patterns and filing."
	default:
		return "You are a noir contact on a wire line helping a detective learn Kubernetes."
	}
}

func (h *HTTP) ContactWire(ctx context.Context, def *scenario.Definition, role string, staticAnchor string) (string, error) {
	if def == nil {
		return "", fmt.Errorf("nil scenario")
	}
	if !h.contactWire {
		return "", ErrUseStaticWire
	}
	prompt := buildContactWirePrompt(def, role, staticAnchor)
	text, err := h.complete(ctx, prompt, modeContactWire)
	if err != nil && h.fallback {
		fmt.Fprintf(os.Stderr, "pod-noir: LLM HTTP error (contact wire): %v — using static wire copy\n", err)
		return "", ErrUseStaticWire
	}
	if err != nil {
		return "", err
	}
	out := strings.TrimSpace(text)
	out = stripMarkdownFence(out)
	if out == "" && h.fallback {
		return "", ErrUseStaticWire
	}
	out = clampRunes(out, 8000)
	return out, nil
}
