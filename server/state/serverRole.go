package state

type ServerRole int

const (
	Leader ServerRole = iota
	Follower
	Candidate
)
