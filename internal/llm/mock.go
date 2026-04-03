package llm

import (
	"context"
	"fmt"
	"strings"

	"podnoir/internal/scenario"
)

type Judgment string

const (
	StoneCold Judgment = "stone_cold"
	Cold      Judgment = "cold"
	Warm      Judgment = "warm"
	Hot       Judgment = "hot"
)

type AccuseResult struct {
	Judgment Judgment
	Reply    string
}

// Mock is the rule-based stand-in LLM (no network).
type Mock struct{}

func (Mock) EvaluateAccusation(ctx context.Context, def *scenario.Definition, hypothesis string) (AccuseResult, error) {
	return evaluateAccusationRule(def, hypothesis), nil
}

func (Mock) Debrief(ctx context.Context, def *scenario.Definition) (string, error) {
	return debriefStatic(def), nil
}

func evaluateAccusationRule(def *scenario.Definition, hypothesis string) AccuseResult {
	h := strings.ToLower(strings.TrimSpace(hypothesis))
	if h == "" {
		return AccuseResult{
			Judgment: StoneCold,
			Reply:    "You sent an empty theory. The case file deserves better.",
		}
	}

	hotScore := scoreHints(h, def.HotHints)
	warmScore := scoreHints(h, def.WarmHints)

	if accusationHot(def, h, hotScore) {
		return hotReply(def)
	}
	if hotScore == 1 {
		return AccuseResult{
			Judgment: Warm,
			Reply: "You're circling the right layer — tighten the theory: name the failing mechanism " +
				"(container start vs pull vs config), not just the symptom.",
		}
	}
	if warmScore >= 1 {
		return warmSymptomReply(def)
	}
	return AccuseResult{
		Judgment: Cold,
		Reply:    "Not enough signal yet. Re-read the logs and events; come back with what failed, not what you guess.",
	}
}

func accusationHot(def *scenario.Definition, h string, hotScore int) bool {
	switch def.ID {
	case scenario.Case001:
		return hotScore >= 2 || (hotScore >= 1 && strings.Contains(h, "settings.json"))
	case scenario.Case002:
		return hotScore >= 2 || (hotScore >= 1 && (strings.Contains(h, "secret") || strings.Contains(h, "signing")))
	case scenario.Case003:
		return hotScore >= 2 || (hotScore >= 1 && (strings.Contains(h, "image") || strings.Contains(h, "pull") || strings.Contains(h, "tag")))
	case scenario.Case004:
		return hotScore >= 2 || (hotScore >= 1 && (strings.Contains(h, "probe") || strings.Contains(h, "liveness") || strings.Contains(h, "8080")))
	case scenario.Case005:
		return hotScore >= 2 || (hotScore >= 1 && (strings.Contains(h, "oom") || strings.Contains(h, "memory") || strings.Contains(h, "limit") || strings.Contains(h, "shm")))
	case scenario.Case006:
		return hotScore >= 2 || (hotScore >= 1 && (strings.Contains(h, "selector") || strings.Contains(h, "endpoint") || strings.Contains(h, "service") || strings.Contains(h, "label")))
	case scenario.Case007:
		return hotScore >= 2 || (hotScore >= 1 && (strings.Contains(h, "init") || strings.Contains(h, "gate") || strings.Contains(h, "initcontainer")))
	default:
		return hotScore >= 2
	}
}

