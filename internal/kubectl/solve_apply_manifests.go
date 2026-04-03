package kubectl

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// EnsureSolveApplyManifests checks manifest files referenced by mutating kubectl -f/--filename
// when the CLI already targets the case namespace. It blocks cluster-scoped kinds and YAML
// that sets metadata.namespace to another namespace.
func EnsureSolveApplyManifests(line, allowedNS, baseDir string) error {
	return ensureSolveApplyManifests(line, allowedNS, baseDir, os.ReadFile)
}

func ensureSolveApplyManifests(line, allowedNS, baseDir string, readFile func(string) ([]byte, error)) error {
	if allowedNS == "" {
		return nil
	}
	s := kubectlCommandBody(line)
	if !mutatingPattern.MatchString(s) || !filenameOptionPresent(s) || !namespaceSpecified(s, allowedNS) {
		return nil
	}
	paths := applyFilenamesFromArgs(s)
	if len(paths) == 0 {
		return nil
	}
	for _, p := range paths {
		if p == "-" {
			return fmt.Errorf("precinct policy: mutating apply from stdin (-f -) is blocked in solve mode")
		}
		if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
			return fmt.Errorf("precinct policy: mutating apply from URL is blocked in solve mode")
		}
		full := p
		if !filepath.IsAbs(p) {
			full = filepath.Join(baseDir, p)
		}
		raw, err := readFile(full)
		if err != nil {
			return fmt.Errorf("precinct policy: cannot read manifest %q for solve check: %w", p, err)
		}
		if err := validateSolveApplyYAML(raw, allowedNS); err != nil {
			return err
		}
	}
	return nil
}

var (
	reFilenameEquals      = regexp.MustCompile(`(?:^|\s)--filename=(\S+)`)
	reFilenameShortEq     = regexp.MustCompile(`(?:^|\s)-f=(\S+)`)
	reFilenameSpace       = regexp.MustCompile(`(?:^|\s)-f\s+(\S+)`)
	reFilenameLongSpace   = regexp.MustCompile(`(?:^|\s)--filename\s+(\S+)`)
	reFilenameSpaceDq     = regexp.MustCompile(`(?:^|\s)-f\s+"([^"]+)"`)
	reFilenameSpaceSq     = regexp.MustCompile(`(?:^|\s)-f\s+'([^']+)'`)
	reFilenameLongSpaceDq = regexp.MustCompile(`(?:^|\s)--filename\s+"([^"]+)"`)
	reFilenameLongSpaceSq = regexp.MustCompile(`(?:^|\s)--filename\s+'([^']+)'`)
)

func applyFilenamesFromArgs(args string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(p string) {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"'`)
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	for _, re := range []*regexp.Regexp{
		reFilenameSpaceDq, reFilenameLongSpaceDq, reFilenameSpaceSq, reFilenameLongSpaceSq,
		reFilenameEquals, reFilenameShortEq, reFilenameSpace, reFilenameLongSpace,
	} {
		for _, m := range re.FindAllStringSubmatch(args, -1) {
			if len(m) > 1 {
				add(m[1])
			}
		}
	}
	return out
}

func validateSolveApplyYAML(raw []byte, allowedNS string) error {
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	for {
		var root map[string]interface{}
		err := dec.Decode(&root)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("precinct policy: manifest is not valid YAML: %w", err)
		}
		if root == nil {
			continue
		}
		if err := walkManifestDoc(root, allowedNS); err != nil {
			return err
		}
	}
	return nil
}

func walkManifestDoc(root map[string]interface{}, allowedNS string) error {
	kind, _ := root["kind"].(string)
	if kind == "List" {
		raw, ok := root["items"]
		if !ok {
			return nil
		}
		items, ok := raw.([]interface{})
		if !ok {
			return nil
		}
		for _, it := range items {
			m, ok := it.(map[string]interface{})
			if !ok {
				continue
			}
			if err := walkManifestDoc(m, allowedNS); err != nil {
				return err
			}
		}
		return nil
	}
	if kind == "" {
		return nil
	}
	if _, cluster := clusterScopedSolveKinds[kind]; cluster {
		return fmt.Errorf("precinct policy: cluster-scoped kind %q cannot be applied in solve mode", kind)
	}
	meta, _ := root["metadata"].(map[string]interface{})
	if meta != nil {
		if ns, ok := meta["namespace"].(string); ok && ns != "" && ns != allowedNS {
			return fmt.Errorf("precinct policy: manifest sets metadata.namespace %q but case namespace is %q", ns, allowedNS)
		}
	}
	return nil
}

// Conservative set of built-in cluster-scoped kinds (plus common node/cluster objects).
var clusterScopedSolveKinds = map[string]struct{}{
	"Namespace":                        {},
	"ClusterRole":                      {},
	"ClusterRoleBinding":               {},
	"PersistentVolume":                 {},
	"StorageClass":                     {},
	"VolumeAttachment":                 {},
	"CSINode":                          {},
	"CSIDriver":                        {},
	"CustomResourceDefinition":         {},
	"MutatingWebhookConfiguration":     {},
	"ValidatingWebhookConfiguration":   {},
	"ValidatingAdmissionPolicy":        {},
	"ValidatingAdmissionPolicyBinding": {},
	"CertificateSigningRequest":        {},
	"Node":                             {},
	"RuntimeClass":                     {},
	"PriorityClass":                    {},
	"APIService":                       {},
	"FlowSchema":                       {},
	"PriorityLevelConfiguration":       {},
}
