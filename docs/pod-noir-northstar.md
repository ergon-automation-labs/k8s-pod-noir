# POD noir — North Star Document

> *A living document. Decisions made here constrain everything downstream.*

---

## 1. Identity

### What It Is

POD noir is a terminal-based Kubernetes learning surface where engineers investigate **embedded, scenario-driven** incidents on their own cluster, building real debugging intuition through detective work rather than tutorials.

### The Problem It Solves

Kubernetes is learned poorly. Documentation teaches you what things are. Tutorials teach you what commands to run. Neither teaches you how to *think* when something is broken at 2am and the logs are lying to you.

POD noir teaches the third thing — the investigative instinct. How to read a broken system. How to form a hypothesis before you touch anything. How to know when you've fixed the symptom but not the cause.

### Who It's For

Engineers who already know the basics — they can deploy a pod, write a manifest, navigate `kubectl` — but who haven't yet built the pattern recognition that comes from surviving real incidents. They know *what* Kubernetes is. They don't yet know how it *breaks*.

### What It Is Not

- Not a tutorial or a guided course
- Not a simulation (it runs against your actual local cluster)
- Not a certification prep tool
- Not a multiplayer game
- Not opinionated about your cluster setup beyond "you have `kubectl` access"

### Design Philosophy

**Real over simulated.** The failures happen in a real cluster. The kubectl output is real. The weirdness is real. This is not a sandbox that pretends — it is a controlled experiment on actual infrastructure.

**Detective work over instruction.** The game never tells you what to do. It surfaces evidence. It asks what you think. It rewards forming a theory before acting.

**Fiction in service of learning.** The noir world is not decoration. The narrative framing — the agency, the cases, the contacts — gives emotional stakes to technical work. You are not debugging a pod. You are solving a case. The distinction matters more than it sounds.

**Portable by default.** POD noir belongs to whoever runs it. No required infrastructure beyond a local cluster and an LLM provider. Integrations with external systems are optional and modular.

---

## 2. The Learning Model

### The Theory

Debugging intuition is built through a specific loop:

1. **Observe** — read the system without touching it
2. **Hypothesize** — commit to a theory before acting
3. **Test** — make one change, see what changes
4. **Update** — revise your model, repeat

Most engineers skip step 2. They go straight from observation to action, which means they learn commands but not reasoning. POD noir enforces the hypothesis step — you cannot attempt a fix until you have accused something. This is the core mechanic.

### The Session Loop

```
Case briefing delivered                    ← narrative sets the scene
        ↓
Investigation phase (open-ended)           ← REPL, observe/examine/trace
        ↓
Hypothesis committed (accuse <theory>)     ← LLM judges: cold/warm/hot
        ↓
Fix attempted (solve sub-mode)             ← apply your theory
        ↓
Confirmation or complication               ← did it work? what else broke?
        ↓
Debrief                                    ← root cause, what you missed, what to study
```

### What Success Looks Like

A player who has run ten POD noir cases can:
- Look at a broken pod's state and know which layer to investigate first
- Form a written hypothesis before running a single remediation command
- Recognize failure patterns they've seen before in new contexts
- Know what they don't know and where to look for it

---

## 3. The World

### Setting

You are a junior investigator at **The Cluster Agency** — a noir detective firm that specializes in infrastructure incidents. Every case is a training exercise. Real incident, real cluster, supervisor watching.

The world is stylized noir: rain-slick streets, flickering terminals, cryptic contacts who speak in half-truths. The technical content is completely real. The wrapper is fiction.

### You

You are a named detective, junior rank, studying to be a better investigator. Each case starts fresh — you carry your reputation and your case history, but not your contacts. You earn trust within a case. You don't carry it out.

### Progression Within a Case

You begin each investigation alone with basic observational tools. As you demonstrate correct investigative behavior — running the right commands, forming plausible hypotheses, narrowing the blast radius — contacts begin to appear. Each contact unlocks a qualitatively different type of help:

| Contact | Unlocks |
|---|---|
| **The Sysadmin** — nervous, owes you a favor | One direct answer about pod state or config |
| **The Network Engineer** — speaks in metaphors | Networking-layer insight, translated |
| **The Archivist** — dry, precise | Historical context: "this pattern appeared in case #7" |
| **The Senior Detective** — no patience for wrong theories | One blunt, honest hint — but judges your hypothesis first |

**Shipping note:** All four personas are on the **wire** in code: **Senior** (`hint`), **Sysadmin** (`hint sysadmin`), **Network Engineer** (`hint network`), **Archivist** (`hint archivist`). Unlock rules are behavior-based (see README / `internal/contacts/`). Each delivers **one** message per case when unlocked.

Contacts don't appear on a timer. They appear when you've earned them through investigative behavior. The game watches what you're doing.

### What Carries Between Cases

- Your **case file history** — a log of every investigation, outcomes, patterns seen
- Your **detective rank** — a narrative title that reflects cases solved, not a mechanical stat
- Your **known failure patterns** — the archivist can reference your history across cases

What does *not* carry: your contacts. Every case, you start alone.

### Tone

Dry. Atmospheric. The game does not celebrate you loudly. A solved case gets a quiet acknowledgment and a thorough debrief. The noir world does not do confetti. Failure is treated with the same seriousness as success — what went wrong, why, what to learn.

The hint system has personality. The Senior Detective is not kind. The Network Engineer will make you work for the translation. The Archivist speaks in case numbers.

---

## 4. The REPL

### Philosophy

The REPL vocabulary is investigative, not administrative. You are not running kubectl commands. You are conducting an investigation. The verbs should feel like actions a detective takes, not aliases for API calls.

### Core Verbs

```
observe                          # system-wide — what's visibly wrong right now
examine <resource> <name>        # deep dive a specific resource
trace <name>                     # follow the ownership chain upward
compare <resource> <name>        # spec vs actual reality
check logs <name>                # logs surfaced as evidence, not raw output
network <name>                   # connectivity, endpoints, service selectors
accuse <hypothesis>              # commit to your theory — required before solve
hint                             # request contact assistance (if earned)
status                           # your current case file — clues gathered so far
solve                            # opens fix sub-mode — only after accuse
debrief                          # full post-case explanation (after resolution)
```

### The Case File

`status` surfaces a running case file the game builds automatically as you investigate. Every significant observation is logged. You don't have to remember what you found — the file remembers for you.

This mirrors real incident response practice: good SREs write things down. The case file makes that behavior native to the tool.

### The Hypothesis Step

`accuse` is the pivot point of every investigation. Nothing in the `solve` sub-mode is available until you commit to a written theory. The LLM evaluates your hypothesis and responds in character — the incident commander on the other end of a fake Slack thread.

Responses are graduated:
- **Stone cold** — you're not even close, here's a nudge back toward evidence
- **Cold** — plausible but missing something
- **Warm** — right layer, wrong specific cause
- **Hot** — correct, proceed to solve

