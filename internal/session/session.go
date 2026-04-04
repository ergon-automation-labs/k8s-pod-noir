package session

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"podnoir/internal/contacts"
	"podnoir/internal/events"
	"podnoir/internal/kubectl"
	"podnoir/internal/llm"
	"podnoir/internal/precinct"
	"podnoir/internal/scenario"
	"podnoir/internal/store"
)

type Session struct {
	Out io.Writer
	In  io.Reader

	Kube      *kubectl.Runner
	Def       *scenario.Definition
	Store     *store.Store
	SessID    int64
	NS        string
	Detective string

	Emitter events.Emitter
	LLM     llm.Provider

	accused       bool
	hotAccusation bool
	solveMode     bool

	inv contacts.InvestigationState

	shownObserveFieldNote bool
	shownExamineFieldNote bool

	// REPL shortcuts / history
	lastExpandedCmd  string
	lastSolveKubectl string
	lastLogsPod      string
	replHistory      []string

	ctx       context.Context
	cancel    context.CancelFunc
	cleanupFn func()
	closeOnce sync.Once
}

// New creates a session, applies scenario manifests, and registers persistence.
func New(ctx context.Context, kube *kubectl.Runner, st *store.Store, def *scenario.Definition, detective, ns string, out io.Writer, in io.Reader, cleanup func(), emitter events.Emitter, llmProvider llm.Provider) (*Session, error) {
	cctx, cancel := context.WithCancel(ctx)
	if emitter == nil {
		emitter = events.StdoutEmitter{W: os.Stdout}
	}
	if llmProvider == nil {
		llmProvider = llm.Mock{}
	}
	s := &Session{
		Out:       out,
		In:        in,
		Kube:      kube,
		Def:       def,
		Store:     st,
		NS:        ns,
		Detective: detective,
		Emitter:   emitter,
		LLM:       llmProvider,
		ctx:       cctx,
		cancel:    cancel,
		cleanupFn: cleanup,
	}
	id, err := st.StartSession(cctx, string(def.ID), detective)
	if err != nil {
		cancel()
		return nil, err
	}
	s.SessID = id

	for i, step := range def.ApplySteps {
		if len(step) == 0 {
			continue
		}
		if err := kube.ApplyYAML(cctx, step); err != nil {
			cancel()
			_ = st.EndSession(context.Background(), id, "apply_failed")
			return nil, fmt.Errorf("apply scenario step %d: %w", i, err)
		}
		if i == 0 && len(def.ApplySteps) > 1 && def.RolloutWaitAfterFirstStep != "" {
			if err := kube.RolloutStatus(cctx, ns, def.RolloutWaitAfterFirstStep, 2*time.Minute); err != nil {
				cancel()
				_ = st.EndSession(context.Background(), id, "rollout_wait_failed")
				return nil, fmt.Errorf("wait first rollout (stable revision): %w", err)
			}
		}
	}

	events.Emit(cctx, emitter, events.SessionStarted, map[string]any{
		"scenario_id": string(def.ID),
		"namespace":   ns,
		"detective":   detective,
	})
	if err := st.TouchCaseOpen(cctx, string(def.ID), detective); err != nil {
		cancel()
		_ = st.EndSession(context.Background(), id, "folder_touch_failed")
		return nil, fmt.Errorf("case folder history: %w", err)
	}
	return s, nil
}

func (s *Session) Close(outcome string) {
	s.closeOnce.Do(func() {
		if s.cleanupFn != nil {
			s.cleanupFn()
		}
		bg := context.Background()
		_ = s.Store.EndSession(bg, s.SessID, outcome)
		_ = s.Store.RecordCaseFolderOutcome(bg, string(s.Def.ID), outcome)
		s.cancel()
	})
}

