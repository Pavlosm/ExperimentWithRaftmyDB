package state

type ServerVars struct {
	CurrentTerm int64
	VotedFor    string
	Role        ServerRole
}