---

## 5. Technical Architecture

### The Cluster Contract

POD noir owns a single dedicated namespace: `pod-noir`. It applies manifests, watches state, and tears down cleanly when a session ends. It touches nothing outside its namespace.

Teardown is guaranteed — even on crash, on startup the game checks for and cleans up any orphaned resources from previous sessions.

### Failure Taxonomy

Failures are organized by layer. Each **scenario** ships fixed manifests that encode a failure mode into the controlled namespace (not generated at runtime).

**Scheduling**
- Resource requests exceed available node capacity
- Taint/toleration mismatch
- Node affinity rules unsatisfiable
- Node pressure (memory, disk)

**Runtime**
- Bad or nonexistent image
- CrashLoopBackOff (bad entrypoint, missing dependency)
- OOMKill
- Liveness/readiness probe misconfiguration

**Networking**
- Service selector mismatch (service routes to wrong pods)
- Wrong port mapping
- NetworkPolicy blocking expected traffic
- DNS resolution failure

**Configuration**
- Missing or wrong environment variables
- Referenced secret doesn't exist
- ConfigMap key mismatch
- Wrong volume mount path

**RBAC**
- Service account lacks required permissions

**Storage**
- PVC stuck in Pending (wrong storage class, no available PV)
- Volume mounted read-only when write required

### Difficulty and Composition

- **Level 1–2**: Single failure, single layer, obvious symptoms
- **Level 3–4**: Single failure, symptoms point at wrong layer
- **Level 5–6**: Two composed failures, one masks the other
- **Level 7+**: Cascading failures — fixing the obvious symptom surfaces the real problem

### LLM Integration Points

| Point | Role |
|---|---|
| Session start | Optional framing / atmosphere (embedded scenarios supply the failure mode) |
| `observe` / `examine` | Surface observations contextually from kubectl output |
| `accuse` | Evaluate hypothesis, respond in character |
| Contact interactions | NPC dialogue, qualitatively different per contact |
| `debrief` | Full root cause explanation, what was missed, what to study |

LLM provider is configuration-driven. Anthropic, Ollama, and OpenAI are first-class targets. The core has no opinion about which one you use.

### Event System

The core emits events at session boundaries and key moments. What handles those events is pluggable:

```go
type EventEmitter interface {
    Emit(event Event) error
}
```

**Built-in adapters:**
- `stdout` — prints events as JSON (default, works standalone)
- `nats` — publishes to a NATS broker (bot army integration)
- `webhook` — HTTP POST to a configured endpoint

**Events emitted:**
```
session.started      { scenario_id, difficulty, failure_layer, detective_name }
hypothesis.made      { text, judgment }
contact.unlocked     { contact_name, trigger }
session.solved       { duration, hints_used, hypotheses_made }
session.abandoned    { duration, last_command }
session.failed       { duration, reason }
```

If no emitter is configured, events are silently dropped. The game never fails because the broker is down.

---

## 6. Open Source Posture

### What's Public

- Full game engine
- Failure library (scenario templates)
- REPL and all core verbs
- All LLM integration points
- NATS, webhook, and stdout adapters
- Documentation

### What Stays Private (Your Fork)

- Bot army NATS topic schema
- Any scenario content tuned to your specific learning gaps
- Personal detective profile and case history

### Configuration Surface

```yaml
cluster:
  context: rancher-desktop        # kubectl context to use
  namespace: pod-noir             # owned namespace

llm:
  provider: anthropic             # anthropic | ollama | openai
  model: claude-sonnet-4-...
  api_key: $ANTHROPIC_API_KEY

events:
  adapter: nats                   # stdout | nats | webhook | none
  nats:
    url: nats://localhost:4222
    prefix: pod-noir

detective:
  name: "Sam Reyes"              # your character name
```

---

## 7. Roadmap

Roadmap is **intent**, not a promise date. For **what the repo actually does today**, see README, AGENTS, and code. **Embedded scenarios** (fixed YAML) are the product; procedural generation is out of scope unless that changes in **[architecture-decisions.md](architecture-decisions.md)**.

### Current — shipping (snapshot)

These are **in the codebase** as of the last northstar refresh:

- **Cluster contract:** dedicated namespace (`pod-noir` by default); apply/teardown; optional `-skip-cleanup`.
- **Scenarios:** **10** embedded cases covering rollout/config drift, missing Secret/`secretKeyRef`, bad image tag/pull, probe/port mismatch, memory/tmpfs OOM, Service selector vs Pod labels (empty endpoints), failing initContainer, ResourceQuota vs requests, PVC/StorageClass Pending, NetworkPolicy egress/DNS.
- **Session loop:** REPL investigative verbs, **accuse** (mock rules or HTTP LLM: cold / warm / hot), **solve** sub-mode with **precinct** kubectl policy, **debrief** (per-scenario mock text; generative debrief when LLM configured).
- **Contacts:** **four** wire-room NPCs — bare **`hint`** prints a **roster** (locked / open / done); **`hint senior`** (unlock: logs+trace or non-HOT **accuse**), **`hint sysadmin`** (unlock: **examine pod**), **`hint network`** (unlock: **trace**), **`hint archivist`** (unlock: **dossier** this session). **Per-scenario** authored wire copy for each NPC; one delivered message per contact per case.
- **Persistence:** SQLite **dossier** / history under `~/.pod-noir` (and related store).
- **Events:** **stdout** default; **NATS** (`POD_NOIR_EVENTS_ADAPTER`, optional bridge envelope to `events.pod_noir.*`).
- **LLM:** **mock** (no network); **Anthropic**, **OpenAI**, **Ollama** via configuration. Optional **LLM-generated wire-room `hint` messages** (with static copy as anchor and fallback); env **`POD_NOIR_LLM_CONTACT_WIRE`**.
- **Ops / quality:** CI **lint → Docker test → kind integration** including `playtest-smoke`; optional **git hooks** (`core.hooksPath=githooks`) for local smoke; **`podnoir doctor`** for reachability.

### Near term — next improvements

- **Docs:** contributor guide; **scenario authoring** note (how to add embedded manifests + `scenario.Definition` + briefing copy).
- **Scenarios:** more cases, **composed** or **multi-signal** failures (hard mode), rotating the matrix in **[docs/playtest-checklist.md](playtest-checklist.md)**.
- **Contacts:** richer per-scenario wire copy for **Sysadmin / Network / Archivist** (beyond current defaults + key cases), optional fifth persona, or LLM-generated contact lines.
- **Events:** **HTTP webhook** emitter (listed in §5 architecture; not yet first-class in the same way as stdout/NATS).
- **Polish:** difficulty labels or tags per scenario (optional; learning value over grind).

### Horizon — larger bets

