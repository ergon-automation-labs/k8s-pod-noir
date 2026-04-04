package settings

import (
	"os"
	"strings"
)

// Settings holds runtime integration options (see FromEnv). Keys use prefix POD_NOIR_.
type Settings struct {
	// EventsAdapter: stdout (default), nats, both (stdout+nats), none.
	EventsAdapter string
	NATSURL       string
	NATSPrefix    string
	// NATSBridge publishes a second copy with a BotArmy-style envelope to events.pod_noir.<type>
	// (for consumers alongside bot_army_llm — do not use llm.*; that triggers inference).
	NATSBridge             bool
	NATSBridgeEventsPrefix string

	LLMProvider string // mock | anthropic | openai | ollama
	LLMAPIKey   string
	LLMModel    string
	LLMBaseURL  string
	// LLMFallbackMock uses rule-based mock when HTTP LLM returns an error.
	LLMFallbackMock bool
	// LLMRepairAccuse sends one follow-up when accusation JSON is unparseable (HTTP providers).
	LLMRepairAccuse bool
	// LLMContactWire enables HTTP LLM to generate wire-room hint messages (fallback to static copy on error if LLMFallbackMock).
	LLMContactWire bool
}

// FromEnv reads POD_NOIR_* variables.
func FromEnv() Settings {
	s := Settings{
		EventsAdapter:          strings.ToLower(strings.TrimSpace(getenvDefault("POD_NOIR_EVENTS_ADAPTER", "stdout"))),
		NATSURL:                strings.TrimSpace(os.Getenv("POD_NOIR_NATS_URL")),
		NATSPrefix:             strings.TrimSpace(getenvDefault("POD_NOIR_NATS_SUBJECT_PREFIX", "pod-noir")),
		LLMProvider:            strings.ToLower(strings.TrimSpace(getenvDefault("POD_NOIR_LLM_PROVIDER", "mock"))),
		LLMAPIKey:              strings.TrimSpace(os.Getenv("POD_NOIR_LLM_API_KEY")),
		LLMModel:               strings.TrimSpace(os.Getenv("POD_NOIR_LLM_MODEL")),
		LLMBaseURL:             strings.TrimSpace(os.Getenv("POD_NOIR_LLM_BASE_URL")),
		LLMFallbackMock:        getenvDefault("POD_NOIR_LLM_FALLBACK_MOCK", "true") != "false",
		LLMRepairAccuse:        getenvDefault("POD_NOIR_LLM_REPAIR", "true") != "false",
		LLMContactWire:         getenvDefault("POD_NOIR_LLM_CONTACT_WIRE", "true") != "false",
		NATSBridge:             getenvDefault("POD_NOIR_NATS_BRIDGE", "false") == "true",
		NATSBridgeEventsPrefix: strings.TrimSpace(getenvDefault("POD_NOIR_NATS_BRIDGE_EVENTS_PREFIX", "events.pod_noir")),
	}
	if s.NATSPrefix == "" {
		s.NATSPrefix = "pod-noir"
	}
	if s.NATSBridgeEventsPrefix == "" {
		s.NATSBridgeEventsPrefix = "events.pod_noir"
	}
	return s
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