func (s *Session) RunREPL() error {
	fmt.Fprintln(s.Out, s.Def.Briefing(s.Detective))
	fmt.Fprintf(s.Out, "%s\n\n", s.Def.CurtainLine())
	fmt.Fprintln(s.Out, "The wire room is yours. Type help. quit closes the file and tears down the namespace.")

	sc := bufio.NewScanner(s.In)
	for {
		if s.solveMode {
			fmt.Fprint(s.Out, "[solve] kubectl> ")
		} else {
			fmt.Fprint(s.Out, "> ")
		}
		if !sc.Scan() {
			break
		}
		orig := strings.TrimSpace(sc.Text())
		if orig == "" {
			continue
		}
		expanded, xerr := s.expandReplShortcuts(orig, s.solveMode)
		if xerr != nil {
			fmt.Fprintln(s.Out, "error:", xerr)
			continue
		}
		if err := s.handle(expanded); err != nil {
			if err == errQuit {
				return nil
			}
			fmt.Fprintln(s.Out, "error:", err)
			continue
		}
		s.recordReplSuccess(orig, expanded, s.solveMode)
	}
	return sc.Err()
}

var errQuit = fmt.Errorf("quit")

func (s *Session) handle(line string) error {
	low := strings.ToLower(line)

	if s.solveMode {
		switch low {
		case "exit", "done", "leave":
			s.solveMode = false
			fmt.Fprintln(s.Out, "Left solve mode.")
			return nil
		case "cabinet", "files":
			fmt.Fprintln(s.Out, precinct.CabinetPeek(s.Def, s.Store))
			return nil
		case "dossier":
			return s.dossier()
		case "hist", "history":
			s.showHistory()
			return nil
		case "quit":
			s.Close("abandoned")
			events.Emit(s.ctx, s.Emitter, events.SessionAbandoned, map[string]any{"last_command": line})
			return errQuit
		}
		if err := kubectl.EnsureSolvePolicy(line, s.NS); err != nil {
			return err
		}
		wd, wdErr := os.Getwd()
		if wdErr != nil {
			wd = "."
		}
		if err := kubectl.EnsureSolveApplyManifests(line, s.NS, wd); err != nil {
			return err
		}
		out, err := s.execKubectl(line)
		if err != nil {
			return err
		}
		s.rememberSolveLine(line)
		fmt.Fprintln(s.Out, string(out))
		return nil
	}

	switch {
	case low == "help" || low == "?":
		fmt.Fprintln(s.Out, strings.TrimSpace(`
Commands:
  observe                         pod list + recent events in the case namespace
  examine pod <name>              kubectl describe pod
  check logs <name>               kubectl logs (tail)
  trace <name>                    pod/deployment chain + rollout history
  accuse <hypothesis>             commit theory (mock LLM judgment)
  solve                           enter kubectl mode (requires HOT accusation)
  status                          case file notes
  cabinet, files                 glance back at the file drawer (other cases)
  dossier                       your local case-folder history (cleared counts)
  hist, history                 last dozen commands this session
  debrief                       close the case once the cluster looks healthy (mock)
  hint                            wire roster — who's unlocked vs locked
  hint senior|sysadmin|network|archivist
                                take that contact's call (one message per case each)
  quit                          exit; namespace is cleared (unless -skip-cleanup)

Shortcuts (normal mode):
  o        observe          t <name>   trace
  l <pod>  check logs       x <pod>    examine pod
  l        repeat last logs (after one check logs)
  r, again repeat last command

Solve mode (examples):
  case 001: kubectl rollout undo deployment/payments-worker -n pod-noir
  case 002: kubectl create secret generic ledger-signing-secret -n pod-noir ...
  case 003: kubectl set image deployment/shipping-notifier notifier=busybox:1.36.1 -n pod-noir
  case 004: patch/remove livenessProbe on deployment/bedside-console
  case 005: raise memory limits or fix start command on deployment/memory-witness
  case 006: kubectl patch service gateway-svc ... selector app=gateway-api
  case 007: fix or remove failing initContainer on deployment/witness-hold -n pod-noir
  case 008: raise ResourceQuota or lower deployment requests — ledger-queue -n pod-noir
  case 009: fix PVC storageClassName / binding — evidence-vol, evidence-worker -n pod-noir
  case 010: relax NetworkPolicy egress or delete lock-the-door — tape-deck -n pod-noir
  kubectl patch ... --type=json OR --type=strategic  (see debrief)

Solve mode: r / again repeats last kubectl. Precinct blocks -A, -k/kustomize, namespace delete, cluster admins, taint, etc.; apply -f needs -n and YAML is checked (no other namespace, no cluster-scoped kinds).`))
		return nil
	case low == "hist" || low == "history":
		s.showHistory()
		return nil
	case low == "quit" || low == "exit":
		s.Close("abandoned")
		events.Emit(s.ctx, s.Emitter, events.SessionAbandoned, map[string]any{"last_command": line})
		return errQuit
	case low == "observe":
		return s.observe()
	case strings.HasPrefix(low, "examine pod "):
		name := strings.TrimSpace(line[len("examine pod "):])
		return s.examinePod(name)
	case strings.HasPrefix(strings.ToLower(line), "check logs "):
		name := strings.TrimSpace(line[len("check logs "):])
		return s.checkLogs(name)
	case strings.HasPrefix(low, "trace "):
		name := strings.TrimSpace(line[len("trace "):])
		return s.trace(name)
	case strings.HasPrefix(low, "accuse "):
		h := strings.TrimSpace(line[len("accuse "):])
		return s.accuse(h)
	case low == "solve":
		return s.enterSolve()
	case low == "status":
		return s.status()
	case strings.HasPrefix(low, "hint"):
		fields := strings.Fields(line)
		if len(fields) == 1 {
			return s.hintWireRoster()
		}
		id, err := contacts.ParseHintTarget(fields[1])
		if err != nil {
			return err
		}
		return s.hintWithTarget(id)
	case low == "cabinet" || low == "files":
		fmt.Fprintln(s.Out, precinct.CabinetPeek(s.Def, s.Store))
		return nil
	case low == "dossier":
		return s.dossier()
	case low == "debrief":
		return s.debrief()
	default:
		fmt.Fprintln(s.Out, "Unknown command. Try help — or cabinet to hear the other folders breathing.")
		return nil
	}
}