- **LLM at session start:** optional atmospheric framing *on top of* fixed manifests (manifests still define the break).
- **Narrative progression:** detective rank or cross-case arc — **lightweight**, story-forward, not a stat treadmill.
- **Taxonomy depth:** “level 6+” composed/cascading failures as optional paths for players who cleared the core set.

### Explicit non-goals (for now)

- **Runtime-procedural** scenario generation replacing embedded YAML.
- **Multiplayer** or shared cluster competition.

---

## 8. Sample Investigations

> These samples serve as the canonical reference for scenario design. Each one defines the full player experience end to end — briefing, clue trail, NPC dialogue, hypothesis evaluation, and debrief. Use these as templates when authoring new scenarios.

---

### Case 001 — "The Overnight Shift"
**Layer:** Runtime | **Difficulty:** 2 | **Failure:** CrashLoopBackOff (bad entrypoint)

---

#### Incident Briefing

```
┌─────────────────────────────────────────────────────────────┐
│  THE CLUSTER AGENCY                                         │
│  Case File #001 — Incoming                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Client: Neon Ledger Financial                             │
│  Reported: 03:47 AM                                        │
│  Contact: "payments-worker has been down since the         │
│  deploy. Nothing is processing. Every transaction          │
│  since midnight is sitting in the queue."                  │
│                                                             │
│  Your supervisor left a note on your desk:                 │
│  "Fresh eyes. Don't touch anything until you know          │
│  what you're looking at. — D."                             │
│                                                             │
│  Namespace: pod-noir                                       │
│  Assigned: You                                             │
│                                                             │
│  [press enter to begin]                                    │
└─────────────────────────────────────────────────────────────┘
```

---

#### What's Actually Wrong (Hidden from Player)

The `payments-worker` deployment was updated to a new image. The new image's entrypoint script references a config file at `/app/config/settings.json` that doesn't exist in the image. The container starts, immediately errors, and Kubernetes restarts it in a loop. The image itself is valid and pullable — the problem is inside it.

**Injected manifest fragment:**
```yaml
containers:
  - name: payments-worker
    image: pod-noir/payments-worker:2.1.0   # image exists, entrypoint fails
    command: ["/app/start.sh"]              # start.sh exits 1 immediately
```

---

#### Clue Trail

**`> observe`**
```
CASE FILE — Initial Observations
─────────────────────────────────────────────────────────────
  payments-worker-7d9f8b-xkq2p    CrashLoopBackOff   2/2    14m
  payments-worker-7d9f8b-xkq2p    CrashLoopBackOff   2/2    14m

  Recent Events:
  → payments-worker-7d9f8b-xkq2p   Back-off restarting failed container
  → payments-worker-7d9f8b-xkq2p   Started container payments-worker
  → payments-worker-7d9f8b-xkq2p   Back-off restarting failed container

  Observation logged: payments-worker in CrashLoopBackOff, multiple restarts.
  The pod starts. Something kills it almost immediately.
─────────────────────────────────────────────────────────────
```

**`> examine pod payments-worker-7d9f8b-xkq2p`**
```
EVIDENCE — payments-worker-7d9f8b-xkq2p
─────────────────────────────────────────────────────────────
  Image:          pod-noir/payments-worker:2.1.0
  State:          Waiting (CrashLoopBackOff)
  Last State:     Terminated — Exit Code 1 — 23 seconds ago
  Restart Count:  18

  Liveness Probe:   HTTP GET /health — not the issue (container exits before probe fires)
  Readiness Probe:  HTTP GET /ready — not the issue

  Observation logged: Container exits with code 1. Not a probe failure —
  the container doesn't survive long enough for probes to matter.
─────────────────────────────────────────────────────────────
```

**`> check logs payments-worker-7d9f8b-xkq2p`**
```
EVIDENCE — Last Container Logs
─────────────────────────────────────────────────────────────
  /app/start.sh: line 4: /app/config/settings.json: No such file or directory
  /app/start.sh: Configuration file not found. Exiting.

  Observation logged: Entrypoint script is looking for /app/config/settings.json.
  File does not exist. Container dies on startup, every time.
─────────────────────────────────────────────────────────────
```

**`> trace payments-worker`**
```
EVIDENCE — Ownership Chain
─────────────────────────────────────────────────────────────
  payments-worker-7d9f8b-xkq2p
    ↑ owned by ReplicaSet: payments-worker-7d9f8b
      ↑ owned by Deployment: payments-worker
        Image: pod-noir/payments-worker:2.1.0
        Previous Image: pod-noir/payments-worker:2.0.3
        Last Rollout: 14 minutes ago

  Observation logged: Image was updated 14 minutes ago — matches reported
  outage window. Version 2.0.3 was running before. 2.1.0 is the change.
─────────────────────────────────────────────────────────────
```

---

#### Contact Unlock

> After `trace` reveals the version change, **The Archivist** appears.

```
─────────────────────────────────────────────────────────────
  [INCOMING MESSAGE — The Archivist]

  "Case #001 is your first, so I'll say this once.
  A container that exits immediately with code 1 and
  a missing file is not a Kubernetes problem.
  Kubernetes is doing exactly what it should.

  The question is what changed at 2.1.0.
  I'd check the image. If you can't pull it apart,
  check what the entrypoint expects to find."

  [The Archivist has shared one observation. File logged.]
─────────────────────────────────────────────────────────────
```

---

#### Hypothesis Evaluation

**`> accuse The new image version 2.1.0 introduced a broken entrypoint that expects a config file that doesn't exist in the image`**

```
─────────────────────────────────────────────────────────────
  [HYPOTHESIS RECEIVED — Evaluating]

  Judgment: HOT

  "That's the case. The entrypoint script assumes
  /app/config/settings.json exists. It doesn't.
  Version 2.0.3 had it baked in. Someone pulled it
  out of 2.1.0 and forgot the script still needed it.

  You can solve this. Go."
─────────────────────────────────────────────────────────────
```

---

#### Solve Sub-Mode

```
[SOLVE MODE — payments-worker]
─────────────────────────────────────────────────────────────
  What's your move?

  > rollback deployment payments-worker

  Rolling back payments-worker to pod-noir/payments-worker:2.0.3...
  Rollback complete.

  Watching pod state...
  payments-worker-6c4d2a-p9m1r    Running    1/1    12s

  payments-worker is running. Queue is processing.
  Neon Ledger Financial has been notified.
─────────────────────────────────────────────────────────────
```

---

#### Debrief

