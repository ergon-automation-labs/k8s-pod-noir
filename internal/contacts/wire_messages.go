package contacts

import (
	"fmt"
	"strings"

	"podnoir/internal/scenario"
)

// SysadminMessage — nervous, owes you a favor; one concrete ops angle.
func SysadminMessage(def *scenario.Definition) string {
	if def == nil {
		return sysadminDefault("")
	}
	switch def.ID {
	case scenario.Case001:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — SYSADMIN / BASEMENT LINE ]

  "Deploy says healthy history until it doesn't. Rollout undo isn't
  shame — it's triage. If the container dies before the probe even
  matters, read describe: command, args, and what file the start script
  swears exists."

  [ line tapped — stairwell echo ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case002:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — SYSADMIN / BASEMENT LINE ]

  "CreateContainerConfigError means the kubelet won't hand your binary
  the env — usually a Secret name that isn't there. I don't do drama;
  I read Events until the kubelet stops being polite."

  [ line tapped — keyring jingle ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case003:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — SYSADMIN / BASEMENT LINE ]

  "ImagePullBackOff is the registry saying *nice try*. Describe shows the
  pull error — tag typo, private repo, or a tag that never docked. Fix
  the image string before you debug code that never ran."

  [ line tapped — harbor draft ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case004:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — SYSADMIN / BASEMENT LINE ]

  "If the chart says HTTP and the process only sleeps, the kubelet isn't
  the villain — it's the timer. Probes dial a port; your app has to pick
  up, or the ward clerk keeps calling 911 on a wrong number."

  [ line tapped — someone owes the agency a coffee ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case005:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — SYSADMIN / BASEMENT LINE ]

  "OOMKilled is the node saying *budget exceeded*. Limits and requests
  aren't suggestions — they're sentences. describe pod: Last State tells
  you if the cgroup ran out of memory or something else."

  [ line tapped — insurance adjuster voicemail ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case006:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — SYSADMIN / BASEMENT LINE ]

  "Pods can be Ready and still useless — Service is a label match, not
  a vibe. I check endpoints and selectors when traffic 'mysteriously'
  dies; the app team loves to blame DNS first."

  [ line tapped — punch clock ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case007:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — SYSADMIN / BASEMENT LINE ]

  "Init runs before the app gets a seat. If init exits non-zero, your
  main container never testifies — describe lists initContainers first.
  Fix the gate, not the keynote speaker."

  [ line tapped — holding cell buzzer ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case008:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — SYSADMIN / BASEMENT LINE ]

  "Quota isn't drama — it's headcount. The scheduler reads requests like
  a bouncer list. You want more CPU on the floor, you raise the ceiling
  or shrink the suit."

  [ line tapped — rubber stamps echo downstairs ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case009:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — SYSADMIN / BASEMENT LINE ]

  "Pending PVC means the volume story never closed — describe pod shows
  FailedMount when the claim won't bind. I don't blame the app until the
  disk exists on paper."

  [ line tapped — evidence cage ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case010:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — SYSADMIN / BASEMENT LINE ]

  "When the start script needs DNS and the pod can't egress, the exit
  code looks like app failure — check NetworkPolicy and whether
  nslookup/kube-dns can leave the namespace. Policy is ops wearing a lawyer hat."

  [ line tapped — basement fan ]
─────────────────────────────────────────────────────────────`)
	default:
		return sysadminDefault(def.Title)
	}
}

func sysadminDefault(title string) string {
	t := strings.TrimSpace(title)
	if t == "" {
		t = "this case"
	}
	return strings.TrimSpace(fmt.Sprintf(`
─────────────────────────────────────────────────────────────
  [ INCOMING — SYSADMIN / BASEMENT LINE ]

  "I saw the describe cross my desk — %s. Pods lie in small print:
  restarts, probe rows, OOM lines. Read the Status block like a receipt
  before you blame the app team."

  [ line tapped — favor logged ]
─────────────────────────────────────────────────────────────`, t))
}

// NetworkEngineerMessage — metaphors, wiring layer.
func NetworkEngineerMessage(def *scenario.Definition) string {
	if def == nil {
		return networkDefault("")
	}
	switch def.ID {
	case scenario.Case001:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — NETWORK / TRUNK LINE ]

  "Rollout churn looks like plumbing from here — every restart is a
  reconnect. When the pipe's wrong it's usually not DNS; it's the binary
  never holding the socket open. Still: I trace Service → Pod when Ready
  lies."

  [ static crackles ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case002:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — NETWORK / TRUNK LINE ]

  "No packet leaves if the container never starts — Secret/env failures
  stop you before the NIC matters. Once the pod runs, then we talk
  Service and endpoints. Order of operations isn't poetry; it's wiring."

  [ she hangs up mid-metaphor ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case003:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — NETWORK / TRUNK LINE ]

  "Registry pull is upstream of your cluster network — CNI can't fix a
  tag that doesn't exist. Fix the image ref; then if packets drop, call
  me with traceroute and a prayer."

  [ foghorn in the distance ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case004:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — NETWORK / TRUNK LINE ]

  "Probes hit an IP:port — if nothing listens, the kubelet thinks the
  world ended. That's not routing; that's a socket nobody opened. Match
  probe to process, then we argue about ingress."

  [ line hums at 8080 ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case005:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — NETWORK / TRUNK LINE ]

  "OOM is the host saying *no more buffers* — not a routing table
  problem. Memory pressure can look like flaky connections later; fix the
  cgroup before you chase TCP dumps."

  [ dry laugh ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case006:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — NETWORK / TRUNK LINE ]

  "Ready replicas mean the band's in the building — the Service is still
  calling the wrong dressing room. Labels and selectors have to agree
  or the switchboard routes to nowhere. Get endpoints to show addresses."

  [ static — then a laugh you didn't ask for ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case007:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — NETWORK / TRUNK LINE ]

  "InitContainers are bouncers — main container doesn't get on stage
  until every init exits clean. No traffic story until the app container
  actually runs; don't waste a packet trace on a pod that never opened."

  [ velvet rope metaphor implied ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case008:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — NETWORK / TRUNK LINE ]

  "Quota blocks scheduling — pods never get a network identity if they
  never land on a node. Fix the budget first; empty endpoints sometimes
  mean *never scheduled*, not *bad routes*."

  [ municipal hold music ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case009:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — NETWORK / TRUNK LINE ]

  "Volume bind is storage's job first — PVC Pending means the pod may not
  even mount; without a mount, your app isn't talking to anything. CNI
  comes after the disk story checks out."

  [ paper rustle ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case010:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — NETWORK / TRUNK LINE ]

  "Policy that denies egress is a hallway with no exit — DNS can't leave,
  the API can't leave. You want names resolved, you carve a door in the
  firewall paragraph, not just hope the pod telepathically knows kube-dns."

  [ line hums — someone reads RFCs for fun ]
─────────────────────────────────────────────────────────────`)
	default:
		return networkDefault(def.Title)
	}
}

