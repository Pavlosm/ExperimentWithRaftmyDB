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

func (c CandidateVars) Reset(t int64) {
	c.Term = t

	for ni := range c.RequestVotes {
		delete(c.RequestVotes, ni)
	}

	for ni := range c.RequestVotes {
		delete(c.VotesGranted, ni)
	}
}
