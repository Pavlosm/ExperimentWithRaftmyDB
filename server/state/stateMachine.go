package state

import (
	"fmt"
	"myDb/server/cfg"
)

type StateMachine struct {
	VolatileState VolatileState
	LeaderVars    LeaderVars
	ServerVars    ServerVars
	LogState      LogVars
	CandidateVars CandidateVars
	ServerConfig  cfg.ServerConfig
}

func NewDefaultStateMachine(is []cfg.NodeId) StateMachine {
	return StateMachine{
		VolatileState: VolatileState{},
		LeaderVars:    NewDefaultLeaderVars(is),
		ServerVars:    NewDefaultServerVars(),
		LogState:      NewDefaultLogVars(),
		CandidateVars: NewDefaultCandidateVars(),
	}
}

func (s *StateMachine) HandleRequestVoteResponse(n cfg.NodeId, t int64, granted bool) {
	s.updateTerm(t)
	s.CandidateVars.VotesGranted[n] = granted
	if len(s.CandidateVars.VotesGranted) >= s.ServerConfig.QuorumCount {
		s.becomeLeader()
	}
}

func (s *StateMachine) becomeLeader() {
	s.ServerVars.Role = Leader
	var nextIndex int64 = 1
	if len(s.LogState.Log) > 0 {
		nextIndex += s.LogState.Log[len(s.LogState.Log)-1].Index
	}

	for _, id := range s.ServerConfig.Servers {
		s.LeaderVars.MatchIndex[id.Id] = 0
		s.LeaderVars.NextIndex[id.Id] = nextIndex
	}
}

func (s *StateMachine) HandleVoteRequest(cId string, t int64, lt int64, li int64) bool {
	s.updateTerm(t)
	grant := t == s.ServerVars.CurrentTerm &&
		s.requestVoteLogOk(lt, li) &&
		s.ServerVars.VotedFor.IsEqualOrEmpty(cId)

	if grant {
		s.ServerVars.VotedFor = cfg.NodeId(cId)
	}

	return grant
}

func (s *StateMachine) updateTerm(t int64) {
	if t <= s.ServerVars.CurrentTerm {
		return
	}

	fmt.Println("My term is less than the sender. Moving to latest current:", s.ServerVars.CurrentTerm, "next:", t)
	s.ServerVars.CurrentTerm = t
	s.ServerVars.VotedFor = ""
	s.ServerVars.Role = Follower
	for ni := range s.CandidateVars.RequestVotes {
		delete(s.CandidateVars.RequestVotes, ni)
	}
	for ni := range s.CandidateVars.VotesGranted {
		delete(s.CandidateVars.VotesGranted, ni)
	}
}

func (s *StateMachine) requestVoteLogOk(lt int64, li int64) bool {

	if len(s.LogState.Log) == 0 {
		return true
	}

	lastLog := s.LogState.Log[len(s.LogState.Log)-1]

	return lt > lastLog.Term || (lt == lastLog.Term && li == lastLog.Index)
}
