package state

type LogState struct {
	Log            []Log
	CommittedIndex int64
}
