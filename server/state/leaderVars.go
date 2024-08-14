package state

import "myDb/server/cfg"

type LeaderVars struct {
	NextIndex  map[cfg.NodeId]int64
	MatchIndex map[cfg.NodeId]int64
}

func NewDefaultLeaderVars(is []cfg.NodeId) LeaderVars {
	l := LeaderVars{
		NextIndex:  make(map[cfg.NodeId]int64),
		MatchIndex: make(map[cfg.NodeId]int64),
	}

	for _, i := range is {
		l.NextIndex[i] = 1
		l.MatchIndex[i] = 0
	}

	return l
}
