package llm

import (
	"fmt"

	"podnoir/internal/settings"
)

// NewFromSettings returns mock or HTTP-backed LLM from POD_NOIR_LLM_* env.
func NewFromSettings(s settings.Settings) (Provider, error) {
	switch s.LLMProvider {
	case "", "mock":
		return Mock{}, nil
	case "anthropic", "openai", "ollama":
		return NewHTTP(s)
	default:
		return nil, fmt.Errorf("unknown POD_NOIR_LLM_PROVIDER %q (mock|anthropic|openai|ollama)", s.LLMProvider)
	}
}
