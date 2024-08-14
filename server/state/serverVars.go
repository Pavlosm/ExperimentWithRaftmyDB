package state

import "myDb/server/cfg"

type ServerVars struct {
	CurrentTerm int64
	VotedFor    cfg.NodeId
	Role        ServerRole
}

func NewDefaultServerVars() ServerVars {
	return ServerVars{
		CurrentTerm: 1,
		VotedFor:    cfg.NodeId(""),
		Role:        Follower,
	}
}
