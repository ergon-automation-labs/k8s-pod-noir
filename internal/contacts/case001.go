package contacts

import (
	"podnoir/internal/llm"
	"podnoir/internal/scenario"
)

// SeniorDetectiveMessage is contact copy for the hint pipeline; varies by scenario.
func SeniorDetectiveMessage(def *scenario.Definition) string {
	if def == nil {
		return SeniorDetectiveStubMessageDefault()
	}
	switch def.ID {
	case scenario.Case001:
		return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE / WIRE TWO ]

  "Listen — revision history doesn't lie, but it whispers. Line up what
  the Deployment *thought* it was shipping with what the container says
  when it quits. Undo buys you time; a patch buys you control."

  [ message logged — burn after reading, we never said that ]
─────────────────────────────────────────────────────────────`
	case scenario.Case002:
		return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE / WIRE TWO ]

  "If env references a Secret the cluster can't touch, the kubelet won't
  even let your binary apologize. Events spell it plain — 'secret not
  found' reads like a ransom note. Create the thing, or fix the name."

  [ message logged — the brass still uses paperclips ]
─────────────────────────────────────────────────────────────`
	case scenario.Case003:
		return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE / WIRE TWO ]

  "ImagePullBackOff is the harbor master saying *come back with a real
  manifesto*. Describe the pod — read the pull error like a dock warrant.
  Swap the tag for one the registry recognizes."

  [ message logged — fog lifts for people who look ]
─────────────────────────────────────────────────────────────`
	case scenario.Case004:
		return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE / WIRE TWO ]

  "If the process only sleeps but your probe dials HTTP on 8080, kubelet
  calls that dying on schedule. Fix the probe, fix the port, or prove the
  app actually listens where the chart claims."

  [ message logged — ward night, long ]
─────────────────────────────────────────────────────────────`
	case scenario.Case005:
		return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE / WIRE TWO ]

  "OOMKilled is blunt force truth — the cgroup ran out of memory budget.
  Raise limits, shrink the workload, or stop filling tmpfs like it's free."

  [ message logged — actuaries have feelings too ]
─────────────────────────────────────────────────────────────`
	case scenario.Case006:
		return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE / WIRE TWO ]

  "Endpoints empty with Ready pods? Your Service is waving at the wrong
  labels. Patch the selector until get endpoints shows addresses — then
  the traffic has somewhere to land."

  [ message logged — follow the wire ]
─────────────────────────────────────────────────────────────`
	case scenario.Case007:
		return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE / WIRE TWO ]

  "Main container never takes the stand? Look left — initContainers run
  first. If the gate fails, the story stops in the hallway. Describe the
  pod and read which box exited angry."

  [ message logged — court runs on order ]
─────────────────────────────────────────────────────────────`
	case scenario.Case008:
		return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE / WIRE TWO ]

  "When the city says you only get a sliver of CPU on the whole floor,
  your witness doesn't get a chair — they get a form. Read ResourceQuota
  like statute: hard limits, no poetry."

  [ message logged — rubber never sleeps ]
─────────────────────────────────────────────────────────────`
	case scenario.Case009:
		return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE / WIRE TWO ]

  "A PVC is a promise to a vault. If the StorageClass is fiction, the
  lock never clicks. Describe the claim before you blame the container —
  Pending means the building inspector hasn't stamped the deed."

  [ message logged — keys have to exist ]
─────────────────────────────────────────────────────────────`
	case scenario.Case010:
		return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE / WIRE TWO ]

  "Policy without egress is a room with no doors — DNS can't leave, the
  API can't leave, nothing can leave. NetworkPolicy isn't drama; it's
  jurisdiction. Read who the selector catches, then read what they're
  allowed to walk to."

  [ message logged — firewalls have addresses ]
─────────────────────────────────────────────────────────────`
	default:
		return SeniorDetectiveStubMessageDefault()
	}
}

func SeniorDetectiveStubMessageDefault() string {
	return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE ]

  "You're not stuck — you're early. The evidence is already on the wire;
  you just haven't cross-examined it. Logs, trace, or a cold theory —
  pick one and make it honest before you patch."

  [ message logged ]
─────────────────────────────────────────────────────────────`
}

// ShouldUnlockSeniorFromAccusation returns true if this accusation judgment should unlock
// the Senior Detective (wrong or weak theory — player gets another voice).
func ShouldUnlockSeniorFromAccusation(j llm.Judgment) bool {
	return j != llm.Hot
}

// ShouldUnlockSeniorFromEvidence — logs + trace (any target) earns the contact
// without requiring a failed accuse.
func ShouldUnlockSeniorFromEvidence(st *InvestigationState) bool {
	return st.SeenLogs && st.SeenTrace
}

// SeniorPath is true when the Senior Detective hint pipeline applies.
func SeniorPath(def *scenario.Definition) bool {
	if def == nil {
		return false
	}
	switch def.ID {
	case scenario.Case001, scenario.Case002, scenario.Case003,
		scenario.Case004, scenario.Case005, scenario.Case006, scenario.Case007,
		scenario.Case008, scenario.Case009, scenario.Case010:
		return true
	default:
		return false
	}
}