func hotReply(def *scenario.Definition) AccuseResult {
	ns := def.Namespace
	dep := def.SolveDeployment
	switch def.ID {
	case scenario.Case001:
		return AccuseResult{
			Judgment: Hot,
			Reply: "That's the shape of it. The entrypoint expects /app/config/settings.json; it isn't there. " +
				"The workload dies immediately — Kubernetes is doing its job.\n\n" +
				"When you're in solve mode: rollback to the previous ReplicaSet if this was a bad rollout " +
				fmt.Sprintf("(`kubectl rollout undo deployment/%s -n %s`), or patch the Deployment ", dep, ns) +
				"so the container start command matches a healthy spec again — debrief has JSON and strategic-merge examples.",
		}
	case scenario.Case002:
		return AccuseResult{
			Judgment: Hot,
			Reply: fmt.Sprintf(
				"You've got it — the Pod can't start because env references Secret `ledger-signing-secret` "+
					"(key signing.pem) and that Secret doesn't exist (or the key is wrong). The kubelet stops at CreateContainerConfig.\n\n"+
					"In solve mode: create the Secret, e.g.\n"+
					"  kubectl create secret generic ledger-signing-secret -n %s --from-literal=signing.pem=dev-placeholder\n"+
					"or patch the Deployment to point at the correct Secret name/key. See debrief for a full checklist.",
				ns,
			),
		}
	case scenario.Case003:
		return AccuseResult{
			Judgment: Hot,
			Reply: fmt.Sprintf(
				"Right — the node can't pull `busybox:9.99.99-noir-invalid-tag` (image/registry error). "+
					"The Pod never runs your command; you stay in ImagePullBackOff / ErrImagePull.\n\n"+
					"In solve mode, patch to a real image, e.g.\n"+
					"  kubectl set image deployment/%[1]s notifier=busybox:1.36.1 -n %[2]s\n"+
					"or `kubectl patch deployment %[1]s -n %[2]s --type=json -p='[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"busybox:1.36.1\"}]'`",
				dep, ns,
			),
		}
	case scenario.Case004:
		return AccuseResult{
			Judgment: Hot,
			Reply: fmt.Sprintf(
				"Exactly — the workload only sleeps; the liveness HTTP probe hits :8080 where nothing listens. "+
					"The kubelet kills and restarts on schedule.\n\n"+
					"In solve mode, remove or fix the probe, e.g.\n"+
					"  kubectl patch deployment %[1]s -n %[2]s --type=json "+
					`-p='[{"op":"remove","path":"/spec/template/spec/containers/0/livenessProbe"}]'`+"\n"+
					"or point httpGet at a real port/path once the app serves HTTP.",
				dep, ns,
			),
		}
	case scenario.Case005:
		return AccuseResult{
			Judgment: Hot,
			Reply: fmt.Sprintf(
				"That's it — the container fills /dev/shm past the cgroup memory cap and gets OOMKilled. "+
					"Limits are doing their job; the story is too big for the room.\n\n"+
					"In solve mode: raise memory limits, shrink the dd, or replace the start command with something sane, e.g.\n"+
					"  kubectl patch deployment %[1]s -n %[2]s --type=json "+
					`-p='[{"op":"replace","path":"/spec/template/spec/containers/0/resources/limits/memory","value":"256Mi"}]'`,
				dep, ns,
			),
		}
	case scenario.Case006:
		return AccuseResult{
			Judgment: Hot,
			Reply: fmt.Sprintf(
				"Bingo — Service selector still says `app=invoice-frontend` but Pods carry `app=gateway-api`. "+
					"Endpoints stay empty; traffic dies in the wall.\n\n"+
					"Patch the selector, e.g.\n"+
					"  kubectl patch service gateway-svc -n %[1]s --type=merge "+
					`-p '{"spec":{"selector":{"app":"gateway-api"}}}'`,
				ns,
			),
		}
	case scenario.Case007:
		return AccuseResult{
			Judgment: Hot,
			Reply: fmt.Sprintf(
				"That's the hold-up — an initContainer (`gate`) exits non-zero before the app container starts. "+
					"The main workload never gets the floor.\n\n"+
					"In solve mode: patch the Deployment so the init succeeds (e.g. `exit 0`) or remove the initContainers stanza, e.g.\n"+
					"  kubectl patch deployment %[1]s -n %[2]s --type=json "+
					`-p='[{"op":"replace","path":"/spec/template/spec/initContainers/0/command","value":["/bin/sh","-c","exit 0"]}]'`,
				dep, ns,
			),
		}
	default:
		return AccuseResult{Judgment: Hot, Reply: "That's the right root cause for this case. Use solve mode and kubectl to fix the workload."}
	}
}