func (s *Session) observe() error {
	pods, err := s.Kube.Run(s.ctx, "get", "pods", "-n", s.NS, "-o", "wide")
	if err != nil {
		return err
	}
	ev, err := s.Kube.Run(s.ctx, "get", "events", "-n", s.NS, "--sort-by=.lastTimestamp")
	if err != nil {
		return err
	}
	fmt.Fprintf(s.Out, "FIELD NOTES — %s (%s)\n", s.Def.Title, s.NS)
	fmt.Fprintln(s.Out, string(pods))
	fmt.Fprintln(s.Out, "Recent events:")
	fmt.Fprintln(s.Out, string(ev))

	if !s.shownObserveFieldNote && strings.TrimSpace(s.Def.FieldNoteAfterObserve) != "" {
		s.shownObserveFieldNote = true
		fmt.Fprintln(s.Out)
		fmt.Fprintln(s.Out, strings.TrimSpace(s.Def.FieldNoteAfterObserve))
	}

	note := "Observed pods and events in namespace " + s.NS
	return s.logNote("observe", note)
}

func (s *Session) examinePod(name string) error {
	if name == "" {
		return fmt.Errorf("pod name required")
	}
	out, err := s.Kube.Run(s.ctx, "describe", "pod", "-n", s.NS, name)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.Out, "EVIDENCE — pod/%s\n", name)
	fmt.Fprintln(s.Out, string(out))
	if !s.shownExamineFieldNote && strings.TrimSpace(s.Def.FieldNoteAfterExamine) != "" {
		s.shownExamineFieldNote = true
		fmt.Fprintln(s.Out)
		fmt.Fprintln(s.Out, strings.TrimSpace(s.Def.FieldNoteAfterExamine))
	}
	if err := s.logNote("examine", "Described pod "+name); err != nil {
		return err
	}
	if contacts.SeniorPath(s.Def) {
		s.tryUnlockSysadmin("examine_pod")
	}
	return nil
}

