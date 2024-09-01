package main

import (
	"myDb/server/cfg"
	"myDb/server/rpc"
	"myDb/server/rpcServer"
	"myDb/server/state"
)

func NewRpcServer(sc cfg.ServerConfig, c chan int, r chan *rpc.RequestVoteRequest, res chan *rpc.RequestVoteResponse) (*rpcServer.RpcServer, state.StateMachine) {
	sv := make([]cfg.NodeId, len(sc.Servers))
	for _, id := range sc.Servers {
		sv = append(sv, id.Id)
	}
	s := state.NewDefaultStateMachine(sv)
	s.ServerConfig = sc
	return &rpcServer.RpcServer{
		State:                   &s,
		UpdateChan:              c,
		RequestVoteResponseChan: res,
		RequestVoteRequestChan:  r,
	}, s
}
