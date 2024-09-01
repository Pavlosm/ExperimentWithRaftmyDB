package state

type ServerRole int

const (
	_ ServerRole = iota
	Leader
	Follower
	Candidate
)
