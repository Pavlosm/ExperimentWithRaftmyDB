package state

type LogVars struct {
	Log            []Log
	CommittedIndex int64
}

func NewDefaultLogVars() *LogVars {
	return &LogVars{
		Log:            make([]Log, 1000),
		CommittedIndex: 0,
	}
}

func (l *LogVars) GetLastLogProps() (index int64, term int64) {
	if len(l.Log) == 0 {
		return 0, 0
	}

	ll := l.Log[len(l.Log)-1]
	return ll.Index, ll.Term
}
