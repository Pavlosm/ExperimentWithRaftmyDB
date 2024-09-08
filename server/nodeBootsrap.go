package main

import (
	"myDb/server/cfg"
	"myDb/server/rpcServer"
	"myDb/server/state"
)

func NewRpcServer(sc cfg.ServerConfig, c *FlowControlChannels) (*rpcServer.RpcServer, *state.StateMachine) {
	sv := make([]cfg.NodeId, len(sc.Servers))
	for _, id := range sc.Servers {
		sv = append(sv, id.Id)
	}
	s := state.NewDefaultStateMachine(sv)
	s.ServerConfig = sc
	return &rpcServer.RpcServer{
		VoteCmdChan:          c.VoteCmd,
		AppendEntriesCmdChan: c.AppendEntriesCmd,
	}, &s
}
