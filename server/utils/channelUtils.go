package utils

type WithReplyChan[T any, V any] struct {
	Data  T
	Reply chan V
}
