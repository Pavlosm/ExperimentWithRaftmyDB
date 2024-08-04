package cfg

type ServerRole int

const (
	Leader ServerRole = iota
	Follower
	Candidate
)