```
┌─────────────────────────────────────────────────────────────┐
│  CASE #001 — CLOSED                                         │
│  "The Overnight Shift"                                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ROOT CAUSE                                                 │
│  Image version 2.1.0 shipped with an entrypoint script      │
│  that referenced /app/config/settings.json. The file was   │
│  removed from the image in the same release. The script     │
│  exited immediately with code 1. Kubernetes restarted the  │
│  container on schedule. Every restart failed the same way. │
│                                                             │
│  WHAT YOU DID WELL                                         │
│  → Read the logs before touching anything                  │
│  → Followed the ownership chain to find the version change │
│  → Formed a precise hypothesis before attempting a fix     │
│                                                             │
│  WHAT TO STUDY                                             │
│  → Container exit codes and what they mean                 │
│  → How CrashLoopBackOff backoff timing works               │
│  → Image layer inspection (docker inspect, dive)           │
│  → Why rollback is sometimes the right first move          │
│                                                             │
│  TIME: 11m 34s  |  HINTS USED: 1  |  HYPOTHESES: 1        │
│  CONTACTS MET: The Archivist                               │
│                                                             │
│  "Not bad for a first case. — D."                          │
└─────────────────────────────────────────────────────────────┘
```

---

### Case 002 — "The Silent Route"
**Layer:** Networking | **Difficulty:** 4 | **Failure:** Service selector mismatch

---

#### Incident Briefing

```
┌─────────────────────────────────────────────────────────────┐
│  THE CLUSTER AGENCY                                         │
│  Case File #002 — Incoming                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Client: Meridian Shipping Co.                             │
│  Reported: 11:22 AM                                        │
│  Contact: "The tracking API is returning 503 on every      │
│  request. Pods look fine in the dashboard. We deployed      │
│  a label refactor this morning — shouldn't have affected   │
│  anything."                                                 │
│                                                             │
│  Note: Client is insistent the pods are healthy.           │
│  Client is not wrong. That's what makes this interesting.  │
│                                                             │
│  Namespace: pod-noir                                       │
│  Assigned: You                                             │
│                                                             │
│  [press enter to begin]                                    │
└─────────────────────────────────────────────────────────────┘
```

---

#### What's Actually Wrong (Hidden from Player)

During a label refactor, the `tracking-api` deployment pods were updated with a new label: `app: tracking-api-v2`. The service selector still points to `app: tracking-api`. The service has no matching endpoints. Every request hits the service and goes nowhere. The pods are perfectly healthy — they're just invisible to the service.

**Injected manifest fragments:**
```yaml
# deployment (updated)
labels:
  app: tracking-api-v2     # changed during refactor

# service (not updated)
selector:
  app: tracking-api        # still pointing at old label
```

---

#### Clue Trail

**`> observe`**
```
CASE FILE — Initial Observations
─────────────────────────────────────────────────────────────
  tracking-api-5f7c9d-r3kp1    Running    1/1    2h14m
  tracking-api-5f7c9d-r3kp1    Running    1/1    2h14m
  tracking-api-5f7c9d-r3kp1    Running    1/1    2h14m

  Services:
  → tracking-api-svc    ClusterIP    10.96.44.12    port 80→8080

  Observation logged: All pods Running. Service exists.
  Client was right — nothing looks wrong on the surface.
─────────────────────────────────────────────────────────────
```

**`> examine pod tracking-api-5f7c9d-r3kp1`**
```
EVIDENCE — tracking-api-5f7c9d-r3kp1
─────────────────────────────────────────────────────────────
  Image:          pod-noir/tracking-api:3.4.1
  State:          Running
  Restart Count:  0
  Labels:         app=tracking-api-v2, version=3.4.1

  Liveness:   HTTP GET /health → 200 OK (last checked 8s ago)
  Readiness:  HTTP GET /ready  → 200 OK (last checked 8s ago)

  Observation logged: Pod is genuinely healthy. Probes passing.
  Label noted: app=tracking-api-v2
─────────────────────────────────────────────────────────────
```

**`> network tracking-api`**
```
EVIDENCE — Network Investigation
─────────────────────────────────────────────────────────────
  Service: tracking-api-svc
    Selector:   app=tracking-api
    ClusterIP:  10.96.44.12
    Port:       80 → 8080

  Endpoints: <none>

  ⚠ No endpoints matched. Service selector finds zero pods.

  Observation logged: Service selector is app=tracking-api.
  Pods are labeled app=tracking-api-v2.
  The service is pointing at nothing.
─────────────────────────────────────────────────────────────
```

**`> compare service tracking-api-svc`**
```
EVIDENCE — Spec vs Reality
─────────────────────────────────────────────────────────────
  SERVICE SELECTOR       POD LABELS (all replicas)
  ─────────────────────  ──────────────────────────
  app: tracking-api      app: tracking-api-v2
                         version: 3.4.1

  Drift detected: selector and pod labels do not match.
  This service currently routes to 0 pods.
  503 is the expected result of every inbound request.

  Observation logged: Mismatch is precise and complete.
  Label refactor updated pods. Service was not updated to match.
─────────────────────────────────────────────────────────────
```

---

#### Contact Unlock

> After `network` reveals zero endpoints, **The Network Engineer** appears.

```
─────────────────────────────────────────────────────────────
  [INCOMING MESSAGE — The Network Engineer]

  "Think of the service as a bouncer with a list.
  The list says 'app: tracking-api'.
  Your pods changed their names this morning.
  The bouncer hasn't seen the new list yet.

  Nobody gets in. Nobody gets out.
  The door works fine. The list is wrong."

  [The Network Engineer has shared one observation. File logged.]
─────────────────────────────────────────────────────────────
```

---

#### Hypothesis Evaluation

**`> accuse The label refactor updated pod labels to tracking-api-v2 but the service selector still points to tracking-api — service has no endpoints and cannot route traffic`**

```
─────────────────────────────────────────────────────────────
  [HYPOTHESIS RECEIVED — Evaluating]

  Judgment: HOT

  "Clean. You found it without touching a thing.
  The service is a dead end. Update the selector
  or roll the labels back. Either works.
  Pick one and explain why."
─────────────────────────────────────────────────────────────
```

---

#### Solve Sub-Mode

```
[SOLVE MODE — tracking-api-svc]
─────────────────────────────────────────────────────────────
  What's your move?

  > patch service tracking-api-svc selector app=tracking-api-v2

  Patching tracking-api-svc selector...
  selector updated: app → tracking-api-v2

  Watching endpoints...
  tracking-api-svc endpoints: 10.244.0.14:8080, 10.244.0.15:8080, 10.244.0.16:8080

  Service is routing. API responding 200.
  Meridian Shipping Co. notified.
─────────────────────────────────────────────────────────────
```

---

#### Debrief