func (s *Session) checkLogs(name string) error {
	if name == "" {
		return fmt.Errorf("pod name required")
	}
	out, err := s.Kube.Run(s.ctx, "logs", "-n", s.NS, name, "--tail=80")
	if err != nil {
		return err
	}
	fmt.Fprintln(s.Out, "EVIDENCE — logs")
	fmt.Fprintln(s.Out, string(out))
	if err := s.logNote("logs", "Fetched logs for "+name); err != nil {
		return err
	}
	s.lastLogsPod = name
	if contacts.SeniorPath(s.Def) {
		s.inv.SeenLogs = true
		s.maybeUnlockFromEvidence()
	}
	return nil
}

func (s *Session) trace(name string) error {
	if name == "" {
		return fmt.Errorf("name required")
	}
	rsOut, errPod := s.Kube.Run(s.ctx, "get", "pod", name, "-n", s.NS, "-o", `jsonpath={.metadata.ownerReferences[?(@.kind=="ReplicaSet")].name}`)
	rs := strings.TrimSpace(string(rsOut))
	if errPod == nil && rs != "" {
		depOut, errDep := s.Kube.Run(s.ctx, "get", "rs", rs, "-n", s.NS, "-o", `jsonpath={.metadata.ownerReferences[?(@.kind=="Deployment")].name}`)
		dep := strings.TrimSpace(string(depOut))
		fmt.Fprintf(s.Out, "EVIDENCE — trace pod/%s\n", name)
		fmt.Fprintf(s.Out, "  ReplicaSet: %s\n", rs)
		if errDep == nil && dep != "" {
			fmt.Fprintf(s.Out, "  Deployment: %s\n", dep)
			s.printRolloutHistory(dep)
		}
		if err := s.logNote("trace", "Traced pod "+name); err != nil {
			return err
		}
		s.markTraceSeen()
		return nil
	}
	out2, err2 := s.Kube.Run(s.ctx, "get", "deploy", "-n", s.NS, name, "-o", `jsonpath={.metadata.name}{"  image: "}{.spec.template.spec.containers[0].image}{"\n"}`)
	if err2 != nil {
		if errPod != nil {
			return fmt.Errorf("trace: not a pod or deployment in %s: %w", s.NS, errPod)
		}
		return fmt.Errorf("trace: not a pod or deployment in %s: %w", s.NS, err2)
	}
	fmt.Fprintf(s.Out, "EVIDENCE — trace deployment/%s\n", name)
	fmt.Fprintln(s.Out, string(out2))
	s.printRolloutHistory(name)
	if err := s.logNote("trace", "Traced deploy "+name); err != nil {
		return err
	}
	s.markTraceSeen()
	return nil
}

func (s *Session) printRolloutHistory(dep string) {
	hist, err := s.Kube.Run(s.ctx, "rollout", "history", "deployment/"+dep, "-n", s.NS)
	if err != nil {
		fmt.Fprintf(s.Out, "(rollout history unavailable: %v)\n", err)
		return
	}
	fmt.Fprintln(s.Out, "Rollout history:")
	fmt.Fprintln(s.Out, strings.TrimSpace(string(hist)))
}

func (s *Session) accuse(h string) error {
	res, err := s.LLM.EvaluateAccusation(s.ctx, s.Def, h)
	if err != nil {
		return err
	}
	s.accused = true
	if res.Judgment == llm.Hot {
		s.hotAccusation = true
	}
	events.Emit(s.ctx, s.Emitter, events.HypothesisMade, map[string]any{
		"text":     h,
		"judgment": string(res.Judgment),
	})
	fmt.Fprintf(s.Out, "[HYPOTHESIS — %s]\n%s\n", res.Judgment, res.Reply)
	if err := s.logNote("accuse", h+" ["+string(res.Judgment)+"]"); err != nil {
		return err
	}
	if contacts.SeniorPath(s.Def) && contacts.ShouldUnlockSeniorFromAccusation(res.Judgment) {
		s.tryUnlockSeniorDetective("accusation_not_hot")
	}
	return nil
}

