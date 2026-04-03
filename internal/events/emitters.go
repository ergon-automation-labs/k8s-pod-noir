package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// StdoutEmitter writes one JSON line per event to w (default: os.Stdout).
type StdoutEmitter struct {
	W io.Writer
}

func (s StdoutEmitter) Emit(ctx context.Context, e Event) error {
	w := s.W
	if w == nil {
		w = os.Stdout
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

func (s StdoutEmitter) Close() error { return nil }

// NopEmitter drops all events.
type NopEmitter struct{}

func (NopEmitter) Emit(ctx context.Context, e Event) error { return nil }
func (NopEmitter) Close() error                            { return nil }

// NATSConnEmitter publishes over one connection:
//   - PrimaryPrefix + "." + type — raw Event JSON (same as historical pod-noir line protocol)
//   - BridgePrefix + "." + type — BotArmy-style envelope (when BridgePrefix is set)
//
// Either prefix may be empty to skip that publish.
type NATSConnEmitter struct {
	Conn          *nats.Conn
	PrimaryPrefix string
	BridgePrefix  string
}

func (n NATSConnEmitter) Emit(ctx context.Context, e Event) error {
	if n.Conn == nil {
		return nil
	}
	if n.PrimaryPrefix != "" {
		subj := subjectWithPrefix(n.PrimaryPrefix, e.Type)
		b, err := json.Marshal(e)
		if err != nil {
			return err
		}
		if err := n.Conn.Publish(subj, b); err != nil {
			return err
		}
	}
	if n.BridgePrefix != "" {
		subj := subjectWithPrefix(n.BridgePrefix, e.Type)
		env := buildBotArmyEnvelope(e)
		b, err := json.Marshal(env)
		if err != nil {
			return err
		}
		if err := n.Conn.Publish(subj, b); err != nil {
			return err
		}
	}
	return nil
}

func (n NATSConnEmitter) Close() error {
	if n.Conn != nil {
		n.Conn.Close()
	}
	return nil
}

// MultiEmitter forwards to all delegates; errors are joined from non-nil emitters.
type MultiEmitter []Emitter

func (m MultiEmitter) Emit(ctx context.Context, e Event) error {
	var first error
	for _, x := range m {
		if x == nil {
			continue
		}
		if err := x.Emit(ctx, e); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (m MultiEmitter) Close() error {
	var first error
	for _, x := range m {
		if x == nil {
			continue
		}
		if err := x.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// NewEmitterFromSettings builds stdout / NATS / bridge from POD_NOIR_* settings.
// adapter: stdout | nats | both | none
// Bridge (POD_NOIR_NATS_BRIDGE=true) duplicates each event to events.pod_noir.<type> with a BotArmy-style envelope.
// Do not publish to llm.* — bot_army_llm treats those as commands (inference).
func NewEmitterFromSettings(s SettingsInput) (Emitter, error) {
	adapter := strings.ToLower(strings.TrimSpace(s.EventsAdapter))
	if adapter == "" {
		adapter = "stdout"
	}

	bridge := s.NATSBridge
	bridgePrefix := strings.TrimSpace(s.NATSBridgeEventsPrefix)
	if bridge && bridgePrefix == "" {
		bridgePrefix = "events.pod_noir"
	}

	primaryPrefix := strings.TrimSpace(s.NATSPrefix)
	if primaryPrefix == "" {
		primaryPrefix = "pod-noir"
	}

	needPrimaryNATS := adapter == "nats" || adapter == "both"
	needBridgeNATS := bridge
	needNATS := needPrimaryNATS || needBridgeNATS

	if needNATS && strings.TrimSpace(s.NATSURL) == "" {
		return nil, fmt.Errorf("POD_NOIR_NATS_URL is required for events adapter %q and/or NATS bridge", adapter)
	}

	var nc *nats.Conn
	if needNATS {
		var err error
		nc, err = nats.Connect(s.NATSURL)
		if err != nil {
			return nil, fmt.Errorf("nats connect: %w", err)
		}
	}

	var primaryNATS string
	if needPrimaryNATS {
		primaryNATS = primaryPrefix
	}
	var bridgeNATS string
	if needBridgeNATS {
		bridgeNATS = bridgePrefix
	}

	switch adapter {
	case "stdout":
		if !needNATS {
			return StdoutEmitter{W: os.Stdout}, nil
		}
		// stdout + bridge only
		return MultiEmitter{
			StdoutEmitter{W: os.Stdout},
			NATSConnEmitter{Conn: nc, PrimaryPrefix: "", BridgePrefix: bridgeNATS},
		}, nil

	case "none", "off":
		if bridge {
			return NATSConnEmitter{Conn: nc, PrimaryPrefix: "", BridgePrefix: bridgeNATS}, nil
		}
		return NopEmitter{}, nil

	case "nats":
		return NATSConnEmitter{Conn: nc, PrimaryPrefix: primaryNATS, BridgePrefix: bridgeNATS}, nil

	case "both":
		return MultiEmitter{
			StdoutEmitter{W: os.Stdout},
			NATSConnEmitter{Conn: nc, PrimaryPrefix: primaryNATS, BridgePrefix: bridgeNATS},
		}, nil

	default:
		if nc != nil {
			nc.Close()
		}
		return nil, fmt.Errorf("unknown POD_NOIR_EVENTS_ADAPTER %q (stdout|nats|both|none)", adapter)
	}
}

// SettingsInput is the subset of settings needed to avoid an import cycle with package settings.
type SettingsInput struct {
	EventsAdapter          string
	NATSURL                string
	NATSPrefix             string
	NATSBridge             bool
	NATSBridgeEventsPrefix string
}

// BuildEvent fills timestamp if empty.
func BuildEvent(t Type, payload map[string]any) Event {
	return Event{
		Type:      t,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Payload:   payload,
	}
}

// Emit is a convenience for packages that only have a static helper.
func Emit(ctx context.Context, em Emitter, t Type, payload map[string]any) {
	if em == nil {
		em = StdoutEmitter{W: os.Stdout}
	}
	e := BuildEvent(t, payload)
	_ = em.Emit(ctx, e)
}