```
┌─────────────────────────────────────────────────────────────┐
│  CASE #002 — CLOSED                                         │
│  "The Silent Route"                                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ROOT CAUSE                                                 │
│  A label refactor updated pod labels from tracking-api to  │
│  tracking-api-v2. The service selector was not updated.    │
│  Kubernetes services route by label match — no match means │
│  no endpoints, no endpoints means every request gets 503.  │
│  The pods were healthy the entire time.                    │
│                                                             │
│  WHAT YOU DID WELL                                         │
│  → Didn't assume the pods were the problem                 │
│  → Used network investigation before touching config       │
│  → Caught the label mismatch with compare before guessing  │
│                                                             │
│  WHAT TO STUDY                                             │
│  → How Kubernetes service endpoint resolution works        │
│  → Label selectors and why exact matching matters          │
│  → Operational discipline around label refactors           │
│  → kubectl get endpoints as a first-line diagnostic tool   │
│                                                             │
│  TIME: 9m 02s  |  HINTS USED: 1  |  HYPOTHESES: 1         │
│  CONTACTS MET: The Network Engineer                        │
│                                                             │
│  "503 with healthy pods. Classic. You'll see this again."  │
└─────────────────────────────────────────────────────────────┘
```

---

### Case 003 — "The Missing Witness"
**Layer:** Configuration | **Difficulty:** 4 | **Failure:** Referenced secret does not exist

---

#### Incident Briefing

```
┌─────────────────────────────────────────────────────────────┐
│  THE CLUSTER AGENCY                                         │
│  Case File #003 — Incoming                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Client: Volta Analytics                                   │
│  Reported: 2:15 PM                                        │
│  Contact: "New environment deployment. The reporting        │
│  service never came up. It's not even trying to start —   │
│  just sits there. We've never seen this before."           │
│                                                             │
│  "Sits there" is not a Kubernetes status.                  │
│  Find out what it actually means.                          │
│                                                             │
│  Namespace: pod-noir                                       │
│  Assigned: You                                             │
│                                                             │
│  [press enter to begin]                                    │
└─────────────────────────────────────────────────────────────┘
```

---

#### What's Actually Wrong (Hidden from Player)

The `reporting-svc` deployment references a secret named `volta-db-credentials` for database connection info. The secret was never created in this namespace — it exists in the production namespace but wasn't carried over to this new environment. The pod is stuck in `Pending` → actually it never gets scheduled because the secret mount fails at admission — pod stays in `ContainerCreating` indefinitely.

**Injected manifest fragment:**
```yaml
volumes:
  - name: db-creds
    secret:
      secretName: volta-db-credentials    # secret does not exist
```

---

#### Clue Trail

**`> observe`**
```
CASE FILE — Initial Observations
─────────────────────────────────────────────────────────────
  reporting-svc-8b3f1c-m7nq4    ContainerCreating    0/1    23m

  Recent Events:
  → reporting-svc-8b3f1c-m7nq4   MountVolume.SetUp failed:
    secret "volta-db-credentials" not found

  Observation logged: Pod has been in ContainerCreating for 23 minutes.
  Volume mount is failing. Secret reference is the cause.
─────────────────────────────────────────────────────────────
```

**`> examine pod reporting-svc-8b3f1c-m7nq4`**
```
EVIDENCE — reporting-svc-8b3f1c-m7nq4
─────────────────────────────────────────────────────────────
  Image:          pod-noir/reporting-svc:1.0.0
  State:          Waiting (ContainerCreating)
  Restart Count:  0

  Volumes:
  → db-creds (secret: volta-db-credentials)   ← MOUNT FAILING
  → reporting-config (configmap: volta-config) ← OK

  Observation logged: Two volumes. ConfigMap mounts fine.
  Secret mount is the blocker. Pod cannot start until
  all volumes are ready.
─────────────────────────────────────────────────────────────
```

**`> compare secret volta-db-credentials`**
```
EVIDENCE — Secret Investigation
─────────────────────────────────────────────────────────────
  Searching for secret: volta-db-credentials
  Namespace: pod-noir

  Result: NOT FOUND

  Searching adjacent namespaces...
  → production: volta-db-credentials    EXISTS (created 4 months ago)
  → staging:    volta-db-credentials    EXISTS (created 2 months ago)
  → pod-noir:   volta-db-credentials    MISSING

  Observation logged: Secret exists in other namespaces.
  Was not created here. New environment deployment — secret
  provisioning was likely missed in the runbook.
─────────────────────────────────────────────────────────────
```

---

#### Contact Unlock

> After `compare` surfaces the missing secret, **The Sysadmin** appears — nervous, speaks quickly.

```
─────────────────────────────────────────────────────────────
  [INCOMING MESSAGE — The Sysadmin]

  "Oh god, yeah — I think I know what happened.
  We have a runbook for new environment setup and
  the secrets step is manual. Someone always forgets it.
  The secret's in 1Password under 'volta-db-credentials-prod'.

  I can't create it for you — different access.
  But that's definitely what's missing."

  [The Sysadmin has shared one observation. File logged.]
─────────────────────────────────────────────────────────────
```

---

#### Hypothesis Evaluation

**`> accuse The secret volta-db-credentials was not created in this namespace during the new environment setup — pod cannot mount the volume and is stuck in ContainerCreating`**

```
─────────────────────────────────────────────────────────────
  [HYPOTHESIS RECEIVED — Evaluating]

  Judgment: HOT

  "Correct. The pod isn't broken. The environment is
  incomplete. Fix the environment.

  And after you're done — note that the real fix
  isn't just creating the secret. It's fixing the
  runbook so the next environment doesn't have this
  conversation at 2pm."
─────────────────────────────────────────────────────────────
```

---

#### Debrief

```
┌─────────────────────────────────────────────────────────────┐
│  CASE #003 — CLOSED                                         │
│  "The Missing Witness"                                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ROOT CAUSE                                                 │
│  Secret volta-db-credentials was never provisioned in the  │
│  pod-noir namespace during new environment setup. The pod  │
│  specification referenced the secret as a volume mount.    │
│  Kubernetes held the pod in ContainerCreating indefinitely │
│  — it will never start until the secret exists.            │
│                                                             │
│  WHAT YOU DID WELL                                         │
│  → Read the event stream immediately — it told you exactly │
│    what was wrong within the first observe                 │
│  → Checked adjacent namespaces to confirm scope            │
│  → Recognized this as an environment problem, not a        │
│    configuration bug                                       │
│                                                             │
│  WHAT TO STUDY                                             │
│  → How Kubernetes handles missing secret volume mounts     │
│  → The difference between Pending and ContainerCreating    │
│  → Secret management patterns across environments          │
│  → Why manual runbook steps fail and how to automate them  │
│                                                             │
│  TIME: 7m 18s  |  HINTS USED: 1  |  HYPOTHESES: 1         │
│  CONTACTS MET: The Sysadmin                                │
│                                                             │
│  "Environment incomplete, not app broken. Good             │
│  distinction. Most people create the secret and            │
│  forget to fix the runbook. Don't be most people."         │
└─────────────────────────────────────────────────────────────┘
```

---

