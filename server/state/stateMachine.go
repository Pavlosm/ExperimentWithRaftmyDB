package state

import (
	"log/slog"
	"myDb/server/cfg"
	"myDb/server/logging"
	"time"
)

type StateMachine struct {
	VolatileState VolatileState
	LeaderVars    LeaderVars
	ServerVars    ServerVars
	*LogVars
	CandidateVars   CandidateVars
	ServerConfig    cfg.ServerConfig
	VoteTimeout     Timer
	PingTimeout     Timer
	StateRepository StateRepositoryHandler
}

type StateMachineHandler interface {
	GetServerIdString() string
	GetCurrentTerm() int64
	GetLastLogProps() (index int64, term int64)
	GetLeaderId() string
	GetServerRole() ServerRole

	GetVoteFireChan() <-chan time.Time
	GetPingFireChan() <-chan time.Time

	Start()

	HandleVoteRequest(cId string, t int64, lt int64, li int64) bool
	HandleAppendEntriesRequest(t int64, cId string, logs []Log)
	HandleVoteResponse(cId string, g bool, t int64)
	HandleVoteTimerFired() bool
	HandlePingRequestSent()
	HandleCommandRequest(cmd string) (Log, error)
}

func NewDefaultStateMachine(sc cfg.ServerConfig) *StateMachine {
	sv := make([]cfg.NodeId, len(sc.Servers))

	for _, id := range sc.Servers {
		sv = append(sv, id.Id)
	}

	return &StateMachine{
		VolatileState:   VolatileState{},
		LeaderVars:      NewDefaultLeaderVars(sv),
		ServerVars:      NewDefaultServerVars(),
		LogVars:         NewDefaultLogVars(),
		CandidateVars:   NewDefaultCandidateVars(),
		VoteTimeout:     NewTimeoutMod(TimerConf{MinMs: 2000, MaxMs: 6000, Name: "Vote"}, 30*time.Second),
		PingTimeout:     NewTimeoutMod(TimerConf{MinMs: 1000, MaxMs: 1000, Name: "Ping"}, 48*time.Hour),
		ServerConfig:    sc,
		StateRepository: InitStateRepository(sc),
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

func (s *StateMachine) GetLeaderId() string {
	return string(s.ServerVars.LeaderId)
}

func (s *StateMachine) GetServerRole() ServerRole {
	return s.ServerVars.Role
}

func (s *StateMachine) Start() {
	l, err := s.StateRepository.ReadEntries()
	if err != nil {
		slog.Error("Error reading entries", "error", err)
		panic(err)
	}
	s.LogVars.Log = l

	go s.VoteTimeout.Start()
}

func (s *StateMachine) HandleVoteRequest(cId string, t int64, lt int64, li int64) bool {

	s.updateTerm(t, cId)

	grant := t == s.ServerVars.CurrentTerm &&
		s.requestVoteLogOk(lt, li) &&
		s.ServerVars.VotedFor.IsEqualOrEmpty(cId)

	if grant {
		s.ServerVars.VotedFor = cfg.NodeId(cId)
	}

	return grant
}

func (s *StateMachine) HandleAppendEntriesRequest(t int64, cId string, logs []Log) {

	s.updateTerm(t, cId)

	if s.ServerVars.Role == Leader {
		s.VoteTimeout.Stop()
		slog.Info("received append entries request from a follower, ignoring. Stopping vote timeout", logging.From, cId, logging.MyTerm, s.ServerVars.CurrentTerm)
		return
	}

	s.ServerVars.LeaderId = cfg.NodeId(cId)
	s.VoteTimeout.Reset()

	if len(logs) > 0 {
		s.StateRepository.WriteEntries(logs...)
	}
}

func (s *StateMachine) HandleVoteResponse(cId string, g bool, t int64) {
	s.updateTerm(t, cId)
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

	s.updateTerm(s.ServerVars.CurrentTerm+1, s.MyId())
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
	s.ServerVars.Role = Leader
	s.ServerVars.LeaderId = s.ServerConfig.Me.Id
	var nextIndex int64 = 1
	if len(s.LogVars.Log) > 0 {
		nextIndex += s.LogVars.Log[len(s.LogVars.Log)-1].Index
	}

	for _, id := range s.ServerConfig.Servers {
		s.LeaderVars.MatchIndex[id.Id] = 0
		s.LeaderVars.NextIndex[id.Id] = nextIndex
	}
	s.PingTimeout.Reset()
	s.VoteTimeout.Stop()

	var ss []string
	for v := range s.CandidateVars.VotesGranted {
		ss = append(ss, string(v))
	}
	slog.Info("just became leader. Stopped vote timeout and reset ping timeout", logging.MyTerm, s.ServerVars.CurrentTerm, logging.Voters, ss)

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

func (s *StateMachine) updateTerm(t int64, sId string) {
	if t <= s.ServerVars.CurrentTerm {
		return
	}

	slog.Warn("My term is less than the sender. Moving to latest current:", logging.From, sId, logging.MyTerm, s.ServerVars.CurrentTerm, logging.MessageTerm, t)
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

func (s *StateMachine) HandleCommandRequest(cmd string) (Log, error) {

	i, t := s.LogVars.GetLastLogProps()

	l := Log{
		Index:   i + 1,
		Term:    t,
		Command: cmd,
	}
	err := s.StateRepository.WriteEntries(l)
	if err != nil {
		return Log{}, err
	}

	return l, nil
}
