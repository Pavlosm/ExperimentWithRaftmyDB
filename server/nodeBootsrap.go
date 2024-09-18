package main

import (
	"myDb/server/cfg"
	"myDb/server/environment"
	"myDb/server/rpcServer"
	"myDb/server/state"
)

func NewRpcServer(sc cfg.ServerConfig, e *environment.Env) (*rpcServer.RpcServer, *state.StateMachine) {
	sv := make([]cfg.NodeId, len(sc.Servers))
	for _, id := range sc.Servers {
		sv = append(sv, id.Id)
	}
	s := state.NewDefaultStateMachine(sv)
	s.ServerConfig = sc
	return &rpcServer.RpcServer{
		VoteCmdChan:          e.Channels.VoteCmd,
		AppendEntriesCmdChan: e.Channels.AppendEntriesCmd,
	}, &s
}
