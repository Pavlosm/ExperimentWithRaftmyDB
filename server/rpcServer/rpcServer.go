package rpcServer

import (
	"context"
	"log"
	"myDb/server/rpc"
	"myDb/server/state"
)

type RpcServer struct {
	rpc.UnimplementedRaftServiceServer
	State                   state.Stater
	UpdateChan              chan int
	RequestVoteResponseChan <-chan *rpc.RequestVoteResponse
	RequestVoteRequestChan  chan<- *rpc.RequestVoteRequest
}

func (s *RpcServer) AppendEntries(ctx context.Context, in *rpc.AppendEntriesRequest) (*rpc.AppendEntriesResponse, error) {

	// TODO send to out channel
	// TODO wait for response

	return &rpc.AppendEntriesResponse{
		Term:       s.State.GetCurrentTerm(),
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

	log.Println("RPC Server: received request vote request")
	s.RequestVoteRequestChan <- in
	log.Println("RPC Server: sent to channel")
	r := <-s.RequestVoteResponseChan
	log.Println("RPC Server: got response from channel")
	return r, nil
	// granted := s.State.HandleVoteRequest(
	// 	in.CandidateId,
	// 	in.Term,
	// 	in.LastLogTerm,
	// 	in.LastLogIndex)

	// return &rpc.RequestVoteResponse{
	// 	Term:        s.State.GetCurrentTerm(),
	// 	VoteGranted: true,
	// }, nil
}

func (s *RpcServer) createAppendEntriesResponse(success bool, matchIndex int64) (*rpc.AppendEntriesResponse, error) {
	return &rpc.AppendEntriesResponse{
		Term:       s.State.GetCurrentTerm(),
		Success:    success,
		MatchIndex: matchIndex,
	}, nil
}
