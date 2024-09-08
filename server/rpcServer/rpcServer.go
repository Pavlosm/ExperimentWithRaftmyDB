package rpcServer

import (
	"context"
	"log"
	"myDb/server/rpc"
	"myDb/server/utils"
)

type RpcServer struct {
	rpc.UnimplementedRaftServiceServer
	VoteCmdChan          chan<- utils.WithReplyChan[*rpc.RequestVoteRequest, *rpc.RequestVoteResponse]
	AppendEntriesCmdChan chan<- utils.WithReplyChan[*rpc.AppendEntriesRequest, *rpc.AppendEntriesResponse]
}

func (s *RpcServer) AppendEntries(ctx context.Context, in *rpc.AppendEntriesRequest) (*rpc.AppendEntriesResponse, error) {

	log.Println("RPC Server: received request vote request")

	req := utils.WithReplyChan[*rpc.AppendEntriesRequest, *rpc.AppendEntriesResponse]{
		Data:  in,
		Reply: make(chan *rpc.AppendEntriesResponse),
	}

	s.AppendEntriesCmdChan <- req

	log.Println("RPC Server: sent to channel")

	r := <-req.Reply

	log.Println("RPC Server: got response from channel")

	return r, nil

	// TODO send to out channel
	// TODO wait for response

	// return &rpc.AppendEntriesResponse{
	// 	Term:       s.State.GetCurrentTerm(),
	// 	Success:    true,
	// 	MatchIndex: 1,
	// }, nil
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

	// log.Println("RPC Server: received request vote request")
	req := utils.WithReplyChan[*rpc.RequestVoteRequest, *rpc.RequestVoteResponse]{
		Data:  in,
		Reply: make(chan *rpc.RequestVoteResponse),
	}

	s.VoteCmdChan <- req
	// log.Println("RPC Server: sent to channel")
	r := <-req.Reply
	// log.Println("RPC Server: got response from channel")
	return r, nil
}
