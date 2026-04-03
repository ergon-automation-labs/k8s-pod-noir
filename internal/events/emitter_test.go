package events

import "testing"

func TestSubjectWithPrefix(t *testing.T) {
	got := subjectWithPrefix("pod-noir", SessionStarted)
	want := "pod-noir.session.started"
	if got != want {
		t.Fatalf("subject: got %q want %q", got, want)
	}
}

func TestBuildBotArmyEnvelope_shape(t *testing.T) {
	e := BuildEvent(SessionStarted, map[string]any{"scenario_id": "case-001"})
	m := buildBotArmyEnvelope(e)
	if m["event"] != "pod_noir.session.started" {
		t.Fatalf("event: %v", m["event"])
	}
	if m["source"] != "pod_noir" {
		t.Fatalf("source: %v", m["source"])
	}
	p := m["payload"].(map[string]any)
	if p["scenario_id"] != "case-001" {
		t.Fatalf("payload: %v", p)
	}
	if p["pod_noir_event_type"] != "session.started" {
		t.Fatalf("pod_noir_event_type: %v", p["pod_noir_event_type"])
	}
}