func (s *Session) enterSolve() error {
	if !s.hotAccusation {
		return fmt.Errorf("solve locked until a HOT accusation — keep investigating")
	}
	fmt.Fprintln(s.Out, "Solve mode: raw kubectl (shell; quotes ok). Try rollout undo or patch — see help. exit leaves solve mode.")
	fmt.Fprintf(s.Out, "Precinct policy: mutating commands must target namespace %q (including apply -f … -n %s); no -A, no namespace/node/cluster-admin nukes.\n", s.NS, s.NS)
	if len(s.Def.SolveHints) > 0 {
		fmt.Fprintln(s.Out, "")
		fmt.Fprintln(s.Out, "Case desk — angles that fit this folder:")
		for _, h := range s.Def.SolveHints {
			fmt.Fprintf(s.Out, "  • %s\n", h)
		}
	}
	s.solveMode = true
	return nil
}

func (s *Session) status() error {
	notes, err := s.Store.Notes(s.ctx, s.SessID)
	if err != nil {
		return err
	}
	fmt.Fprintln(s.Out, "CASE FILE — status")
	for _, n := range notes {
		fmt.Fprintf(s.Out, "  • [%s] %s\n", n.Kind, n.Body)
	}
	fmt.Fprintf(s.Out, "  • accused=%v hot=%v solveMode=%v\n", s.accused, s.hotAccusation, s.solveMode)
	if contacts.SeniorPath(s.Def) {
		fmt.Fprintf(s.Out, "  • senior_detective unlocked=%v hint_delivered=%v\n",
			s.inv.SeniorDetectiveUnlocked, s.inv.SeniorHintDelivered)
		fmt.Fprintf(s.Out, "  • sysadmin unlocked=%v hint_delivered=%v\n",
			s.inv.SysadminUnlocked, s.inv.SysadminHintDelivered)
		fmt.Fprintf(s.Out, "  • network_engineer unlocked=%v hint_delivered=%v\n",
			s.inv.NetworkEngineerUnlocked, s.inv.NetworkEngineerHintDelivered)
		fmt.Fprintf(s.Out, "  • archivist unlocked=%v hint_delivered=%v\n",
			s.inv.ArchivistUnlocked, s.inv.ArchivistHintDelivered)
	}
	return nil
}

func (s *Session) debrief() error {
	if err := kubectl.VictoryForDefinition(s.ctx, s.Kube, s.NS, s.Def, kubectl.DefaultVictoryTimeout); err != nil {
		fmt.Fprintln(s.Out, strings.TrimSpace(fmt.Sprintf(`
The duty sergeant slides the form back unread. The stamp stays in the drawer until the cluster stops lying on the stand.

%v

Tend the workload first — then debrief when observe would make you proud.`, err)))
		return nil
	}
	text, err := s.LLM.Debrief(s.ctx, s.Def)
	if err != nil {
		return err
	}
	fmt.Fprintln(s.Out, text)
	s.Close("solved")
	events.Emit(s.ctx, s.Emitter, events.SessionSolved, map[string]any{
		"scenario_id": string(s.Def.ID),
	})
	return errQuit
}

func (s *Session) dossier() error {
	fmt.Fprintln(s.Out, strings.TrimSpace(`
DOSSIER — pulled from the local clerk (history.db). "Cleared" means you debriefed
after the cluster passed the precinct health check for that scenario.`))
	m, err := s.Store.CaseFolderMap(s.ctx)
	if err != nil {
		return err
	}
	for _, id := range scenario.List() {
		f, ok := m[string(id)]
		if !ok {
			fmt.Fprintf(s.Out, "  %s — tab untouched\n", id)
			continue
		}
		last := f.LastOutcome
		if last == "" {
			last = "—"
		}
		fmt.Fprintf(s.Out, "  %s — opened %d×, cleared %d×, last stamp: %s\n",
			id, f.OpenCount, f.SolvedCount, last)
	}
	if contacts.SeniorPath(s.Def) {
		s.tryUnlockArchivist("dossier")
	}
	return nil
}

func (s *Session) logNote(kind, body string) error {
	return s.Store.AddNote(s.ctx, s.SessID, kind, body)
}

func (s *Session) markTraceSeen() {
	if contacts.SeniorPath(s.Def) {
		s.inv.SeenTrace = true
		s.maybeUnlockFromEvidence()
		s.tryUnlockNetworkEngineer("trace")
	}
}

