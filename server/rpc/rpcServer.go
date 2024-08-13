package rpc

import (
	"context"
	"fmt"
	"myDb/server/state"
)

type RpcServer struct {
	UnimplementedRaftServiceServer
	state state.StateMachine
}

func (s *RpcServer) AppendEntries(ctx context.Context, in *AppendEntriesRequest) (*AppendEntriesResponse, error) {

	if in.Term > s.state.ServerVars.CurrentTerm {
		fmt.Println("My term is less than the sender. Moving to latest current:", s.state.ServerVars.CurrentTerm, "next:", in.Term)
		s.state.ServerVars.CurrentTerm = in.Term
	}

	logOk := in.PrevLogIndex >= 0 ||
		(in.PrevLogIndex <= int64(len(s.state.LogState.Log)) &&
			in.PrevLogTerm == int64(s.state.LogState.Log[in.PrevLogIndex].Term))

	isOk := in.Term < s.state.ServerVars.CurrentTerm ||
		(in.Term == s.state.ServerVars.CurrentTerm && !logOk && s.state.ServerVars.Role == state.Follower)

	if !isOk {
		fmt.Println("Received Append Entries from", in.LeaderId, "but log was not OK", in)
		return createAppendEntriesResponse(s.state.ServerVars.CurrentTerm, false, 0, nil)
	}

	appendEntriesToLog(s.state.LogState, in.Entries)

	return createAppendEntriesResponse(s.state.ServerVars.CurrentTerm, true, 0, nil)
}

func (s *RpcServer) RequestVote(ctx context.Context, in *RequestVoteRequest) (*RequestVoteResponse, error) {
	fmt.Printf("Received: %v", in)

	return &RequestVoteResponse{
		Term:        s.state.ServerVars.CurrentTerm,
		VoteGranted: true,
	}, nil
}

func createAppendEntriesResponse(term int64, success bool, matchIndex int64, err error) (*AppendEntriesResponse, error) {
	return &AppendEntriesResponse{
		Term:       term,
		Success:    success,
		MatchIndex: matchIndex,
	}, nil
}

func appendEntriesToLog(l state.LogState, es []*Entry) {
	for _, e := range es {
		l.Log = append(l.Log, state.Log{
			Term:    e.Term,
			Index:   e.Index,
			Command: e.Command,
		})
	}
}
