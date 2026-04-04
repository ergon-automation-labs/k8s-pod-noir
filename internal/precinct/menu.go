package precinct

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"podnoir/internal/scenario"
	"podnoir/internal/store"
)

// Intro is the first-day / file cabinet framing before case selection.
func Intro(detective string) string {
	if strings.TrimSpace(detective) == "" {
		detective = "Detective"
	}
	return strings.TrimSpace(fmt.Sprintf(`
The city never sleeps; the cluster only blinks.

================================================================
   THE CLUSTER AGENCY  —  Municipal Adjuster Unit (training)
================================================================

You already know what objects are called. Here you learn how they fail
when a manifest promises one thing and the node files a different story.

First day on the floor. Ozone and cooling fans, bitter coffee going
cold beside a rotary phone that nobody dares answer. Oak filing cabinet
across the aisle: brass green with age, label crooked — INCIDENT FOLDERS.

Pinned note, blue pencil: "Welcome, %s. Take a folder. Names lie;
pods and events don't."

The drawer groans when you touch it.
`, detective))
}

func folderStamp(st *store.Store, id scenario.ID) string {
	if st == nil {
		return ""
	}
	f, err := st.FolderOrZero(context.Background(), string(id))
	if err != nil || f.OpenCount == 0 {
		return ""
	}
	return fmt.Sprintf("dossier: cleared %d× / opened %d×", f.SolvedCount, f.OpenCount)
}

// FileCabinetList lists scenarios; st may be nil (no dossier stamps).
func FileCabinetList(st *store.Store) string {
	var b strings.Builder
	fmt.Fprintf(&b, "┌─────────────────────────────────────────────────────────────┐\n")
	boxInner(&b, "OPEN FOLDERS — incident intake (training floor)")
	fmt.Fprintf(&b, "├─────────────────────────────────────────────────────────────┤\n")
	for i, id := range scenario.List() {
		def, err := scenario.ByID(id)
		if err != nil {
			continue
		}
		boxInner(&b, fmt.Sprintf("[%d]  %s", i+1, def.Title))
		boxInner(&b, "     "+def.FolderTease)
		boxInner(&b, "     "+string(id))
		if stamp := folderStamp(st, id); stamp != "" {
			boxInner(&b, "     "+stamp)
		}
		boxInner(&b, "")
	}
	fmt.Fprintf(&b, "└─────────────────────────────────────────────────────────────┘\n")
	return strings.TrimRight(b.String(), "\n")
}

func boxInner(b *strings.Builder, text string) {
	const w = 59
	if len(text) > w {
		text = text[:w-3] + "..."
	}
	fmt.Fprintf(b, "│  %s%s│\n", text, strings.Repeat(" ", w-len(text)))
}

// LeavingCopy when the player backs out at the cabinet.
func LeavingCopy() string {
	return strings.TrimSpace(`
You ease the drawer shut. The latch finds its teeth — small, final sound.
Corridor lights buzz. Somewhere a control plane keeps score you'll learn
to read another night.`)
}

// CabinetPeek reminds the player of other cases while one is open.
func CabinetPeek(open *scenario.Definition, st *store.Store) string {
	var b strings.Builder
	fmt.Fprintln(&b, strings.TrimSpace(`
You run a thumb along the cabinet — same metal, same dust. The other
folders are still here; you've just got this one splayed on the blotter.`))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, FileCabinetList(st))
	fmt.Fprintf(&b, "\nOpen now: %s  (%s)\n", open.Title, open.ID)
	fmt.Fprintln(&b, strings.TrimSpace(`
To choose a different case, quit — you'll step back to the cabinet (namespace tears down unless -skip-cleanup).`))
	return strings.TrimSpace(b.String())
}

// SelectCase runs the file cabinet until the player picks a valid scenario or quits.
func SelectCase(in io.Reader, out io.Writer, detective string, st *store.Store) (scenario.ID, error) {
	fmt.Fprintln(out, Intro(detective))
	fmt.Fprintln(out)
	fmt.Fprintln(out, FileCabinetList(st))
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Pick a folder [1–%d], paste a case id, or walk away before the phone rings.\n\n", len(scenario.List()))

	sc := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, "cabinet> ")
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				return "", err
			}
			return "", scenario.ErrMenuQuit
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		id, err := scenario.MatchCaseChoice(line)
		if err != nil {
			if errors.Is(err, scenario.ErrMenuQuit) {
				return "", scenario.ErrMenuQuit
			}
			fmt.Fprintf(out, "  %v\n", err)
			continue
		}
		return id, nil
	}
}