func (s *Session) maybeUnlockFromEvidence() {
	if !contacts.SeniorPath(s.Def) {
		return
	}
	if !contacts.ShouldUnlockSeniorFromEvidence(&s.inv) {
		return
	}
	s.tryUnlockSeniorDetective("logs_and_trace")
}

func (s *Session) tryUnlockSeniorDetective(reason string) {
	if s.inv.SeniorDetectiveUnlocked {
		return
	}
	s.inv.SeniorDetectiveUnlocked = true
	events.Emit(s.ctx, s.Emitter, events.ContactUnlocked, map[string]any{
		"contact": string(contacts.SeniorDetective),
		"reason":  reason,
	})
	_ = s.Store.AddNote(s.ctx, s.SessID, "contact", "Senior Detective unlocked — "+reason)
}

func (s *Session) tryUnlockSysadmin(reason string) {
	if s.inv.SysadminUnlocked {
		return
	}
	s.inv.SysadminUnlocked = true
	events.Emit(s.ctx, s.Emitter, events.ContactUnlocked, map[string]any{
		"contact": string(contacts.Sysadmin),
		"reason":  reason,
	})
	_ = s.Store.AddNote(s.ctx, s.SessID, "contact", "Sysadmin unlocked — "+reason)
}

func (s *Session) tryUnlockNetworkEngineer(reason string) {
	if s.inv.NetworkEngineerUnlocked {
		return
	}
	s.inv.NetworkEngineerUnlocked = true
	events.Emit(s.ctx, s.Emitter, events.ContactUnlocked, map[string]any{
		"contact": string(contacts.NetworkEngineer),
		"reason":  reason,
	})
	_ = s.Store.AddNote(s.ctx, s.SessID, "contact", "Network Engineer unlocked — "+reason)
}

func (s *Session) tryUnlockArchivist(reason string) {
	if s.inv.ArchivistUnlocked {
		return
	}
	s.inv.ArchivistUnlocked = true
	events.Emit(s.ctx, s.Emitter, events.ContactUnlocked, map[string]any{
		"contact": string(contacts.Archivist),
		"reason":  reason,
	})
	_ = s.Store.AddNote(s.ctx, s.SessID, "contact", "Archivist unlocked — "+reason)
}

func (s *Session) hintWireRoster() error {
	fmt.Fprintln(s.Out, strings.TrimRight(contacts.WireRoster(&s.inv), "\n"))
	return nil
}

func (s *Session) hintWithTarget(which contacts.ID) error {
	switch which {
	case contacts.SeniorDetective:
		return s.hintSenior()
	case contacts.Sysadmin:
		return s.hintSysadmin()
	case contacts.NetworkEngineer:
		return s.hintNetworkEngineer()
	case contacts.Archivist:
		return s.hintArchivist()
	default:
		return fmt.Errorf("unsupported contact %q", which)
	}
}

// resolveContactWire uses HTTP LLM when configured (ContactWirer); otherwise static copy from contacts.
func (s *Session) resolveContactWire(which contacts.ID) (string, error) {
	static := contacts.StaticWireMessage(which, s.Def)
	wc, ok := s.LLM.(llm.ContactWirer)
	if !ok {
		return static, nil
	}
	out, err := wc.ContactWire(s.ctx, s.Def, string(which), static)
	if err != nil {
		if errors.Is(err, llm.ErrUseStaticWire) {
			return static, nil
		}
		return "", err
	}
	t := strings.TrimSpace(out)
	if t == "" {
		return static, nil
	}
	return t, nil
}

