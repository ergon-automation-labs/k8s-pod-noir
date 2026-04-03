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
	default:
		return SeniorDetectiveStubMessageDefault()
	}
}

func SeniorDetectiveStubMessageDefault() string {
	return `
─────────────────────────────────────────────────────────────
  [ INCOMING — SENIOR DETECTIVE ]

  "You're not stuck — you're early. Logs, trace, or a theory that isn't
  on fire yet. The cluster will talk if you ask the right questions."

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
		scenario.Case004, scenario.Case005, scenario.Case006:
		return true
	default:
		return false
	}
}
