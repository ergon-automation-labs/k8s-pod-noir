package contacts

import "strings"

// WireRoster prints who's locked, open, or already spoke — for `hint` with no arguments.
func WireRoster(inv *InvestigationState) string {
	if inv == nil {
		inv = &InvestigationState{}
	}
	var b strings.Builder
	b.WriteString("WIRE ROOM — who's on the line\n\n")
	b.WriteString(rowRoster("Senior Detective", "senior", inv.SeniorDetectiveUnlocked, inv.SeniorHintDelivered,
		"logs + trace, or a cold/warm accuse"))
	b.WriteString(rowRoster("Sysadmin", "sysadmin", inv.SysadminUnlocked, inv.SysadminHintDelivered,
		"examine pod <name>"))
	b.WriteString(rowRoster("Network Engineer", "network", inv.NetworkEngineerUnlocked, inv.NetworkEngineerHintDelivered,
		"trace <name>"))
	b.WriteString(rowRoster("Archivist", "archivist", inv.ArchivistUnlocked, inv.ArchivistHintDelivered,
		"dossier (this session)"))
	b.WriteString("\nTo take a call: hint senior | hint sysadmin | hint network | hint archivist\n")
	return strings.TrimRight(b.String(), "\n")
}

func rowRoster(title, hintCmd string, unlocked, delivered bool, lockHelp string) string {
	switch {
	case !unlocked:
		return "  " + padTo(title, 22) + "locked — " + lockHelp + "\n"
	case delivered:
		return "  " + padTo(title, 22) + "done (already spoke this case)\n"
	default:
		return "  " + padTo(title, 22) + "open — type: hint " + hintCmd + "\n"
	}
}

func padTo(s string, w int) string {
	if len(s) >= w {
		return s + " "
	}
	return s + strings.Repeat(" ", w-len(s))
}
