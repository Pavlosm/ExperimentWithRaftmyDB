package state

type CandidateVars struct {
	RequestVotes map[string]RequestVote
	VotesGranted map[string]VoteResponse
}
