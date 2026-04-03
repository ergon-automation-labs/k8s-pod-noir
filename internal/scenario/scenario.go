package scenario

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed manifests/case001-rev1.yaml
var case001Rev1 []byte

//go:embed manifests/case001-rev2.yaml
var case001Rev2 []byte

//go:embed manifests/case002.yaml
var case002 []byte

//go:embed manifests/case003.yaml
var case003 []byte

type ID string

const (
	Case001 ID = "case-001-overnight-shift"
	Case002 ID = "case-002-ghost-credential"
	Case003 ID = "case-003-dead-letter-harbour"
)

type Definition struct {
	ID        ID
	Title     string
	Namespace string

	// FolderTease is the manila-tab line in the file cabinet menu.
	FolderTease string

	ApplySteps [][]byte

	RolloutWaitAfterFirstStep string

	HotHints  []string
	WarmHints []string

	// SolveDeployment is the main workload the player patches or rollbacks (help/debrief text).
	SolveDeployment string
}

func ByID(id ID) (*Definition, error) {
	switch id {
	case Case001:
		return &Definition{
			ID:          Case001,
			Title:       "The Overnight Shift",
			Namespace:   "pod-noir",
			FolderTease: "Graveyard deploy — the payments worker goes dark",
			ApplySteps: [][]byte{
				case001Rev1,
				case001Rev2,
			},
			RolloutWaitAfterFirstStep: "payments-worker",
			SolveDeployment:           "payments-worker",
			HotHints: []string{
				"settings.json",
				"config",
				"entrypoint",
				"/app/config",
				"missing file",
				"start.sh",
			},
			WarmHints: []string{
				"crashloop",
				"crash",
				"restart",
				"oom",
				"2.1.0",
				"deploy",
			},
		}, nil
	case Case002:
		return &Definition{
			ID:                        Case002,
			Title:                     "The Ghost Credential",
			Namespace:                 "pod-noir",
			FolderTease:               "Cutover clean, API never Ready — a name on paper nobody filed",
			ApplySteps:                [][]byte{case002},
			RolloutWaitAfterFirstStep: "",
			SolveDeployment:           "ledger-api",
			HotHints: []string{
				"secret",
				"secretkeyref",
				"ledger-signing",
				"signing.pem",
				"createcontainerconfig",
				"env",
				"credential",
			},
			WarmHints: []string{
				"pending",
				"pod",
				"crash",
				"deploy",
			},
		}, nil
	case Case003:
		return &Definition{
			ID:                        Case003,
			Title:                     "Dead Letter Harbour",
			Namespace:                 "pod-noir",
			FolderTease:               "YAML reads like poetry; the harbor won't take the ship",
			ApplySteps:                [][]byte{case003},
			RolloutWaitAfterFirstStep: "",
			SolveDeployment:           "shipping-notifier",
			HotHints: []string{
				"image",
				"pull",
				"imagepullbackoff",
				"errimagepull",
				"busybox",
				"invalid",
				"tag",
			},
			WarmHints: []string{
				"registry",
				"pending",
				"crash",
				"deploy",
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown scenario %q", id)
	}
}

func List() []ID {
	return []ID{Case001, Case002, Case003}
}

// boxContentWidth is inner text width between "│  " and "│" (full row = 63 chars).
const boxContentWidth = 59

func boxRow(b *strings.Builder, text string) {
	if len(text) > boxContentWidth {
		text = text[:boxContentWidth-3] + "..."
	}
	fmt.Fprintf(b, "│  %s%s│\n", text, strings.Repeat(" ", boxContentWidth-len(text)))
}

func (d *Definition) Briefing(detective string) string {
	var b strings.Builder
	switch d.ID {
	case Case001:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY ~ wire room copy, training floor")
		boxRow(&b, fmt.Sprintf(`CASE FILE 001 — "%s"`, d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
		boxRow(&b, "Rain on the glass. Someone pushed a button they trusted.")
		boxRow(&b, "")
		boxRow(&b, "Client.. Neon Ledger Financial, uptown stack")
		boxRow(&b, "Call.... 03:47 — voice flat, tired of holding")
		boxRow(&b, `Says.... "payments-worker went quiet after the deploy."`)
		boxRow(&b, "")
		boxRow(&b, `Handwritten margin (D.): "Observe before you accuse."`)
		boxRow(&b, "")
	case Case002:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY ~ wire room copy, training floor")
		boxRow(&b, fmt.Sprintf(`CASE FILE 002 — "%s"`, d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
		boxRow(&b, "Morning light like cheap bourbon. Cutover was supposed to")
		boxRow(&b, "be clean — silence where Ready should be.")
		boxRow(&b, "")
		boxRow(&b, "Client.. Harbor & Ledger Trust Co.")
		boxRow(&b, "Call.... 11:18 — ops lead, jaw tight")
		boxRow(&b, `Says.... "ledger-api never comes up; boxes look fine."`)
		boxRow(&b, "")
		boxRow(&b, `Handwritten margin (D.): "Read the events, not the hope."`)
		boxRow(&b, "")
	case Case003:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY ~ wire room copy, training floor")
		boxRow(&b, fmt.Sprintf(`CASE FILE 003 — "%s"`, d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
		boxRow(&b, "Fog off the docks. Dev swears the manifest is scripture.")
		boxRow(&b, "The node has its own religion: pull, or do not run.")
		boxRow(&b, "")
		boxRow(&b, "Client.. Strandline Logistics")
		boxRow(&b, "Call.... 06:02 — too early for metaphor, late for tea")
		boxRow(&b, `Says.... "notifier hung up; YAML's pretty — pod isn't."`)
		boxRow(&b, "")
		boxRow(&b, `Handwritten margin (D.): "Phase lies less than people."`)
		boxRow(&b, "")
	default:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY")
		boxRow(&b, fmt.Sprintf("OPEN FILE — %s", d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
	}
	boxRow(&b, fmt.Sprintf("Namespace: %s", d.Namespace))
	boxRow(&b, fmt.Sprintf("Assigned: %s", detective))
	boxRow(&b, "")
	fmt.Fprintf(&b, "└─────────────────────────────────────────────────────────────┘\n")
	return b.String()
}

// CurtainLine is a short atmospheric beat after the formal briefing.
func (d *Definition) CurtainLine() string {
	switch d.ID {
	case Case001:
		return "You thumb the edge of the folder. Somewhere a revision number ticks upward like a second hand."
	case Case002:
		return "The coffee ring on the form could be a halo or a warning. You open the namespace like a drawer."
	case Case003:
		return "Harbor lights don't lie — they just don't tell you what's in the container until you look."
	default:
		return "The paperclip is bent. The cluster is honest in its own language."
	}
}