func networkDefault(title string) string {
	t := strings.TrimSpace(title)
	if t == "" {
		t = "the case"
	}
	return strings.TrimSpace(fmt.Sprintf(`
─────────────────────────────────────────────────────────────
  [ INCOMING — NETWORK / TRUNK LINE ]

  "You traced %s — good. Follow Service → Endpoints → Pod labels when
  traffic 'should' flow and doesn't. The cluster routes on what matches,
  not what the README promises."

  [ message logged — she never draws diagrams ]
─────────────────────────────────────────────────────────────`, t))
}

// ArchivistMessage — dry, case numbers, history.
func ArchivistMessage(def *scenario.Definition) string {
	if def == nil {
		return archivistDefault("")
	}
	switch def.ID {
	case scenario.Case001:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — ARCHIVIST / STACKS ]

  "Revision history is a diary — same case file, different handwriting.
  When payments-worker goes dark after a deploy, I file it under
  'operator optimism.' Compare revisions before you chase ghosts."

  [ message logged — acid-free folder ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case002:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — ARCHIVIST / STACKS ]

  "Cross-reference the manifest against what the cluster actually holds.
  Secret names are filed exactly — a typo is a missing witness. I keep
  case 002 under *credential fiction*."

  [ stamp: REFERENCE ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case003:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — ARCHIVIST / STACKS ]

  "Dock ledgers don't forgive bad tags. ImagePullBackOff repeats across
  fleets — I tag those folders *registry reality check*. The manifest
  is wishful; Events are the receipt."

  [ rubber band snaps ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case004:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — ARCHIVIST / STACKS ]

  "Probe configuration is filed with the chart revision — wrong port
  reads like a wrong patient ID. Case 004 always looks like *app broken*
  until you open the liveness stanza."

  [ chart clipped to board ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case005:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — ARCHIVIST / STACKS ]

  "OOM patterns repeat — tmpfs and limits share a drawer with *witness
  overstated memory*. I cross-match describe's Last State with requests;
  the story's usually in the cgroup column."

  [ red string on corkboard ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case006:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — ARCHIVIST / STACKS ]

  "Empty endpoints are filed under *label drift* — case 006 is the
  poster child. Service YAML and Pod labels must match like filing
  codes; one digit off and traffic sits in the lobby forever."

  [ index card: gateway-svc ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case007:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — ARCHIVIST / STACKS ]

  "InitContainers are numbered exhibits — they run in order. Case 007
  folders stall when exhibit A won't sign; the main act never makes the
  docket."

  [ court calendar rustles ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case008:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — ARCHIVIST / STACKS ]

  "ResourceQuota lives in the municipal annex — hard limits, soft hopes.
  Case 008: the witness requests more than the floor allows; scheduling
  refuses before the story starts."

  [ stamp: INSUFFICIENT CEILING ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case009:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — ARCHIVIST / STACKS ]

  "Pending PVCs are promises we never bound. I've seen case 009-shaped
  paper before — StorageClass names that don't exist on this cluster.
  The claim stays in the inbox until someone files the right class."

  [ message logged — key ring rattles ]
─────────────────────────────────────────────────────────────`)
	case scenario.Case010:
		return strings.TrimSpace(`
─────────────────────────────────────────────────────────────
  [ INCOMING — ARCHIVIST / STACKS ]

  "NetworkPolicy is filed under building code — egress rules are exits
  on the blueprint. Case 010: DNS never leaves the room; the tape never
  spools. Read policy like a deed restriction."

  [ index: lock-the-door ]
─────────────────────────────────────────────────────────────`)
	default:
		return archivistDefault(def.Title)
	}
}

func archivistDefault(title string) string {
	t := strings.TrimSpace(title)
	if t == "" {
		t = "this folder"
	}
	return strings.TrimSpace(fmt.Sprintf(`
─────────────────────────────────────────────────────────────
  [ INCOMING — ARCHIVIST / STACKS ]

  "Your dossier already knows how often you've opened %s. Patterns repeat
  — rollout, pull, quota, policy. Name the mechanism from the file, not
  from the panic in chat."

  [ message logged — no highlighter ]
─────────────────────────────────────────────────────────────`, t))
}
