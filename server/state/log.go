package state

type Log struct {
	Term    int64
	Index   int64
	Command string
}
