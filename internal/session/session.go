package session

import (
	"bufio"
	"bytes"
	"context"
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
	return s, nil
}

func (s *Session) Close(outcome string) {
	s.closeOnce.Do(func() {
		if s.cleanupFn != nil {
			s.cleanupFn()
		}
		bg := context.Background()
		_ = s.Store.EndSession(bg, s.SessID, outcome)
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
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if err := s.handle(line); err != nil {
			if err == errQuit {
				return nil
			}
			fmt.Fprintln(s.Out, "error:", err)
		}
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
			fmt.Fprintln(s.Out, precinct.CabinetPeek(s.Def))
			return nil
		case "quit":
			s.Close("abandoned")
			events.Emit(s.ctx, s.Emitter, events.SessionAbandoned, map[string]any{"last_command": line})
			return errQuit
		}
		if err := kubectl.EnsureMutatingUsesGameNamespace(line, s.NS); err != nil {
			return err
		}
		out, err := s.execKubectl(line)
		if err != nil {
			return err
		}
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
  cabinet, files                glance back at the file drawer (other cases)
  debrief                         after resolution — close the case (mock)
  hint                            Senior Detective when unlocked (logs+trace or non-HOT accuse)
  quit                            exit; namespace is cleared (unless -skip-cleanup)

Solve mode (examples):
  case 001: kubectl rollout undo deployment/payments-worker -n pod-noir
  case 002: kubectl create secret generic ledger-signing-secret -n pod-noir ...
  case 003: kubectl set image deployment/shipping-notifier notifier=busybox:1.36.1 -n pod-noir
  kubectl patch ... --type=json OR --type=strategic  (see debrief)`))
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
	case low == "hint":
		return s.hint()
	case low == "cabinet" || low == "files":
		fmt.Fprintln(s.Out, precinct.CabinetPeek(s.Def))
		return nil
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
	return s.logNote("examine", "Described pod "+name)
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
		fmt.Fprintf(s.Out, "  • senior_detective_unlocked=%v hint_delivered=%v\n",
			s.inv.SeniorDetectiveUnlocked, s.inv.SeniorHintDelivered)
	}
	return nil
}

func (s *Session) debrief() error {
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

func (s *Session) logNote(kind, body string) error {
	return s.Store.AddNote(s.ctx, s.SessID, kind, body)
}

func (s *Session) markTraceSeen() {
	if contacts.SeniorPath(s.Def) {
		s.inv.SeenTrace = true
		s.maybeUnlockFromEvidence()
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

func (s *Session) hint() error {
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
	fmt.Fprintln(s.Out, strings.TrimSpace(contacts.SeniorDetectiveMessage(s.Def)))
	s.inv.SeniorHintDelivered = true
	events.Emit(s.ctx, s.Emitter, events.HintDelivered, map[string]any{
		"contact": string(contacts.SeniorDetective),
	})
	return s.logNote("hint", "Senior Detective — stub message delivered")
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