### Case 004 — "Double Exposure"
**Layer:** Composed | **Difficulty:** 6 | **Failure:** OOMKill masking a readiness probe misconfiguration

---

#### Incident Briefing

```
┌─────────────────────────────────────────────────────────────┐
│  THE CLUSTER AGENCY                                         │
│  Case File #004 — Incoming                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Client: Axiom Data Platform                               │
│  Reported: 9:04 AM                                        │
│  Contact: "The ingest pipeline is flapping. It comes up   │
│  for a few minutes then dies. We bumped memory limits      │
│  yesterday — thought that fixed it. Still happening."      │
│                                                             │
│  They bumped memory and it's still happening.              │
│  Whatever they think the problem is, they're half right.  │
│                                                             │
│  Namespace: pod-noir                                       │
│  Assigned: You                                             │
│                                                             │
│  [press enter to begin]                                    │
└─────────────────────────────────────────────────────────────┘
```

---

#### What's Actually Wrong (Hidden from Player)

There are two failures layered together:

**Failure A (visible):** The `ingest-pipeline` pod has a memory limit that's still too low. It gets OOMKilled roughly every 8 minutes. This is what the client noticed and partially addressed — but didn't fix completely.

**Failure B (hidden, deeper):** The readiness probe is checking `/ready` on port 8080. After the memory bump, the application was reconfigured to serve health checks on port 9090. The probe still hits 8080. The pod oscillates: starts up, gets marked not-ready immediately (probe fails), never receives traffic, then gets OOMKilled when background processing builds up — which obscures the probe issue entirely. Fixing only the OOM will reveal the probe failure. This is the real lesson.

```yaml
# memory limit still low (was bumped but not enough)
resources:
  limits:
    memory: "256Mi"     # app needs ~400Mi under load

# readiness probe on wrong port
readinessProbe:
  httpGet:
    path: /ready
    port: 8080          # app moved health checks to 9090
  initialDelaySeconds: 5
  periodSeconds: 10
```

---

#### Clue Trail

**`> observe`**
```
CASE FILE — Initial Observations
─────────────────────────────────────────────────────────────
  ingest-pipeline-9c2e7f-v4px8    Running    0/1    4m12s
  (1 restart in last 10 minutes)

  Recent Events:
  → ingest-pipeline-9c2e7f-v4px8   OOMKilled — 7 minutes ago
  → ingest-pipeline-9c2e7f-v4px8   Started
  → ingest-pipeline-9c2e7f-v4px8   OOMKilled — 19 minutes ago

  Observation logged: Pod is Running but not Ready (0/1).
  OOMKills visible in recent history.
  Currently alive — but something is still wrong.
─────────────────────────────────────────────────────────────
```

**`> examine pod ingest-pipeline-9c2e7f-v4px8`**
```
EVIDENCE — ingest-pipeline-9c2e7f-v4px8
─────────────────────────────────────────────────────────────
  Image:          pod-noir/ingest-pipeline:4.2.0
  State:          Running (but not Ready)
  Restart Count:  3

  Memory Limit:   256Mi
  Memory Usage:   201Mi (current) — rising

  Readiness Probe: HTTP GET /ready:8080
    Last result:   FAIL (connection refused)
    Consecutive failures: 12

  Observation logged: Two signals here.
  Memory is rising toward the limit — OOMKill is coming again.
  Readiness probe is failing — port 8080 refusing connections.
  Pod is Running but will never serve traffic in this state.
─────────────────────────────────────────────────────────────
```

**`> network ingest-pipeline`**
```
EVIDENCE — Network Investigation
─────────────────────────────────────────────────────────────
  Service: ingest-pipeline-svc
    Selector: app=ingest-pipeline   ← matches
    Endpoints: <none>

  Reason: Pod is Running but not Ready.
  Kubernetes withholds not-ready pods from service endpoints.

  Observation logged: Service is correctly configured.
  Pod is excluded from routing because readiness probe is failing.
  Traffic cannot reach ingest-pipeline regardless of OOM status.
─────────────────────────────────────────────────────────────
```

**`> check logs ingest-pipeline-9c2e7f-v4px8`**
```
EVIDENCE — Current Container Logs
─────────────────────────────────────────────────────────────
  [INFO]  ingest-pipeline starting — version 4.2.0
  [INFO]  HTTP server listening on :9090
  [INFO]  Health endpoints: /ready, /health on :9090
  [WARN]  Memory pressure detected — background queue backlog: 14,203 items
  [WARN]  Memory pressure detected — background queue backlog: 31,447 items

  Observation logged: Application is serving on port 9090.
  Health endpoints are on 9090.
  Readiness probe is checking 8080.
  8080 is not open. Probe will fail every time.

  Memory pressure is also real and rising.
─────────────────────────────────────────────────────────────
```

---

#### Contact Unlocks

> After `examine` reveals both signals, **The Senior Detective** appears.

```
─────────────────────────────────────────────────────────────
  [INCOMING MESSAGE — The Senior Detective]

  "You've got two things wrong here. Don't fix one
  and call it done.

  The OOM is real. Fix it. But ask yourself why
  the pod is never ready. A pod that never serves
  traffic still runs background jobs. What happens
  to a background job queue when nothing is
  processing requests?"

  [The Senior Detective will not say more.]
─────────────────────────────────────────────────────────────
```

> After `check logs` confirms port 9090, **The Network Engineer** appears.

```
─────────────────────────────────────────────────────────────
  [INCOMING MESSAGE — The Network Engineer]

  "The probe is knocking on the wrong door.
  8080 is dark. Has been since 4.2.0.
  9090 is where the light is.

  Fix the probe. Then fix the memory.
  In that order — you'll understand why
  once you do."
─────────────────────────────────────────────────────────────
```

---

#### Hypothesis Evaluation

**`> accuse Two failures: readiness probe checking port 8080 but app moved health endpoints to 9090 in 4.2.0 — pod never becomes ready and never drains the queue, causing memory to spike and OOMKill. Fixing only the OOM would have revealed the probe failure and the cycle would continue.`**

```
─────────────────────────────────────────────────────────────
  [HYPOTHESIS RECEIVED — Evaluating]

  Judgment: HOT

  "That's the complete picture. Most investigators
  fix the OOM, close the ticket, and get paged again
  in eight minutes.

  You saw both. You understood the interaction.
  That's the difference between someone who fixes
  symptoms and someone who closes cases.

  Go."
─────────────────────────────────────────────────────────────
```

---

#### Solve Sub-Mode

