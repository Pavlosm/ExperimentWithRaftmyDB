package state

type LeaderState struct {
	NextIndex  map[string]int
	MatchIndex map[string]int
}
