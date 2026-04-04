package session

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

const fakeTestNS = "ns-test"

func TestHandleObserveFakeKube(t *testing.T) {
	t.Parallel()
	k := &fakeKube{run: func(_ context.Context, args []string) ([]byte, error) {
		if len(args) >= 6 && args[0] == "get" && args[1] == "pods" && args[3] == fakeTestNS {
			return []byte("NAME   READY\npod-a  1/1\n"), nil
		}
		if len(args) >= 5 && args[0] == "get" && args[1] == "events" && args[3] == fakeTestNS {
			return []byte("LAST    TYPE\n1m      Normal\n"), nil
		}
		return nil, fmt.Errorf("unexpected kubectl: %v", args)
	}}
	s, buf := newTestSessionKube(t, k)
	s.Def.FieldNoteAfterObserve = "Field note: trust nothing on the first pass."

	if err := s.handle("observe"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, sub := range []string{
		"FIELD NOTES — Repl Test",
		fakeTestNS,
		"NAME   READY",
		"Recent events:",
		"Field note: trust nothing",
	} {
		if !strings.Contains(out, sub) {
			t.Fatalf("missing %q in:\n%s", sub, out)
		}
	}
}

func TestHandleTraceDeploymentFakeKube(t *testing.T) {
	t.Parallel()
	k := &fakeKube{run: func(_ context.Context, args []string) ([]byte, error) {
		// get pod <name> -n <ns> -o jsonpath…
		if len(args) >= 3 && args[0] == "get" && args[1] == "pod" && args[2] == "gateway" {
			return nil, fmt.Errorf("NotFound")
		}
		// get deploy -n <ns> <name> -o jsonpath…
		if len(args) >= 6 && args[0] == "get" && args[1] == "deploy" && args[4] == "gateway" {
			return []byte(`gateway  image: nginx:1.25
`), nil
		}
		if len(args) >= 5 && args[0] == "rollout" && args[1] == "history" && args[2] == "deployment/gateway" {
			return []byte("deployment gateway\n1        <none>\n"), nil
		}
		return nil, fmt.Errorf("unexpected kubectl: %v", args)
	}}
	s, buf := newTestSessionKube(t, k)
	if err := s.handle("trace gateway"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, sub := range []string{
		"EVIDENCE — trace deployment/gateway",
		"gateway  image: nginx:1.25",
		"Rollout history:",
	} {
		if !strings.Contains(out, sub) {
			t.Fatalf("missing %q in:\n%s", sub, out)
		}
	}
}

func TestHandleTracePodReplicaSetFakeKube(t *testing.T) {
	t.Parallel()
	jpRS := `jsonpath={.metadata.ownerReferences[?(@.kind=="ReplicaSet")].name}`
	jpDep := `jsonpath={.metadata.ownerReferences[?(@.kind=="Deployment")].name}`
	k := &fakeKube{run: func(_ context.Context, args []string) ([]byte, error) {
		if len(args) >= 7 && args[0] == "get" && args[1] == "pod" && args[2] == "witness" && args[len(args)-1] == jpRS {
			return []byte("rs-77\n"), nil
		}
		if len(args) >= 7 && args[0] == "get" && args[1] == "rs" && args[2] == "rs-77" && args[len(args)-1] == jpDep {
			return []byte("payments-worker\n"), nil
		}
		if len(args) >= 5 && args[0] == "rollout" && args[1] == "history" && args[2] == "deployment/payments-worker" {
			return []byte("deployment payments-worker\n1        <none>\n"), nil
		}
		return nil, fmt.Errorf("unexpected kubectl: %v", args)
	}}
	s, buf := newTestSessionKube(t, k)
	if err := s.handle("trace witness"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, sub := range []string{
		"EVIDENCE — trace pod/witness",
		"ReplicaSet: rs-77",
		"Deployment: payments-worker",
		"Rollout history:",
	} {
		if !strings.Contains(out, sub) {
			t.Fatalf("missing %q in:\n%s", sub, out)
		}
	}
}