```
[SOLVE MODE — ingest-pipeline]
─────────────────────────────────────────────────────────────
  Two fixes required. Apply both.

  > patch deployment ingest-pipeline readinessProbe port=9090

  Patching readiness probe port: 8080 → 9090
  Pod restarting to apply changes...

  Watching probe status...
  Readiness probe: PASS (200 OK on /ready:9090)
  ingest-pipeline-9c2e7f-n8wk2: Running    1/1    22s ✓

  > patch deployment ingest-pipeline memory-limit=512Mi

  Patching memory limit: 256Mi → 512Mi
  Pod restarting...
  ingest-pipeline-9c2e7f-t1qr6: Running    1/1    18s ✓

  Memory usage stabilizing at 310Mi. Queue draining.
  Service endpoints restored. Traffic flowing.
  Axiom Data Platform notified.
─────────────────────────────────────────────────────────────
```

---

#### Debrief

```
┌─────────────────────────────────────────────────────────────┐
│  CASE #004 — CLOSED                                         │
│  "Double Exposure"                                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ROOT CAUSE                                                 │
│  Two failures interacting:                                  │
│                                                             │
│  1. Readiness probe checking :8080 after application       │
│  moved health endpoints to :9090 in version 4.2.0.        │
│  Pod started but was never marked Ready. Never received    │
│  traffic. Background queue accumulated.                    │
│                                                             │
│  2. Memory limit of 256Mi insufficient for the queue       │
│  backlog that built up because the pod wasn't serving      │
│  traffic. OOMKill restarted the pod before the probe       │
│  misconfiguration could be identified.                     │
│                                                             │
│  The OOM masked the probe failure. The client bumped       │
│  memory, the OOM came back slower, the probe failure       │
│  was never identified. Both needed to be fixed.            │
│                                                             │
│  WHAT YOU DID WELL                                         │
│  → Didn't stop at the first explanation (OOM)             │
│  → Read logs to confirm actual serving port               │
│  → Understood the causal relationship between failures     │
│  → Fixed both in the right order                          │
│                                                             │
│  WHAT TO STUDY                                             │
│  → How readiness vs liveness probes differ in behavior     │
│  → Why not-ready pods accumulate background work           │
│  → Probe configuration — port, path, timing parameters    │
│  → Composed failure patterns and how to spot them          │
│                                                             │
│  TIME: 18m 44s  |  HINTS USED: 2  |  HYPOTHESES: 1        │
│  CONTACTS MET: The Senior Detective, The Network Engineer  │
│                                                             │
│  "This is the shape of real incidents. Something obvious   │
│  hiding something that matters. You found both.            │
│  That's the job." — D.                                     │
└─────────────────────────────────────────────────────────────┘
```

---

---

### Case 005 — "The Velvet Rope"
**Layer:** Scheduling | **Difficulty:** 5 | **Failure:** Node affinity + stale maintenance taint, pod stuck Pending forever

---

#### Incident Briefing

```
┌─────────────────────────────────────────────────────────────┐
│  THE CLUSTER AGENCY                                         │
│  Case File #005 — Incoming                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Client: Ironclad Platform Engineering                     │
│  Reported: 4:33 PM                                        │
│  Contact: "We scaled up the inference service after        │
│  maintenance this morning. New replicas never came up.     │
│  The original pods are fine. The new ones just sit         │
│  there. We have capacity — I can see the nodes."           │
│                                                             │
│  They can see the nodes. The nodes can see the pods.       │
│  Something between them disagrees.                         │
│                                                             │
│  Namespace: pod-noir                                       │
│  Assigned: You                                             │
│                                                             │
│  [press enter to begin]                                    │
└─────────────────────────────────────────────────────────────┘
```

---

#### What's Actually Wrong (Hidden from Player)

Two scheduling constraints interact to make placement impossible:

**Constraint A:** The `inference-svc` deployment has a node affinity rule requiring nodes labeled `zone=us-west-2a`. This is intentional — the inference workload needs to colocate with a GPU node pool that only exists in that zone.

**Constraint B:** During this morning's maintenance window, the ops team applied a taint `maintenance=true:NoSchedule` to all nodes in `us-west-2a`. Maintenance ended. The taint was never removed. The pods the deployment is trying to schedule have no toleration for `maintenance=true`.

The original pods are fine because they were scheduled *before* the taint was applied. New pods — from the scale-up — cannot land anywhere. Every node that satisfies the affinity rule is tainted. Every untainted node fails the affinity rule.

```yaml
# deployment affinity (correct, intentional)
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: zone
              operator: In
              values: ["us-west-2a"]

# no tolerations defined
# nodes in us-west-2a still carry:
#   maintenance=true:NoSchedule   ← never cleaned up
```

---

#### Clue Trail

**`> observe`**
```
CASE FILE — Initial Observations
─────────────────────────────────────────────────────────────
  inference-svc-6d4a9c-r8wm1    Running    1/1    6h44m
  inference-svc-6d4a9c-p2kx7    Running    1/1    6h44m
  inference-svc-6d4a9c-tn3q9    Pending    0/1    47m
  inference-svc-6d4a9c-w1vr4    Pending    0/1    47m
  inference-svc-6d4a9c-bz8s2    Pending    0/1    47m

  Recent Events:
  → inference-svc-6d4a9c-tn3q9   0/3 nodes are available:
    3 node(s) had untolerated taint {maintenance: true}
  → inference-svc-6d4a9c-w1vr4   0/3 nodes are available:
    3 node(s) had untolerated taint {maintenance: true}

  Observation logged: Original pods running. New pods Pending 47 minutes.
  Scheduler cannot place them. Taint mentioned in events.
  But the client said maintenance is over.
─────────────────────────────────────────────────────────────
```

**`> examine pod inference-svc-6d4a9c-tn3q9`**
```
EVIDENCE — inference-svc-6d4a9c-tn3q9
─────────────────────────────────────────────────────────────
  Image:          pod-noir/inference-svc:6.0.1
  State:          Pending
  Scheduled:      False — no node selected
  Restart Count:  0

  Node Affinity (required):
  → zone=us-west-2a

  Tolerations:
  → node.kubernetes.io/not-ready:NoExecute (default)
  → node.kubernetes.io/unreachable:NoExecute (default)
  → maintenance=true: NOT PRESENT

  Observation logged: Pod must land in zone=us-west-2a.
  Pod has no toleration for maintenance=true.
  If any node in us-west-2a carries that taint,
  this pod has nowhere to go.
─────────────────────────────────────────────────────────────
```

**`> trace inference-svc`**
```
EVIDENCE — Ownership Chain
─────────────────────────────────────────────────────────────
  inference-svc-6d4a9c-tn3q9
    ↑ owned by ReplicaSet: inference-svc-6d4a9c
      ↑ owned by Deployment: inference-svc
        Replicas desired: 5
        Replicas ready:   2
        Replicas pending: 3

  Scale event: 2 → 5 replicas at 3:46 PM (47 minutes ago)
  Triggered by: manual kubectl scale

  Observation logged: Scale-up happened after maintenance window.
  Original 2 pods predate the taint.
  New 3 pods were created into a tainted environment.
─────────────────────────────────────────────────────────────
```

