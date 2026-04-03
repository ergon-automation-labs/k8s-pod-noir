// Package version holds build metadata (overridden via -ldflags).
package version

// Version is the release or image tag (e.g. 0.2.0).
var Version = "dev"

// Commit is the git SHA at build time.
var Commit = "none"

// Summary returns one line for CLI printing.
func Summary() string {
	if Commit != "" && Commit != "none" {
		return Version + " (" + Commit + ")"
	}
	return Version
}
