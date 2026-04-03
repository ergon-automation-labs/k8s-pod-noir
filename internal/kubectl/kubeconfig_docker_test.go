package kubectl

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPatchKubeconfigForDocker_serverAndTLS(t *testing.T) {
	in := `
clusters:
- cluster:
    certificate-authority-data: Q0E=
    server: https://127.0.0.1:6443
  name: test
contexts: []
current-context: ""
kind: Config
preferences: {}
users: []
`
	out, err := patchKubeconfigForDocker([]byte(in), "host.docker.internal", true)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]interface{}
	if err := yaml.Unmarshal(out, &root); err != nil {
		t.Fatal(err)
	}
	clusters := root["clusters"].([]interface{})
	cm := clusters[0].(map[string]interface{})
	cl := cm["cluster"].(map[string]interface{})
	if cl["server"] != "https://host.docker.internal:6443" {
		t.Fatalf("server: %v", cl["server"])
	}
	if cl["insecure-skip-tls-verify"] != true {
		t.Fatalf("expected insecure-skip-tls-verify")
	}
	if _, ok := cl["certificate-authority-data"]; ok {
		t.Fatal("ca data should be removed when insecure")
	}
}

func TestRewriteLocalAPIServerBytes(t *testing.T) {
	in := []byte(`server: https://127.0.0.1:6443`)
	out := string(rewriteLocalAPIServerBytes(in, "host.docker.internal"))
	if !strings.Contains(out, "https://host.docker.internal:6443") {
		t.Fatalf("got %q", out)
	}
}
