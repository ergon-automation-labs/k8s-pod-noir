package contacts

import "podnoir/internal/scenario"

// StaticWireMessage returns authored wire copy (used when LLM is mock/disabled or as anchor for HTTP).
func StaticWireMessage(id ID, def *scenario.Definition) string {
	switch id {
	case SeniorDetective:
		return SeniorDetectiveMessage(def)
	case Sysadmin:
		return SysadminMessage(def)
	case NetworkEngineer:
		return NetworkEngineerMessage(def)
	case Archivist:
		return ArchivistMessage(def)
	default:
		return ""
	}
}
