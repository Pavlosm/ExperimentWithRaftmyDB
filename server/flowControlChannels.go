package main

import (
	"myDb/server/rpc"
	"myDb/server/utils"
)

type FlowControlChannels struct {
	VoteCmdChan          chan utils.WithReplyChan[*rpc.RequestVoteRequest, *rpc.RequestVoteResponse]
	AppendEntriesCmdChan chan utils.WithReplyChan[*rpc.AppendEntriesRequest, *rpc.AppendEntriesResponse]
}

func NewFlowControlChannels() FlowControlChannels {
	return FlowControlChannels{
		VoteCmdChan:          make(chan utils.WithReplyChan[*rpc.RequestVoteRequest, *rpc.RequestVoteResponse]),
		AppendEntriesCmdChan: make(chan utils.WithReplyChan[*rpc.AppendEntriesRequest, *rpc.AppendEntriesResponse]),
	}
}
