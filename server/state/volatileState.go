package state

type VolatileState struct {
	CommittedIndex int64
	LastApplied    int64
}
