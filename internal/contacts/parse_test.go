package contacts

import "testing"

func TestParseHintTarget(t *testing.T) {
	tests := []struct {
		in   string
		want ID
	}{
		{"senior", SeniorDetective},
		{"SD", SeniorDetective},
		{"sysadmin", Sysadmin},
		{"basement", Sysadmin},
		{"network", NetworkEngineer},
		{"trunk", NetworkEngineer},
		{"archivist", Archivist},
	}
	for _, tt := range tests {
		got, err := ParseHintTarget(tt.in)
		if err != nil {
			t.Fatalf("%q: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("%q: got %q want %q", tt.in, got, tt.want)
		}
	}
	if _, err := ParseHintTarget("nope"); err == nil {
		t.Fatal("expected error")
	}
}
