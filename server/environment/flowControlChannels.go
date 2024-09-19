package environment

import (
	"myDb/rpc"
	"myDb/server/utils"
)

type FlowControlChannels struct {
	VoteCmd            chan utils.WithReplyChan[*rpc.RequestVoteRequest, *rpc.RequestVoteResponse]
	AppendEntriesCmd   chan utils.WithReplyChan[*rpc.AppendEntriesRequest, *rpc.AppendEntriesResponse]
	VoteReply          chan *rpc.RequestVoteResponse
	AppendEntriesReply chan *rpc.AppendEntriesResponse
	CommandCmd         chan utils.WithReplyChan[*rpc.CommandRequest, *rpc.CommandResponse]
	Err                chan error
}

func NewFlowControlChannels() FlowControlChannels {
	return FlowControlChannels{
		VoteCmd:            make(chan utils.WithReplyChan[*rpc.RequestVoteRequest, *rpc.RequestVoteResponse]),
		AppendEntriesCmd:   make(chan utils.WithReplyChan[*rpc.AppendEntriesRequest, *rpc.AppendEntriesResponse]),
		VoteReply:          make(chan *rpc.RequestVoteResponse),
		AppendEntriesReply: make(chan *rpc.AppendEntriesResponse),
		CommandCmd:         make(chan utils.WithReplyChan[*rpc.CommandRequest, *rpc.CommandResponse]),
		Err:                make(chan error),
	}
}
