package scenario

import (
	"bytes"
	"io"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEmbeddedManifestsAreValidYAML(t *testing.T) {
	for _, id := range List() {
		def, err := ByID(id)
		if err != nil {
			t.Fatalf("ByID %s: %v", id, err)
		}
		for i, step := range def.ApplySteps {
			if len(step) == 0 {
				t.Fatalf("%s step %d: empty manifest", id, i)
			}
			dec := yaml.NewDecoder(bytes.NewReader(step))
			n := 0
			for {
				var doc any
				err := dec.Decode(&doc)
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("%s apply step %d document %d: yaml: %v", id, i, n, err)
				}
				if doc == nil {
					t.Fatalf("%s apply step %d document %d: null document", id, i, n)
				}
				n++
			}
			if n == 0 {
				t.Fatalf("%s apply step %d: no yaml documents", id, i)
			}
		}
	}
}
