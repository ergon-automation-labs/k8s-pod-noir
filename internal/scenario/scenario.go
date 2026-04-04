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

//go:embed manifests/case004.yaml
var case004 []byte

//go:embed manifests/case005.yaml
var case005 []byte

//go:embed manifests/case006.yaml
var case006 []byte

//go:embed manifests/case007.yaml
var case007 []byte

//go:embed manifests/case008.yaml
var case008 []byte

//go:embed manifests/case009.yaml
var case009 []byte

//go:embed manifests/case010.yaml
var case010 []byte

type ID string

const (
	Case001 ID = "case-001-overnight-shift"
	Case002 ID = "case-002-ghost-credential"
	Case003 ID = "case-003-dead-letter-harbour"
	Case004 ID = "case-004-wrong-number"
	Case005 ID = "case-005-thin-margin"
	Case006 ID = "case-006-ghost-wire"
	Case007 ID = "case-007-waiting-on-a-witness"
	Case008 ID = "case-008-the-red-tape-room"
	Case009 ID = "case-009-evidence-locker-blues"
	Case010 ID = "case-010-the-silent-corridor"
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

	// VictoryMode: "" or "rollout" (default) — successful rollout of SolveDeployment;
	// "endpoints" — VictoryService must have endpoint addresses.
	VictoryMode    string
	VictoryService string

	// FieldNoteAfterObserve/Examine are in-universe teaching beats (shown once).
	FieldNoteAfterObserve string
	FieldNoteAfterExamine string

	// SolveHints are precinct-safe kubectl reminders shown when entering solve mode.
	SolveHints []string
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
			VictoryMode:               "rollout",
			FieldNoteAfterObserve:     "Training note: rollout history is a confession; events are the alibi.",
			FieldNoteAfterExamine:     "Training note: describe pod tells you which container dies — and what the kubelet saw.",
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
			SolveHints: []string{
				"kubectl rollout history deployment/payments-worker -n pod-noir",
				"kubectl rollout undo deployment/payments-worker -n pod-noir",
			},
		}, nil
	case Case002:
		return &Definition{
			ID:                        Case002,
			Title:                     "The Ghost Credential",
			Namespace:                 "pod-noir",
			FolderTease:               "Cutover clean on paper — API never Ready; Secret never filed",
			ApplySteps:                [][]byte{case002},
			RolloutWaitAfterFirstStep: "",
			SolveDeployment:           "ledger-api",
			VictoryMode:               "rollout",
			FieldNoteAfterObserve:     "Training note: pods that never start often fail before your binary runs — scan events for Secret/Config.",
			FieldNoteAfterExamine:     "Training note: look for CreateContainerConfigError in Events — that's the kubelet refusing the spec.",
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
			SolveHints: []string{
				"kubectl create secret generic ledger-signing-secret -n pod-noir --from-file=signing.pem=...",
				"Confirm the deployment references the secret name your manifest expects.",
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
			VictoryMode:               "rollout",
			FieldNoteAfterObserve:     "Training note: ImagePullBackOff stays on the marquee — describe pod reads the registry's verdict.",
			FieldNoteAfterExamine:     "Training note: Failed + FailedPullImage in Events beats a handsome Deployment YAML.",
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
			SolveHints: []string{
				"kubectl set image deployment/shipping-notifier notifier=busybox:1.36.1 -n pod-noir",
				"kubectl describe pod -n pod-noir — confirm Events show pull vs start failures.",
			},
		}, nil
	case Case004:
		return &Definition{
			ID:                        Case004,
			Title:                     "The Wrong Number",
			Namespace:                 "pod-noir",
			FolderTease:               "Chart says the patient's breathing; the probe keeps calling a dead line",
			ApplySteps:                [][]byte{case004},
			RolloutWaitAfterFirstStep: "",
			SolveDeployment:           "bedside-console",
			VictoryMode:               "rollout",
			FieldNoteAfterObserve:     "Training note: CrashLoopBackOff with restarts but no app logs often means probes, not business logic.",
			FieldNoteAfterExamine:     "Training note: Unhealthy + HTTP probe failures point at the wrong port/path — compare probe to what actually listens.",
			HotHints: []string{
				"probe", "liveness", "readiness", "http", "8080", "port",
				"unhealthy",
			},
			WarmHints: []string{
				"crash", "restart", "deploy", "sleep",
			},
			SolveHints: []string{
				"kubectl patch deployment bedside-console -n pod-noir — fix liveness/readiness port or path",
				"Compare probe port to what the container actually listens on.",
			},
		}, nil
	case Case005:
		return &Definition{
			ID:                        Case005,
			Title:                     "The Thin Margin",
			Namespace:                 "pod-noir",
			FolderTease:               "Witness swears it ran yesterday; cgroup says the memory story doesn't fit",
			ApplySteps:                [][]byte{case005},
			RolloutWaitAfterFirstStep: "",
			SolveDeployment:           "memory-witness",
			VictoryMode:               "rollout",
			FieldNoteAfterObserve:     "Training note: OOMKilled is the node's edit — limits are sentences; exceeding them is contempt.",
			FieldNoteAfterExamine:     "Training note: Last State: Terminated, Reason: OOMKilled — that's the coroner's stamp, not a flake.",
			HotHints: []string{
				"oom", "oomkilled", "memory", "limit", "tmpfs", "shm",
				"cgroup",
			},
			WarmHints: []string{
				"crash", "restart", "evicted", "deploy",
			},
			SolveHints: []string{
				"kubectl describe pod -n pod-noir — OOMKilled / limits on deployment memory-witness",
				"kubectl patch deployment memory-witness -n pod-noir — raise memory limits or replace the greedy start command",
			},
		}, nil
	case Case006:
		return &Definition{
			ID:                        Case006,
			Title:                     "The Ghost Wire",
			Namespace:                 "pod-noir",
			FolderTease:               "Pods answer roll call; the Service knocks on an empty apartment",
			ApplySteps:                [][]byte{case006},
			RolloutWaitAfterFirstStep: "",
			SolveDeployment:           "gateway-api",
			VictoryMode:               "endpoints",
			VictoryService:            "gateway-svc",
			FieldNoteAfterObserve:     "Training note: a Ready Deployment with no Service traffic often means the wire — selector ↔ labels.",
			FieldNoteAfterExamine:     "Training note: kubectl get endpoints — if subsets are empty, the Service isn't talking to your Pods.",
			HotHints: []string{
				"selector", "service", "endpoints", "label",
				"gateway", "invoice",
			},
			WarmHints: []string{
				"network", "ready", "clusterip", "deploy",
			},
			SolveHints: []string{
				"kubectl get endpoints gateway-svc -n pod-noir — empty subsets mean the Service is not selecting Pods",
				"Patch gateway-svc.spec.selector to match labels on gateway-api pods (same namespace).",
			},
		}, nil
	case Case007:
		return &Definition{
			ID:                        Case007,
			Title:                     "Waiting on a Witness",
			Namespace:                 "pod-noir",
			FolderTease:               "Deployment dressed for court; an init gate won't open the door",
			ApplySteps:                [][]byte{case007},
			RolloutWaitAfterFirstStep: "",
			SolveDeployment:           "witness-hold",
			VictoryMode:               "rollout",
			FieldNoteAfterObserve:     "Training note: Init:0/1 or init crash — the main container never runs until every initContainer succeeds.",
			FieldNoteAfterExamine:     "Training note: describe pod lists init container state before app logs exist; read that block first.",
			HotHints: []string{
				"init", "initcontainer", "initializing", "podinitializing",
				"gate", "crash", "exit",
			},
			WarmHints: []string{
				"pending", "pod", "deploy", "restart", "stuck",
			},
			SolveHints: []string{
				"kubectl describe pod -n pod-noir — read initContainers status and last termination reason",
				"kubectl patch deployment witness-hold -n pod-noir — fix or remove the failing initContainer (e.g. make gate exit 0)",
			},
		}, nil
	case Case008:
		return &Definition{
			ID:                        Case008,
			Title:                     "The Red-Tape Room",
			Namespace:                 "pod-noir",
			FolderTease:               "Ledger clerk stamped insufficient ceiling — no chair for your witness",
			ApplySteps:                [][]byte{case008},
			RolloutWaitAfterFirstStep: "",
			SolveDeployment:           "ledger-queue",
			VictoryMode:               "rollout",
			FieldNoteAfterObserve:     "Training note: Pending + 'exceeded quota' in events means the namespace budget, not the app binary.",
			FieldNoteAfterExamine:     "Training note: compare Pod `resources.requests` to `kubectl describe resourcequota` in the same namespace.",
			HotHints: []string{
				"quota", "resourcequota", "exceeded", "cpu", "request", "limit",
			},
			WarmHints: []string{
				"pending", "schedule", "deploy", "namespace",
			},
			SolveHints: []string{
				"kubectl describe resourcequota -n pod-noir",
				"kubectl patch resourcequota precinct-paperwork -n pod-noir — raise limits, or patch deployment ledger-queue to lower requests",
			},
		}, nil
	case Case009:
		return &Definition{
			ID:                        Case009,
			Title:                     "Evidence Locker Blues",
			Namespace:                 "pod-noir",
			FolderTease:               "Evidence room has a lock; nobody filed the combination with storage",
			ApplySteps:                [][]byte{case009},
			RolloutWaitAfterFirstStep: "",
			SolveDeployment:           "evidence-worker",
			VictoryMode:               "rollout",
			FieldNoteAfterObserve:     "Training note: PVC Pending + Pod stuck mounting — triage the claim before blaming the container.",
			FieldNoteAfterExamine:     "Training note: Events often say `FailedMount` or `unbound PersistentVolumeClaims` — follow the PVC first.",
			HotHints: []string{
				"pvc", "persistentvolume", "storageclass", "pending", "bound",
				"mount", "volume",
			},
			WarmHints: []string{
				"storage", "claim", "disk", "deploy",
			},
			SolveHints: []string{
				"kubectl describe pvc evidence-vol -n pod-noir",
				"kubectl patch pvc evidence-vol -n pod-noir — fix storageClassName (e.g. standard or cluster default) or provision a matching PV",
			},
		}, nil
	case Case010:
		return &Definition{
			ID:                        Case010,
			Title:                     "The Silent Corridor",
			Namespace:                 "pod-noir",
			FolderTease:               "The wire hums inside the pod; the hallway to the cluster DNS is another jurisdiction",
			ApplySteps:                [][]byte{case010},
			RolloutWaitAfterFirstStep: "",
			SolveDeployment:           "tape-deck",
			VictoryMode:               "rollout",
			FieldNoteAfterObserve:     "Training note: CrashLoop with no image pull error — check startup commands, policies, and whether the pod can reach DNS or the API.",
			FieldNoteAfterExamine:     "Training note: NetworkPolicy is namespaced law — `kubectl get networkpolicy -n` shows who may talk to whom.",
			HotHints: []string{
				"networkpolicy", "egress", "firewall", "dns", "policy", "wire",
			},
			WarmHints: []string{
				"crash", "network", "connect", "timeout", "deploy",
			},
			SolveHints: []string{
				"kubectl describe networkpolicy lock-the-door -n pod-noir",
				"kubectl delete networkpolicy lock-the-door -n pod-noir — or patch egress to allow DNS (udp/53) and needed destinations",
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown scenario %q", id)
	}
}

func List() []ID {
	return []ID{
		Case001, Case002, Case003, Case004, Case005, Case006, Case007,
		Case008, Case009, Case010,
	}
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
	case Case004:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY ~ wire room copy, training floor")
		boxRow(&b, fmt.Sprintf(`CASE FILE 004 — "%s"`, d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
		boxRow(&b, "Ward clerk says the bedside console 'restarts like it owes money'.")
		boxRow(&b, "Telemetry chart is a flat line — nobody's home on 8080.")
		boxRow(&b, "")
		boxRow(&b, "Client.. Midtown General outpatient IT")
		boxRow(&b, "Call.... 22:10 — night shift, voice tight")
		boxRow(&b, `Says.... "console pod never settles; Nursing can't clear beds."`)
		boxRow(&b, "")
		boxRow(&b, `Handwritten margin (D.): "Listen for probes lying about ports."`)
		boxRow(&b, "")
	case Case005:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY ~ wire room copy, training floor")
		boxRow(&b, fmt.Sprintf(`CASE FILE 005 — "%s"`, d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
		boxRow(&b, "Insurance auditor flagged a witness workload — memory sketch")
		boxRow(&b, "says forty-eight megs; the story inside wants a cathedral.")
		boxRow(&b, "")
		boxRow(&b, "Client.. Meridian Casualty internal tools")
		boxRow(&b, "Call.... 14:40 — actuarial tempers run hot")
		boxRow(&b, `Says.... "memory-witness keeps dying; nobody touched the code."`)
		boxRow(&b, "")
		boxRow(&b, `Handwritten margin (D.): "Cgroups don't negotiate."`)
		boxRow(&b, "")
	case Case006:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY ~ wire room copy, training floor")
		boxRow(&b, fmt.Sprintf(`CASE FILE 006 — "%s"`, d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
		boxRow(&b, "Gateway team swears traffic should flow; curl from another pod")
		boxRow(&b, "gets you dial tone forever — maps don't match territory.")
		boxRow(&b, "")
		boxRow(&b, "Client.. North Harbor API guild")
		boxRow(&b, "Call.... 09:51 — polite, furious")
		boxRow(&b, `Says.... "gateway-svc is a ghost; deployments look Ready."`)
		boxRow(&b, "")
		boxRow(&b, `Handwritten margin (D.): "Follow the wire, not the README."`)
		boxRow(&b, "")
	case Case007:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY ~ wire room copy, training floor")
		boxRow(&b, fmt.Sprintf(`CASE FILE 007 — "%s"`, d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
		boxRow(&b, "Records say witness-hold should testify; the hallway outside")
		boxRow(&b, "never clears — something won't sign the release form.")
		boxRow(&b, "")
		boxRow(&b, "Client.. City clerk — digital evidence vault")
		boxRow(&b, "Call.... 16:55 — they want green before the judge gavels")
		boxRow(&b, `Says.... "Pod shows up dressed for work but never takes the stand."`)
		boxRow(&b, "")
		boxRow(&b, `Handwritten margin (D.): "Gates before testimony — check who blocks the door."`)
		boxRow(&b, "")
	case Case008:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY ~ wire room copy, training floor")
		boxRow(&b, fmt.Sprintf(`CASE FILE 008 — "%s"`, d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
		boxRow(&b, "City hall says the floor plan allows one watt of ambition;")
		boxRow(&b, "your witness filed for three. The stamp says *insufficient ceiling*.")
		boxRow(&b, "")
		boxRow(&b, "Client.. Municipal payments queue (batch overnight)")
		boxRow(&b, "Call.... 08:01 — comptroller, voice like a rubber stamp")
		boxRow(&b, `Says.... "ledger-queue never schedules; paperwork says we're fine."`)
		boxRow(&b, "")
		boxRow(&b, `Handwritten margin (D.): "Quotas are the law — read the ResourceQuota first."`)
		boxRow(&b, "")
	case Case009:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY ~ wire room copy, training floor")
		boxRow(&b, fmt.Sprintf(`CASE FILE 009 — "%s"`, d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
		boxRow(&b, "Evidence chain is immaculate on paper; the locker key is typed to a")
		boxRow(&b, "vault number that was never built. The archivist paces in the hall.")
		boxRow(&b, "")
		boxRow(&b, "Client.. County records digitization")
		boxRow(&b, "Call.... 13:22 — archivist, polite panic")
		boxRow(&b, `Says.... "evidence-worker can't mount; PVC says Pending forever."`)
		boxRow(&b, "")
		boxRow(&b, `Handwritten margin (D.): "StorageClass is the deed — no deed, no lock."`)
		boxRow(&b, "")
	case Case010:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY ~ wire room copy, training floor")
		boxRow(&b, fmt.Sprintf(`CASE FILE 010 — "%s"`, d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
		boxRow(&b, "The tape deck spins inside the interview room; the corridor to the")
		boxRow(&b, "DNS switchboard is sealed — someone posted a policy with no exits.")
		boxRow(&b, "")
		boxRow(&b, "Client.. Radio evidence archive")
		boxRow(&b, "Call.... 01:44 — night engineer, static on the line")
		boxRow(&b, `Says.... "tape-deck crashes before it spools; nothing listens on the wire."`)
		boxRow(&b, "")
		boxRow(&b, `Handwritten margin (D.): "NetworkPolicy is the building code — read egress."`)
		boxRow(&b, "")
	default:
		fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
		boxRow(&b, "THE CLUSTER AGENCY ~ wire room copy, training floor")
		boxRow(&b, fmt.Sprintf(`OPEN FILE — "%s"`, d.Title))
		fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
		boxRow(&b, "Thin routing slip — full briefing not on file yet. The")
		boxRow(&b, "namespace still answers to the same laws: observe first.")
		boxRow(&b, "")
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
		return "The coffee ring on the form could be a halo or a warning. You crack the namespace like a desk drawer — careful what was never filed."
	case Case003:
		return "Harbor lights don't lie — they just don't tell you what's in the container until you look."
	case Case004:
		return "The monitor beeps reassurance. Somewhere a probe keeps calling a number that never answers."
	case Case005:
		return "The witness statement and the cgroup verdict don't match — one of them is perjury."
	case Case006:
		return "Ready replicas hum in the back room. Out front, the switchboard still says *nobody home*."
	case Case007:
		return "The witness chair is empty, but someone in the antechamber keeps answering wrong on purpose."
	case Case008:
		return "The rubber stamp has more authority than the witness — until someone raises the ceiling."
	case Case009:
		return "The key fits a lock that was never hung on the wall — StorageClass is the deed; no deed, no door."
	case Case010:
		return "The reel turns in the dark; the hallway outside can't hear the subpoena for a name."
	default:
		return "The paperclip is bent. Whatever lives in this namespace still has to testify — start with observe."
	}
}
