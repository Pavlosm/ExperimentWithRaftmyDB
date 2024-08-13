package state

type StateMachine struct {
	VolatileState VolatileState
	LeaderState   LeaderState
	ServerVars    ServerVars
	LogState      LogState
}
