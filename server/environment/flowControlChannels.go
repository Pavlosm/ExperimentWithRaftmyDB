package environment

import (
	"myDb/server/rpc"
	"myDb/server/utils"
)

type FlowControlChannels struct {
	VoteCmd            chan utils.WithReplyChan[*rpc.RequestVoteRequest, *rpc.RequestVoteResponse]
	AppendEntriesCmd   chan utils.WithReplyChan[*rpc.AppendEntriesRequest, *rpc.AppendEntriesResponse]
	VoteReply          chan *rpc.RequestVoteResponse
	AppendEntriesReply chan *rpc.AppendEntriesResponse
	Err                chan error
}

func NewFlowControlChannels() FlowControlChannels {
	return FlowControlChannels{
		VoteCmd:            make(chan utils.WithReplyChan[*rpc.RequestVoteRequest, *rpc.RequestVoteResponse]),
		AppendEntriesCmd:   make(chan utils.WithReplyChan[*rpc.AppendEntriesRequest, *rpc.AppendEntriesResponse]),
		VoteReply:          make(chan *rpc.RequestVoteResponse),
		AppendEntriesReply: make(chan *rpc.AppendEntriesResponse),
		Err:                make(chan error),
	}
}
