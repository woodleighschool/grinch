package policies

// Status represents the state of a machine's policy assignment.
type Status int16

const (
	StatusUnassigned Status = 0
	StatusPending    Status = 1
	StatusUpToDate   Status = 2
)
