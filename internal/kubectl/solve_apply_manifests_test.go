package kubectl

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestEnsureSolveApplyManifests_wrongMetadataNamespace(t *testing.T) {
	ns := "pod-noir"
	yml := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: x\n  namespace: kube-system\n"
	files := map[string][]byte{"/tmp/pn-test.yaml": []byte(yml)}
	rf := func(p string) ([]byte, error) {
		b, ok := files[p]
		if !ok {
			return nil, fmt.Errorf("not found: %s", p)
		}
		return b, nil
	}
	line := "kubectl apply -f /tmp/pn-test.yaml -n " + ns
	if err := ensureSolveApplyManifests(line, ns, "/", rf); err == nil {
		t.Fatal("expected error for wrong metadata.namespace")
	}
}

func TestEnsureSolveApplyManifests_okImplicitNamespace(t *testing.T) {
	ns := "pod-noir"
	yml := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: x\nspec:\n  containers: []\n"
	base := "/work"
	line := "kubectl apply -f manifest.yaml -n " + ns
	rf := func(p string) ([]byte, error) {
		want := filepath.Join(base, "manifest.yaml")
		if p != want {
			return nil, fmt.Errorf("unexpected path %q want %q", p, want)
		}
		return []byte(yml), nil
	}
	if err := ensureSolveApplyManifests(line, ns, base, rf); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureSolveApplyManifests_clusterKindBlocked(t *testing.T) {
	ns := "pod-noir"
	yml := "apiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole\nmetadata:\n  name: x\n"
	files := map[string][]byte{"/tmp/cr.yaml": []byte(yml)}
	rf := func(p string) ([]byte, error) { return files[p], nil }
	line := "kubectl apply -f /tmp/cr.yaml -n " + ns
	if err := ensureSolveApplyManifests(line, ns, "/", rf); err == nil {
		t.Fatal("expected block ClusterRole")
	}
}

func TestEnsureSolveApplyManifests_listItems(t *testing.T) {
	ns := "pod-noir"
	yml := `apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: a
    namespace: default
`
	files := map[string][]byte{"/tmp/list.yaml": []byte(yml)}
	rf := func(p string) ([]byte, error) { return files[p], nil }
	line := "kubectl apply -f /tmp/list.yaml -n " + ns
	if err := ensureSolveApplyManifests(line, ns, "/", rf); err == nil {
		t.Fatal("expected error: item namespace default != case ns")
	}
}

func TestEnsureSolveApplyManifests_stdinBlocked(t *testing.T) {
	ns := "pod-noir"
	line := "kubectl apply -f - -n " + ns
	if err := ensureSolveApplyManifests(line, ns, "/", func(string) ([]byte, error) {
		t.Fatal("should not read")
		return nil, nil
	}); err == nil {
		t.Fatal("expected stdin blocked")
	}
}

func TestEnsureSolveApplyManifests_listItems_okWhenNSMatches(t *testing.T) {
	ns := "pod-noir"
	yml := `apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: a
    namespace: pod-noir
`
	files := map[string][]byte{"/tmp/list.yaml": []byte(yml)}
	rf := func(p string) ([]byte, error) { return files[p], nil }
	line := "kubectl apply -f /tmp/list.yaml -n " + ns
	if err := ensureSolveApplyManifests(line, ns, "/", rf); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureSolveApplyManifests_multipleFilenameFlags(t *testing.T) {
	ns := "pod-noir"
	a := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: a\n"
	b := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: b\n"
	files := map[string][]byte{
		"/tmp/a.yaml": []byte(a),
		"/tmp/b.yaml": []byte(b),
	}
	rf := func(p string) ([]byte, error) { return files[p], nil }
	line := "kubectl apply -f /tmp/a.yaml -f /tmp/b.yaml -n " + ns
	if err := ensureSolveApplyManifests(line, ns, "/", rf); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureSolveApplyManifests_nonMutatingSkipped(t *testing.T) {
	ns := "pod-noir"
	line := "kubectl get pods -n " + ns
	if err := ensureSolveApplyManifests(line, ns, "/", func(string) ([]byte, error) {
		t.Fatal("should not read file")
		return nil, nil
	}); err != nil {
		t.Fatal(err)
	}
}
