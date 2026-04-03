package llm

import (
	"context"

	"podnoir/internal/scenario"
)

// Provider evaluates accusations and generates debrief text.
type Provider interface {
	EvaluateAccusation(ctx context.Context, def *scenario.Definition, hypothesis string) (AccuseResult, error)
	Debrief(ctx context.Context, def *scenario.Definition) (string, error)
}
