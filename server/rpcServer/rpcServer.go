package rpcServer

import (
	"context"
	"myDb/server/rpc"
	"myDb/server/state"
)

type RpcServer struct {
	rpc.UnimplementedRaftServiceServer
	State state.StateMachine
}

func (s *RpcServer) AppendEntries(ctx context.Context, in *rpc.AppendEntriesRequest) (*rpc.AppendEntriesResponse, error) {

	return &rpc.AppendEntriesResponse{
		Term:       s.State.ServerVars.CurrentTerm,
		Success:    true,
		MatchIndex: 1,
	}, nil
	// s.updateTerm(in.Term)

	// logOk := in.PrevLogIndex >= 0 ||
	// 	(in.PrevLogIndex <= int64(len(s.state.LogState.Log)) &&
	// 		in.PrevLogTerm == int64(s.state.LogState.Log[in.PrevLogIndex].Term))

	// isOk := in.Term < s.state.ServerVars.CurrentTerm ||
	// 	(in.Term == s.state.ServerVars.CurrentTerm && !logOk && s.state.ServerVars.Role == state.Follower)

	// if !isOk {
	// 	fmt.Println("Received Append Entries from", in.LeaderId, "but log was not OK", in)
	// 	return s.createAppendEntriesResponse(false, 0)
	// }

	// appendEntriesToLog(s.state.LogState, in.Entries)

	// return s.createAppendEntriesResponse(true, 0)
}

func (s *RpcServer) RequestVote(ctx context.Context, in *rpc.RequestVoteRequest) (*rpc.RequestVoteResponse, error) {

	granted := s.State.HandleVoteRequest(
		in.CandidateId,
		in.Term,
		in.LastLogTerm,
		in.LastLogIndex)

	return &rpc.RequestVoteResponse{
		Term:        s.State.ServerVars.CurrentTerm,
		VoteGranted: granted,
	}, nil
}

func (s *RpcServer) createAppendEntriesResponse(success bool, matchIndex int64) (*rpc.AppendEntriesResponse, error) {
	return &rpc.AppendEntriesResponse{
		Term:       s.State.ServerVars.CurrentTerm,
		Success:    success,
		MatchIndex: matchIndex,
	}, nil
}

func appendEntriesToLog(l state.LogVars, es []*rpc.Entry) {
	for _, e := range es {
		l.Log = append(l.Log, state.Log{
			Term:    e.Term,
			Index:   e.Index,
			Command: e.Command,
		})
	}
}
