package state

import "myDb/server/cfg"

type CandidateVars struct {
	Term         int64
	RequestVotes map[cfg.NodeId]bool
	VotesGranted map[cfg.NodeId]bool
}

func NewDefaultCandidateVars() CandidateVars {
	return CandidateVars{
		RequestVotes: make(map[cfg.NodeId]bool),
		VotesGranted: make(map[cfg.NodeId]bool),
	}
}
