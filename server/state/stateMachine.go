package state

import (
	"log"
	"log/slog"
	"myDb/server/cfg"
	"myDb/server/logging"
	"myDb/server/utils"
	"time"
)

type StateMachine struct {
	VolatileState VolatileState
	LeaderVars    LeaderVars
	ServerVars    ServerVars
	*LogVars
	CandidateVars CandidateVars
	ServerConfig  cfg.ServerConfig
	VoteTimeout   Timer
	PingTimeout   Timer
}

type StateMachineHandler interface {
	GetServerIdString() string
	GetCurrentTerm() int64
	GetLastLogProps() (index int64, term int64)

	GetVoteFireChan() <-chan time.Time
	GetPingFireChan() <-chan time.Time

	Start()

	HandleVoteRequest(cId string, t int64, lt int64, li int64) bool
	HandleAppendEntriesRequest(t int64)
	HandleVoteResponse(cId string, g bool, t int64)
	HandleVoteTimerFired() bool
	HandlePingRequestSent()
}

func NewDefaultStateMachine(is []cfg.NodeId) StateMachine {
	return StateMachine{
		VolatileState: VolatileState{},
		LeaderVars:    NewDefaultLeaderVars(is),
		ServerVars:    NewDefaultServerVars(),
		LogVars:       NewDefaultLogVars(),
		CandidateVars: NewDefaultCandidateVars(),
		VoteTimeout:   NewTimeoutMod(TimerConf{MinMs: 15000, MaxMs: 20000, Name: "Vote"}),
		PingTimeout:   NewTimeoutMod(TimerConf{MinMs: 5000, MaxMs: 5000, Name: "Ping"}),
	}
}

func (s *StateMachine) GetServerIdString() string {
	return string(s.ServerConfig.Me.Id)
}

func (s *StateMachine) GetCurrentTerm() int64 {
	return s.ServerVars.CurrentTerm
}

func (s *StateMachine) GetVoteFireChan() <-chan time.Time {
	return s.VoteTimeout.GetFireChan()
}

func (s *StateMachine) GetPingFireChan() <-chan time.Time {
	return s.PingTimeout.GetFireChan()
}

func (s *StateMachine) Start() {
	go s.VoteTimeout.Start()
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

func (s *StateMachine) HandleAppendEntriesRequest(t int64) {
	if s.ServerVars.Role == Leader {
		s.VoteTimeout.Stop()
	} else {
		s.VoteTimeout.Reset()
	}
}

func (s *StateMachine) HandleVoteResponse(cId string, g bool, t int64) {
	s.updateTerm(t)
	if s.ServerVars.Role != Candidate {
		return
	}

	s.CandidateVars.VotesGranted[cfg.NodeId(cId)] = g

	var vFor, vAgainst int
	for _, v := range s.CandidateVars.VotesGranted {
		if v {
			vFor++
		} else {
			vAgainst++
		}
	}

	m := (len(s.ServerConfig.Servers) / 2) + 1
	if vFor >= m {
		s.becomeLeader()
	} else if vAgainst >= m {
		s.becomeFollower()
	}
}

func (s *StateMachine) HandleVoteTimerFired() bool {
	if s.ServerVars.Role == Leader {
		s.VoteTimeout.Stop()
		return false
	}

	s.updateTerm(s.ServerVars.CurrentTerm + 1)
	s.ServerVars.Role = Candidate
	s.ServerVars.VotedFor = s.ServerConfig.Me.Id
	s.CandidateVars.Reset()
	s.CandidateVars.VotesGranted[s.ServerConfig.Me.Id] = true
	return true
}

func (s *StateMachine) HandlePingRequestSent() {
	if s.ServerVars.Role == Leader {
		s.PingTimeout.Reset()
	}
}

func (s *StateMachine) becomeLeader() {
	var ss []string
	for v := range s.CandidateVars.VotesGranted {
		ss = append(ss, string(v))
	}

	slog.Info("Becoming leader", logging.MyTerm, s.ServerVars.CurrentTerm, logging.Voters, ss)

	s.ServerVars.Role = Leader
	var nextIndex int64 = 1
	if len(s.LogVars.Log) > 0 {
		nextIndex += s.LogVars.Log[len(s.LogVars.Log)-1].Index
	}

	for _, id := range s.ServerConfig.Servers {
		s.LeaderVars.MatchIndex[id.Id] = 0
		s.LeaderVars.NextIndex[id.Id] = nextIndex
	}
	s.PingTimeout.Reset()
	log.Printf("%s \n", utils.Green()("just became leader and reset ping timeout"))
	s.VoteTimeout.Stop()
	log.Printf("%s \n", utils.Green()("just became leader and reset ping timeout"))
}

func (s *StateMachine) becomeFollower() {
	s.ServerVars.Role = Follower
	s.CandidateVars.Reset()
	s.VoteTimeout.Reset()
}

func (s *StateMachine) requestVoteLogOk(lt int64, li int64) bool {

	if len(s.LogVars.Log) == 0 {
		return true
	}

	lastLog := s.LogVars.Log[len(s.LogVars.Log)-1]

	return lt > lastLog.Term || (lt == lastLog.Term && li == lastLog.Index)
}

func (s *StateMachine) updateTerm(t int64) {
	if t <= s.ServerVars.CurrentTerm {
		return
	}

	slog.Warn("My term is less than the sender. Moving to latest current:", logging.MyTerm, s.ServerVars.CurrentTerm, logging.MessageTerm, t)
	s.ServerVars.CurrentTerm = t
	s.ServerVars.VotedFor = ""
	if s.ServerVars.Role == Leader {
		s.VoteTimeout.Reset()
	}
	s.ServerVars.Role = Follower
	s.CandidateVars.Reset()
}

func (s *StateMachine) MyId() string {
	return string(s.ServerConfig.Me.Id)
}
