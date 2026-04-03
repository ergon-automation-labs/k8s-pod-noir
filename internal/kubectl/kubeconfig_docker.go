package kubectl

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ResolveKubeconfigPath returns a kubeconfig path suitable for kubectl.
// When POD_NOIR_KUBE_IN_DOCKER is true (typical for docker compose run), it copies
// the kubeconfig to a temp file and:
//   - rewrites API server URLs from 127.0.0.1 / localhost to POD_NOIR_KUBE_API_HOST
//     (default host.docker.internal) so the API is reachable from inside the container;
//   - unless POD_NOIR_KUBE_TLS_INSECURE=false, sets insecure-skip-tls-verify on those
//     clusters and drops embedded CA data. Local clusters (Rancher Desktop, etc.) issue
//     certs whose SANs include localhost but not host.docker.internal, so TLS would
//     otherwise fail after rewriting the hostname.
//
// The cleanup function removes the temp file when non-nil.
func ResolveKubeconfigPath() (path string, cleanup func(), err error) {
	cleanup = func() {}
	p := strings.TrimSpace(os.Getenv("KUBECONFIG"))
	if p == "" {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", cleanup, herr
		}
		p = filepath.Join(home, ".kube", "config")
	}

	inDocker := os.Getenv("POD_NOIR_KUBE_IN_DOCKER") == "true" || os.Getenv("POD_NOIR_KUBE_IN_DOCKER") == "1"
	if !inDocker {
		return p, cleanup, nil
	}

	b, err := os.ReadFile(p)
	if err != nil {
		return "", cleanup, err
	}
	host := strings.TrimSpace(os.Getenv("POD_NOIR_KUBE_API_HOST"))
	if host == "" {
		host = "host.docker.internal"
	}
	// Default true when unset: local Docker + rewritten host almost always needs skip-verify.
	tlsInsecure := strings.TrimSpace(os.Getenv("POD_NOIR_KUBE_TLS_INSECURE")) != "false"

	out, err := patchKubeconfigForDocker(b, host, tlsInsecure)
	if err != nil {
		return "", cleanup, err
	}

	f, err := os.CreateTemp("", "pod-noir-kubeconfig-*.yaml")
	if err != nil {
		return "", cleanup, err
	}
	tmpPath := f.Name()
	if _, err := f.Write(out); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return "", cleanup, err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", cleanup, err
	}

	cleanup = func() { _ = os.Remove(tmpPath) }
	return tmpPath, cleanup, nil
}

// patchKubeconfigForDocker rewrites cluster server URLs and optionally enables insecure TLS
// for Docker-to-host API access.
func patchKubeconfigForDocker(b []byte, host string, tlsInsecure bool) ([]byte, error) {
	var root map[string]interface{}
	if err := yaml.Unmarshal(b, &root); err != nil {
		return nil, err
	}
	clusters, ok := root["clusters"].([]interface{})
	if !ok {
		return rewriteLocalAPIServerBytes(b, host), nil
	}
	for _, item := range clusters {
		cm, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		cl, ok := cm["cluster"].(map[string]interface{})
		if !ok {
			continue
		}
		srv, _ := cl["server"].(string)
		cl["server"] = rewriteServerURL(srv, host)
		if tlsInsecure {
			cl["insecure-skip-tls-verify"] = true
			delete(cl, "certificate-authority-data")
			delete(cl, "certificate-authority")
		}
	}
	return yaml.Marshal(root)
}

func rewriteServerURL(server, host string) string {
	if server == "" {
		return server
	}
	replacements := [][2]string{
		{"https://127.0.0.1:6443", "https://" + host + ":6443"},
		{"https://localhost:6443", "https://" + host + ":6443"},
		{"http://127.0.0.1:6443", "http://" + host + ":6443"},
		{"http://localhost:6443", "http://" + host + ":6443"},
	}
	s := server
	for _, pair := range replacements {
		s = strings.ReplaceAll(s, pair[0], pair[1])
	}
	return s
}

// rewriteLocalAPIServerBytes is a fallback when clusters[] is missing or malformed.
func rewriteLocalAPIServerBytes(b []byte, host string) []byte {
	s := string(b)
	replacements := [][2]string{
		{"https://127.0.0.1:6443", "https://" + host + ":6443"},
		{"https://localhost:6443", "https://" + host + ":6443"},
		{"http://127.0.0.1:6443", "http://" + host + ":6443"},
		{"http://localhost:6443", "http://" + host + ":6443"},
	}
	for _, pair := range replacements {
		s = strings.ReplaceAll(s, pair[0], pair[1])
	}
	return []byte(s)
}
