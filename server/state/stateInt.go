package state

import "myDb/server/cfg"

type Stater interface {
	GetCurrentTerm() int64
	HandleRequestVoteResponse(n cfg.NodeId, t int64, granted bool)
	HandleVoteRequest(cId string, t int64, lt int64, li int64) bool
}
