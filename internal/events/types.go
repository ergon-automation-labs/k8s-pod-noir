package events

import "context"

type Type string

const (
	SessionStarted   Type = "session.started"
	HypothesisMade   Type = "hypothesis.made"
	SessionSolved    Type = "session.solved"
	SessionAbandoned Type = "session.abandoned"
	SessionFailed    Type = "session.failed"
	ContactUnlocked  Type = "contact.unlocked"
	HintDelivered    Type = "hint.delivered"
)

type Event struct {
	Type      Type           `json:"type"`
	Timestamp string         `json:"timestamp"`
	Payload   map[string]any `json:"payload"`
}

// Emitter sends structured session events (stdout JSON, NATS, etc.).
type Emitter interface {
	Emit(ctx context.Context, e Event) error
	Close() error
}
