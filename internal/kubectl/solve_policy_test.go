package kubectl

import "testing"

func TestEnsureSolvePolicy(t *testing.T) {
	ns := "pod-noir"
	if err := EnsureSolvePolicy("kubectl get pods -n "+ns, ns); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSolvePolicy("kubectl apply -f foo.yaml -n "+ns, ns); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSolvePolicy("kubectl apply --filename=foo.yaml --namespace="+ns, ns); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSolvePolicy("kubectl delete pod x -n "+ns, ns); err != nil {
		t.Fatal(err)
	}

	if EnsureSolvePolicy("kubectl apply -f foo.yaml", ns) == nil {
		t.Fatal("expected block apply -f without namespace in solve mode")
	}
	if EnsureSolvePolicy("kubectl create -f foo.yaml", ns) == nil {
		t.Fatal("expected block create -f without namespace")
	}
	if EnsureSolvePolicy("kubectl get pods -A", ns) == nil {
		t.Fatal("expected block -A")
	}
	if EnsureSolvePolicy("kubectl get pods --all-namespaces", ns) == nil {
		t.Fatal("expected block --all-namespaces")
	}
	if EnsureSolvePolicy("kubectl get pods -n "+ns+" -A", ns) == nil {
		t.Fatal("expected block combined -n and -A")
	}
	if EnsureSolvePolicy("kubectl delete namespace foo", ns) == nil {
		t.Fatal("expected block namespace delete")
	}
	if EnsureSolvePolicy("kubectl delete clusterrole foo", ns) == nil {
		t.Fatal("expected block clusterrole delete")
	}
	if EnsureSolvePolicy("kubectl adm upgrade", ns) == nil {
		t.Fatal("expected block adm")
	}
	if EnsureSolvePolicy("kubectl taint nodes foo bar=baz:NoSchedule", ns) == nil {
		t.Fatal("expected block taint")
	}
	if EnsureSolvePolicy("kubectl apply -k ./deploy -n "+ns, ns) == nil {
		t.Fatal("expected block -k")
	}
	if EnsureSolvePolicy("kubectl kustomize .", ns) == nil {
		t.Fatal("expected block kustomize subcommand")
	}
}

// Lowercase -a must not be treated as --all-namespaces (regression vs lowercasing the whole line).
func TestEnsureSolvePolicyLowercaseANotAllNamespaces(t *testing.T) {
	ns := "pod-noir"
	// Hypothetical / future-safe: policy only blocks uppercase -A, not unrelated -a flags.
	if err := EnsureSolvePolicy("kubectl get pods -n "+ns+" -a", ns); err != nil {
		t.Fatalf("lowercase -a should not trigger -A block: %v", err)
	}
}