func warmSymptomReply(def *scenario.Definition) AccuseResult {
	switch def.ID {
	case scenario.Case004:
		return AccuseResult{
			Judgment: Warm,
			Reply:    "Restarts with thin logs often mean the kubelet, not your code — narrow it: probes, hooks, or policy?",
		}
	case scenario.Case005:
		return AccuseResult{
			Judgment: Warm,
			Reply:    "Something's starving — good ear. Say whether it's hard limits, eviction, or growth inside the container.",
		}
	case scenario.Case006:
		return AccuseResult{
			Judgment: Warm,
			Reply:    "Traffic story — tighten it: Service? Endpoints? Or ingress two hops away?",
		}
	case scenario.Case003:
		return AccuseResult{
			Judgment: Warm,
			Reply: "Something is stuck in scheduling or pull — good instinct. Now say exactly what phase the Pod is in " +
				"and what the events claim (pull vs backoff).",
		}
	case scenario.Case007:
		return AccuseResult{
			Judgment: Warm,
			Reply:    "Something runs before the app — good ear. Say whether it's init, sidecar policy, or the main container never scheduling.",
		}
	default:
		return AccuseResult{
			Judgment: Warm,
			Reply: "Something is unhealthy, yes — but that's a symptom. What mechanism explains why the container " +
				"never gets to a running state?",
		}
	}
}

func scoreHints(text string, hints []string) int {
	n := 0
	for _, hint := range hints {
		if strings.Contains(text, strings.ToLower(hint)) {
			n++
		}
	}
	return n
}