func (s *Session) hintSenior() error {
	if !s.inv.SeniorDetectiveUnlocked {
		fmt.Fprintln(s.Out, strings.TrimSpace(`
The wire's quiet — Senior hasn't picked up yet. Show the work: pull logs,
trace something that matters, or put a theory on the record. A lukewarm
accusation opens that line, too — nobody's grading your ego, just your evidence.`))
		return nil
	}
	if s.inv.SeniorHintDelivered {
		fmt.Fprintln(s.Out, "The Senior Detective already sent a message this case — see your case file (status).")
		return nil
	}
	text, err := s.resolveContactWire(contacts.SeniorDetective)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.Out, "%s\n", text)
	s.inv.SeniorHintDelivered = true
	events.Emit(s.ctx, s.Emitter, events.HintDelivered, map[string]any{
		"contact": string(contacts.SeniorDetective),
	})
	return s.logNote("hint", "Senior Detective — message delivered")
}

func (s *Session) hintSysadmin() error {
	if !s.inv.SysadminUnlocked {
		fmt.Fprintln(s.Out, strings.TrimSpace(`
The basement line is dead — Sysadmin won't pick up until you've described
a pod for real (examine pod <name>). They don't do vibes.`))
		return nil
	}
	if s.inv.SysadminHintDelivered {
		fmt.Fprintln(s.Out, "The sysadmin already sent a message this case — see status.")
		return nil
	}
	text, err := s.resolveContactWire(contacts.Sysadmin)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.Out, "%s\n", text)
	s.inv.SysadminHintDelivered = true
	events.Emit(s.ctx, s.Emitter, events.HintDelivered, map[string]any{
		"contact": string(contacts.Sysadmin),
	})
	return s.logNote("hint", "Sysadmin — message delivered")
}

func (s *Session) hintNetworkEngineer() error {
	if !s.inv.NetworkEngineerUnlocked {
		fmt.Fprintln(s.Out, strings.TrimSpace(`
Nobody's on the trunk line — Network won't answer until you've traced a pod
or deployment (trace <name>) so they know which junction box to curse.`))
		return nil
	}
	if s.inv.NetworkEngineerHintDelivered {
		fmt.Fprintln(s.Out, "The network engineer already sent a message this case — see status.")
		return nil
	}
	text, err := s.resolveContactWire(contacts.NetworkEngineer)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.Out, "%s\n", text)
	s.inv.NetworkEngineerHintDelivered = true
	events.Emit(s.ctx, s.Emitter, events.HintDelivered, map[string]any{
		"contact": string(contacts.NetworkEngineer),
	})
	return s.logNote("hint", "Network Engineer — message delivered")
}

func (s *Session) hintArchivist() error {
	if !s.inv.ArchivistUnlocked {
		fmt.Fprintln(s.Out, strings.TrimSpace(`
The stacks are closed — Archivist doesn't open a file until you've pulled
your dossier once this session (dossier) so they know you're not wasting
carbon paper.`))
		return nil
	}
	if s.inv.ArchivistHintDelivered {
		fmt.Fprintln(s.Out, "The Archivist already sent a message this case — see status.")
		return nil
	}
	text, err := s.resolveContactWire(contacts.Archivist)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.Out, "%s\n", text)
	s.inv.ArchivistHintDelivered = true
	events.Emit(s.ctx, s.Emitter, events.HintDelivered, map[string]any{
		"contact": string(contacts.Archivist),
	})
	return s.logNote("hint", "Archivist — message delivered")
}

func (s *Session) execKubectl(user string) ([]byte, error) {
	line := strings.TrimSpace(user)
	if line == "" {
		return nil, fmt.Errorf("empty kubectl command")
	}
	if line != "kubectl" && !strings.HasPrefix(line, "kubectl ") {
		line = "kubectl " + line
	}
	if s.Kube.Context != "" {
		line = strings.Replace(line, "kubectl ", "kubectl --context="+s.Kube.Context+" ", 1)
	}
	c := exec.CommandContext(s.ctx, "sh", "-c", line)
	c.Env = os.Environ()
	if s.Kube.Kubeconfig != "" {
		c.Env = append(c.Env, "KUBECONFIG="+s.Kube.Kubeconfig)
	}
	var buf bytes.Buffer
	c.Stdout = &buf
	c.Stderr = &buf
	if err := c.Run(); err != nil {
		out := strings.TrimSpace(buf.String())
		if out != "" {
			return buf.Bytes(), fmt.Errorf("%w\n%s", err, out)
		}
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}
