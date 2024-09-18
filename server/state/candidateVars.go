package state

import "myDb/server/cfg"

type CandidateVars struct {
	RequestVotes map[cfg.NodeId]bool
	VotesGranted map[cfg.NodeId]bool
}

func NewDefaultCandidateVars() CandidateVars {
	return CandidateVars{
		RequestVotes: make(map[cfg.NodeId]bool),
		VotesGranted: make(map[cfg.NodeId]bool),
	}
}

func (c CandidateVars) Reset() {
	for ni := range c.RequestVotes {
		delete(c.RequestVotes, ni)
	}

	for ni := range c.VotesGranted {
		delete(c.VotesGranted, ni)
	}
}
