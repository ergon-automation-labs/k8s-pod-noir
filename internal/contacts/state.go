package contacts

// ID identifies a noir contact (NPC).
type ID string

const (
	SeniorDetective ID = "senior_detective"
)

// InvestigationState tracks evidence and unlocks for hint pipeline (per session).
type InvestigationState struct {
	SeniorDetectiveUnlocked bool
	SeniorHintDelivered     bool

	// SeenLogs + SeenTrace unlock the Senior Detective without a non-hot accusation (all MVP cases).
	SeenLogs  bool
	SeenTrace bool
}
