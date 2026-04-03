package events

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

// buildBotArmyEnvelope returns a JSON-serializable map aligned with bot_army event envelopes
// (event, event_id, timestamp, source, source_node, schema_version, payload).
// The "event" string is pod_noir.<type> with dots preserved, e.g. pod_noir.session.started.
func buildBotArmyEnvelope(e Event) map[string]any {
	host, _ := os.Hostname()
	id := uuid.New().String()
	eventName := "pod_noir." + string(e.Type)

	payload := make(map[string]any, len(e.Payload)+1)
	for k, v := range e.Payload {
		payload[k] = v
	}
	payload["pod_noir_event_type"] = string(e.Type)

	return map[string]any{
		"event":          eventName,
		"event_id":       id,
		"timestamp":      e.Timestamp,
		"source":         "pod_noir",
		"source_node":    fmt.Sprintf("%s:%d", host, os.Getpid()),
		"schema_version": "1.0",
		"payload":        payload,
	}
}

func subjectWithPrefix(prefix string, t Type) string {
	p := strings.Trim(prefix, ".")
	return fmt.Sprintf("%s.%s", p, string(t))
}