func debriefStatic(def *scenario.Definition) string {
	ns := def.Namespace
	switch def.ID {
	case scenario.Case001:
		return strings.TrimSpace(fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────┐
│  CASE #001 — CLOSED                                         │
│  "The Overnight Shift" (mock LLM debrief)                   │
├─────────────────────────────────────────────────────────────┤
│  ROOT CAUSE                                                 │
│  Revision 2.1.0 replaced a stable start command with one    │
│  that exits: logs reference /app/config/settings.json.      │
│  There is a known-good revision (2.0.3) still in history.   │
│                                                             │
│  FIX PATHS (any one is valid)                               │
│  A) Rollback one revision:                                  │
│     kubectl rollout undo deployment/payments-worker -n %[1]s │
│  B) JSON patch (replace command by index):                   │
│     kubectl patch deployment payments-worker -n %[1]s \      │
│       --type='json' -p='[{"op":"replace","path":"/spec/template/spec/containers/0/command","value":["/bin/sh","-c","while true; do sleep 3600; done"]}]' │
│  C) Strategic merge patch (merge on container name):        │
│     kubectl patch deployment payments-worker -n %[1]s \      │
│       --type=strategic -p \                                 │
│       '{"spec":{"template":{"spec":{"containers":[{"name":"payments-worker","command":["/bin/sh","-c","while true; do sleep 3600; done"]}]}}}}' │
│                                                             │
│  WHAT TO STUDY                                              │
│  → kubectl rollout history / undo                           │
│  → JSON patch vs strategic merge (list merge keys)          │
│  → CrashLoopBackOff vs probe failures                        │
└─────────────────────────────────────────────────────────────┘`, ns))
	case scenario.Case002:
		return strings.TrimSpace(fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────┐
│  CASE #002 — CLOSED                                         │
│  "The Ghost Credential" (mock LLM debrief)                  │
├─────────────────────────────────────────────────────────────┤
│  ROOT CAUSE                                                 │
│  Deployment ledger-api references secretKeyRef               │
│  ledger-signing-secret / signing.pem which was never        │
│  created in namespace %[1]s.                                 │
│                                                             │
│  FIX PATHS (any one is valid)                               │
│  A) Create the Secret (dev placeholder):                    │
│     kubectl create secret generic ledger-signing-secret -n %[1]s \ │
│       --from-literal=signing.pem=REPLACE_ME                    │
│  B) Fix the reference (wrong name/key):                     │
│     kubectl patch deployment ledger-api -n %[1]s ...           │
│  C) Remove the env stanza only if the app can run without it │
│     (usually not for signing keys in prod stories).          │
│                                                             │
│  WHAT TO STUDY                                              │
│  → kubectl describe pod (Events: FailedCreatePodSandBox /    │
│    CreateContainerConfigError messages)                      │
│  → difference between mount and secretKeyRef failures          │
└─────────────────────────────────────────────────────────────┘`, ns))
	case scenario.Case003:
		return strings.TrimSpace(fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────┐
│  CASE #003 — CLOSED                                         │
│  "Dead Letter Harbour" (mock LLM debrief)                    │
├─────────────────────────────────────────────────────────────┤
│  ROOT CAUSE                                                 │
│  Container image busybox:9.99.99-noir-invalid-tag does not  │
│  exist (pull failed). Pod stays Pending / ImagePullBackOff.  │
│                                                             │
│  FIX PATHS (any one is valid)                               │
│  A) kubectl set image deployment/shipping-notifier \         │
│       notifier=busybox:1.36.1 -n %[1]s                       │
│  B) JSON patch image field (container index 0):              │
│     kubectl patch deployment shipping-notifier -n %[1]s \      │
│       --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"busybox:1.36.1"}]' │
│                                                             │
│  WHAT TO STUDY                                              │
│  → kubectl describe pod — Events show Failed + pull reason    │
│  → image name vs tag vs registry auth (here: typo/tag)       │
└─────────────────────────────────────────────────────────────┘`, ns))
	case scenario.Case004:
		return strings.TrimSpace(fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────┐
│  CASE #004 — CLOSED                                         │
│  "The Wrong Number"                                         │
├─────────────────────────────────────────────────────────────┤
│  ROOT CAUSE                                                 │
│  livenessProbe httpGet on :8080 but nothing listens —       │
│  kubelet restarts the container on probe failure.           │
│                                                             │
│  FIX PATHS                                                  │
│  A) Remove probe (training quick fix):                      │
│     kubectl patch deployment bedside-console -n %[1]s --type=json \ │
│       -p='[{"op":"remove","path":"/spec/template/spec/containers/0/livenessProbe"}]' │
│  B) Point probe at real HTTP once app serves it             │
│                                                             │
│  STUDY                                                      │
│  → probe fields vs process actually bound                   │
└─────────────────────────────────────────────────────────────┘`, ns))
	case scenario.Case005:
		return strings.TrimSpace(`
┌─────────────────────────────────────────────────────────────┐
│  CASE #005 — CLOSED                                         │
│  "The Thin Margin"                                          │
├─────────────────────────────────────────────────────────────┤
│  ROOT CAUSE                                                 │
│  dd fills /dev/shm past memory limit → OOMKilled.           │
│                                                             │
│  FIX PATHS                                                  │
│  A) Raise memory limits on deployment memory-witness        │
│  B) Replace start command with harmless sleep only          │
│                                                             │
│  STUDY                                                      │
│  → resources.limits vs OOMKilled / tmpfs                    │
└─────────────────────────────────────────────────────────────┘`)
	case scenario.Case006:
		return strings.TrimSpace(fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────┐
│  CASE #006 — CLOSED                                         │
│  "The Ghost Wire"                                           │
├─────────────────────────────────────────────────────────────┤
│  ROOT CAUSE                                                 │
│  Service gateway-svc selector app=invoice-frontend; Pods are  │
│  app=gateway-api — Endpoints stay empty.                    │
│                                                             │
│  FIX                                                        │
│  kubectl patch service gateway-svc -n %[1]s --type=merge \   │
│    -p '{"spec":{"selector":{"app":"gateway-api"}}}'          │
│                                                             │
│  STUDY                                                      │
│  → kubectl get endpoints vs get svc -o wide                 │
└─────────────────────────────────────────────────────────────┘`, ns))
	case scenario.Case007:
		return strings.TrimSpace(fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────┐
│  CASE #007 — CLOSED                                         │
│  "Waiting on a Witness"                                     │
├─────────────────────────────────────────────────────────────┤
│  ROOT CAUSE                                                 │
│  initContainer gate exits 1 — Pod never reaches the app     │
│  container; rollout cannot go Ready.                        │
│                                                             │
│  FIX PATHS                                                  │
│  A) Make init succeed (training):                           │
│     kubectl patch deployment witness-hold -n %[1]s --type=json \ │
│       -p='[{"op":"replace","path":"/spec/template/spec/initContainers/0/command","value":["/bin/sh","-c","exit 0"]}]' │
│  B) Remove initContainers (if story allows)                 │
│                                                             │
│  STUDY                                                      │
│  → init container lifecycle vs CrashLoop on main container   │
└─────────────────────────────────────────────────────────────┘`, ns))
	default:
		return "Case closed. (No debrief text for this scenario.)"
	}
}
