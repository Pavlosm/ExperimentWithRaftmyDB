package rpcServer

import (
	"context"
	"myDb/rpc"
	"myDb/server/environment"
	"myDb/server/utils"
)

type RpcServer struct {
	rpc.UnimplementedRaftServiceServer
	env *environment.Env
}

func (s *RpcServer) SendCommand(ctx context.Context, in *rpc.CommandRequest) (*rpc.CommandResponse, error) {

	req := utils.WithReplyChan[*rpc.CommandRequest, *rpc.CommandResponse]{
		Data:  in,
		Reply: make(chan *rpc.CommandResponse),
	}

	s.env.Channels.CommandCmd <- req

	r := <-req.Reply
	return r, nil
}
func (s *RpcServer) AppendEntries(ctx context.Context, in *rpc.AppendEntriesRequest) (*rpc.AppendEntriesResponse, error) {

	req := utils.WithReplyChan[*rpc.AppendEntriesRequest, *rpc.AppendEntriesResponse]{
		Data:  in,
		Reply: make(chan *rpc.AppendEntriesResponse),
	}

	s.env.Channels.AppendEntriesCmd <- req

	r := <-req.Reply

	return r, nil
}

func (s *RpcServer) RequestVote(ctx context.Context, in *rpc.RequestVoteRequest) (*rpc.RequestVoteResponse, error) {

	req := utils.WithReplyChan[*rpc.RequestVoteRequest, *rpc.RequestVoteResponse]{
		Data:  in,
		Reply: make(chan *rpc.RequestVoteResponse),
	}

	s.env.Channels.VoteCmd <- req
	r := <-req.Reply
	return r, nil
}

func NewRpcServer(env *environment.Env) *RpcServer {
	return &RpcServer{
		env: env,
	}
}
