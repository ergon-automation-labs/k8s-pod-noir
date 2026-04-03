package kubectl

import "testing"

func TestEnsureMutatingUsesGameNamespace(t *testing.T) {
	ns := "pod-noir"
	if err := EnsureMutatingUsesGameNamespace("get pods -n "+ns, ns); err != nil {
		t.Fatalf("get should not error: %v", err)
	}
	if err := EnsureMutatingUsesGameNamespace("delete pod foo -n "+ns, ns); err != nil {
		t.Fatalf("delete with namespace: %v", err)
	}
	if err := EnsureMutatingUsesGameNamespace("kubectl delete pod foo", ns); err == nil {
		t.Fatal("expected error without namespace")
	}
	if err := EnsureMutatingUsesGameNamespace("kubectl apply -f ./manifest.yaml", ns); err != nil {
		t.Fatalf("apply with -f should allow: %v", err)
	}
	if err := EnsureMutatingUsesGameNamespace("kubectl rollout undo deployment/payments-worker -n "+ns, ns); err != nil {
		t.Fatalf("rollout undo with namespace: %v", err)
	}
}