**`> examine node worker-us-west-2a-01`**
```
EVIDENCE — worker-us-west-2a-01
─────────────────────────────────────────────────────────────
  Zone:           us-west-2a
  Status:         Ready
  CPU:            14% utilized
  Memory:         31% utilized

  Taints:
  → maintenance=true:NoSchedule   (applied: 9:12 AM today)

  Labels:
  → zone=us-west-2a
  → node-pool=gpu-inference

  Observation logged: Node is healthy and has capacity.
  Taint applied at 9:12 AM — start of maintenance window.
  Maintenance ended hours ago. Taint remains.
  Same taint present on all us-west-2a nodes (checked).
─────────────────────────────────────────────────────────────
```

---

#### Contact Unlocks

> After `examine pod` surfaces the missing toleration, **The Archivist** appears.

```
─────────────────────────────────────────────────────────────
  [INCOMING MESSAGE — The Archivist]

  "Case #002 was a selector that didn't match.
  This is the same shape — a pod and a node that
  cannot find each other — but the mechanism differs.

  Selectors say where a pod wants to go.
  Taints say who a node will refuse.
  Tolerations are the exception.

  Your pod wants the right place.
  The place won't have it.
  One of those facts needs to change."

  [The Archivist has shared one observation. File logged.]
─────────────────────────────────────────────────────────────
```

> After `examine node` confirms taint timestamp, **The Sysadmin** appears.

```
─────────────────────────────────────────────────────────────
  [INCOMING MESSAGE — The Sysadmin]

  "Oh no. Oh no no no.
  I applied those taints at 9 AM. I was supposed
  to remove them when we finished at noon.
  I got pulled into something and I completely —

  It's been seven hours.

  I'm sorry. That's my fault. I can remove
  the taints right now if that's the fix."

  [The Sysadmin has shared one observation. File logged.]
─────────────────────────────────────────────────────────────
```

---

#### Hypothesis Evaluation

**`> accuse Maintenance taint maintenance=true:NoSchedule was applied to all us-west-2a nodes this morning and never removed after maintenance ended. New pods require zone=us-west-2a via affinity but have no toleration for the maintenance taint — scheduler cannot place them anywhere. Original pods are unaffected because they were scheduled before the taint was applied.`**

```
─────────────────────────────────────────────────────────────
  [HYPOTHESIS RECEIVED — Evaluating]

  Judgment: HOT

  "Complete. You identified both constraints and
  understood why the original pods were immune.
  That last part is what separates a real diagnosis
  from a lucky guess.

  Two paths forward: remove the stale taint, or
  add a toleration to the deployment. Think about
  which one is correct before you act.

  One of them treats the cause. One treats
  the symptom."
─────────────────────────────────────────────────────────────
```

---

#### Solve Sub-Mode

```
[SOLVE MODE — worker-us-west-2a nodes]
─────────────────────────────────────────────────────────────
  Two options. You choose.

  > remove taint worker-us-west-2a-01 maintenance=true:NoSchedule
  > remove taint worker-us-west-2a-02 maintenance=true:NoSchedule
  > remove taint worker-us-west-2a-03 maintenance=true:NoSchedule

  Taints removed from all us-west-2a nodes.

  Watching pending pods...
  inference-svc-6d4a9c-tn3q9    Running    1/1    8s
  inference-svc-6d4a9c-w1vr4    Running    1/1    9s
  inference-svc-6d4a9c-bz8s2    Running    1/1    11s

  All 5 replicas running. Inference service at capacity.
  Ironclad Platform Engineering notified.
─────────────────────────────────────────────────────────────
```

---

#### Debrief

```
┌─────────────────────────────────────────────────────────────┐
│  CASE #005 — CLOSED                                         │
│  "The Velvet Rope"                                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ROOT CAUSE                                                 │
│  Maintenance taint maintenance=true:NoSchedule applied to  │
│  all nodes in us-west-2a was never removed after the       │
│  maintenance window closed. The inference-svc deployment   │
│  requires zone=us-west-2a via node affinity. New pods had  │
│  no toleration for the taint. Every eligible node refused  │
│  them. Scheduler had no valid placement — pods sat         │
│  Pending indefinitely.                                     │
│                                                             │
│  WHY ORIGINAL PODS WERE UNAFFECTED                        │
│  Taints with NoSchedule effect prevent new scheduling.     │
│  They do not evict pods already running on the node.       │
│  Pods scheduled before 9 AM were never touched.            │
│                                                             │
│  WHY REMOVING THE TAINT WAS CORRECT                       │
│  Adding a toleration to the deployment would have worked   │
│  mechanically — but would have left a permanent exception  │
│  for a condition that should not exist. The taint was      │
│  stale. The right fix was to remove the stale state,       │
│  not adapt the application to it.                          │
│                                                             │
│  WHAT YOU DID WELL                                         │
│  → Checked the nodes, not just the pods                   │
│  → Noticed the taint timestamp matched the maintenance     │
│    window — connected timing to cause                      │
│  → Understood why existing pods were immune                │
│  → Chose the correct fix: remove stale state               │
│                                                             │
│  WHAT TO STUDY                                             │
│  → Taint effects: NoSchedule vs NoExecute vs PreferNoSchedule │
│  → How tolerations work and when to use them               │
│  → Node affinity required vs preferred rules               │
│  → Operational discipline: taints as temporary state       │
│  → Why runbooks need cleanup steps, not just apply steps   │
│                                                             │
│  TIME: 16m 21s  |  HINTS USED: 2  |  HYPOTHESES: 1        │
│  CONTACTS MET: The Archivist, The Sysadmin                 │
│                                                             │
│  "The nodes had capacity the entire time. The client       │
│  was right about that. The scheduler just wouldn't         │
│  let the pods through the door. Now you know why."  — D.  │
└─────────────────────────────────────────────────────────────┘
```

---

---

## Document governance & maintenance

**This file** is the product constitution. Update it when **identity**, **learning model**, **world tone**, **REPL contract**, or **cluster/namespace behavior** change in ways that affect what players experience or what contributors assume. Follow **`.cursor/rules/northstar-sync.mdc`**.

**Sweeping** system or contributor-contract changes (solve policy, precinct rules, session lifecycle, CI game contract, hooks strategy, LLM wiring) also get a row in **[architecture-decisions.md](architecture-decisions.md)**. Routine scenario tweaks and small fixes use **progress** bullets only.

Roadmap sections below describe **direction**; **current** scenario count, verbs, and tooling are defined by the repo (**README**, **AGENTS**, code).

---

*Last updated: 2026-04-02*
*Status: Living — Identity, Learning Model, World, REPL, Architecture, OSS Posture, Roadmap; sample investigations illustrative; governance + AD log for accuracy*
