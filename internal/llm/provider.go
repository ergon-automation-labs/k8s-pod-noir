package llm

import (
	"context"

	"podnoir/internal/scenario"
)

// Provider evaluates accusations and generates debrief text.
// Optional: ContactWirer (see contact_wire.go) adds LLM wire-room hints; HTTP implements it.
type Provider interface {
	EvaluateAccusation(ctx context.Context, def *scenario.Definition, hypothesis string) (AccuseResult, error)
	Debrief(ctx context.Context, def *scenario.Definition) (string, error)
}
