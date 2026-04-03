package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"podnoir/internal/config"
	"podnoir/internal/events"
	"podnoir/internal/kubectl"
	"podnoir/internal/llm"
	"podnoir/internal/precinct"
	"podnoir/internal/scenario"
	"podnoir/internal/session"
	"podnoir/internal/settings"
	"podnoir/internal/store"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "doctor" {
		runDoctor(os.Args[2:])
		return
	}

	cfgPath := flag.String("config", "", "path to config YAML (optional)")
	dataDir := flag.String("data-dir", "", "app data directory for sqlite (default ~/.pod-noir)")
	scenID := flag.String("scenario", "", "omit for the file cabinet menu, or: case-001-overnight-shift | case-002-ghost-credential | case-003-dead-letter-harbour")
	skipCleanup := flag.Bool("skip-cleanup", false, "do not delete the scenario namespace on exit")
	flag.Parse()

	ctx := context.Background()
	env := settings.FromEnv()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		die(err)
	}

	dir := *dataDir
	if dir == "" {
		home, herr := os.UserHomeDir()
		if herr != nil {
			die(herr)
		}
		dir = filepath.Join(home, ".pod-noir")
	}

	stor, err := store.Open(dir)
	if err != nil {
		die(err)
	}
	defer stor.Close()

	var chosen scenario.ID
	if strings.TrimSpace(*scenID) != "" {
		chosen = scenario.ID(strings.TrimSpace(*scenID))
		if _, err := scenario.ByID(chosen); err != nil {
			die(err)
		}
	} else {
		var merr error
		chosen, merr = precinct.SelectCase(os.Stdin, os.Stdout, cfg.Detective.Name)
		if merr != nil {
			if errors.Is(merr, scenario.ErrMenuQuit) {
				fmt.Fprintln(os.Stdout, precinct.LeavingCopy())
				return
			}
			die(merr)
		}
		fmt.Fprintln(os.Stdout, "\nYou slide the folder free. Paper cuts and possibility.\n")
	}

	def, err := scenario.ByID(chosen)
	if err != nil {
		die(err)
	}

	ns := cfg.Cluster.Namespace
	if ns == "" {
		ns = "pod-noir"
	}

	kubePath, kubeCleanup, err := kubectl.ResolveKubeconfigPath()
	if err != nil {
		die(fmt.Errorf("kubeconfig: %w", err))
	}
	defer kubeCleanup()

	k := &kubectl.Runner{
		Kubeconfig: kubePath,
		Context:    cfg.Cluster.Context,
	}

	cleanupNS := func() {
		if *skipCleanup {
			return
		}
		_ = k.DeleteNamespace(context.Background(), ns)
	}

	emitter, err := events.NewEmitterFromSettings(events.SettingsInput{
		EventsAdapter:          env.EventsAdapter,
		NATSURL:                env.NATSURL,
		NATSPrefix:             env.NATSPrefix,
		NATSBridge:             env.NATSBridge,
		NATSBridgeEventsPrefix: env.NATSBridgeEventsPrefix,
	})
	if err != nil {
		die(err)
	}
	defer func() { _ = emitter.Close() }()

	llmProv, err := llm.NewFromSettings(env)
	if err != nil {
		die(err)
	}

	sess, err := session.New(ctx, k, stor, def, cfg.Detective.Name, ns, os.Stdout, os.Stdin, cleanupNS, emitter, llmProv)
	if err != nil {
		die(err)
	}
	defer sess.Close("stopped")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		sess.Close("interrupted")
		_ = stor.Close()
		os.Exit(130)
	}()

	if err := sess.RunREPL(); err != nil {
		die(err)
	}
}

func runDoctor(args []string) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	cfgPath := fs.String("config", "", "path to config YAML (optional)")
	if err := fs.Parse(args); err != nil {
		die(err)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		die(err)
	}

	path, cleanup, err := kubectl.ResolveKubeconfigPath()
	if err != nil {
		die(fmt.Errorf("kubeconfig: %w", err))
	}
	defer cleanup()

	inDocker := os.Getenv("POD_NOIR_KUBE_IN_DOCKER") == "true" || os.Getenv("POD_NOIR_KUBE_IN_DOCKER") == "1"
	host := strings.TrimSpace(os.Getenv("POD_NOIR_KUBE_API_HOST"))
	if host == "" {
		host = "host.docker.internal"
	}
	fmt.Fprintf(os.Stdout, "pod-noir doctor\n")
	fmt.Fprintf(os.Stdout, "  POD_NOIR_KUBE_IN_DOCKER=%v\n", inDocker)
	fmt.Fprintf(os.Stdout, "  POD_NOIR_KUBE_API_HOST=%q (used when rewriting localhost)\n", host)
	tlsInsecure := strings.TrimSpace(os.Getenv("POD_NOIR_KUBE_TLS_INSECURE")) != "false"
	fmt.Fprintf(os.Stdout, "  POD_NOIR_KUBE_TLS_INSECURE=%v (set false only if your cert includes the Docker host name)\n", tlsInsecure)
	fmt.Fprintf(os.Stdout, "  effective kubeconfig: %s\n", path)
	fmt.Fprintf(os.Stdout, "  context: %q\n", cfg.Cluster.Context)

	k := &kubectl.Runner{Kubeconfig: path, Context: cfg.Cluster.Context}
	ctx := context.Background()

	if out, err := k.Run(ctx, "version", "--client"); err == nil {
		// First line is enough for older kubectl (no --short in Alpine kubectl 1.29)
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 {
			fmt.Fprintf(os.Stdout, "  kubectl client: %s\n", lines[0])
		}
	} else {
		fmt.Fprintf(os.Stderr, "  kubectl client: %v\n", err)
	}

	out, err := k.Run(ctx, "cluster-info", "--request-timeout=10s")
	if err != nil {
		die(fmt.Errorf("cluster unreachable: %w\n\nHints:\n  • With Docker, keep POD_NOIR_KUBE_IN_DOCKER=true and POD_NOIR_KUBE_TLS_INSECURE unset (default: skip TLS verify for rewritten host).\n  • Rancher Desktop / Lima certs usually list localhost, not host.docker.internal — skipping verify is expected for local dev.\n  • On Linux, set POD_NOIR_KUBE_API_HOST to your host gateway if needed.", err))
	}
	fmt.Fprintf(os.Stdout, "  cluster:\n%s\n", string(out))
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
