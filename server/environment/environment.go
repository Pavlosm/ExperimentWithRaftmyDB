package environment

import (
	"log/slog"
	"os"
)

type Env struct {
	Logger   *slog.Logger
	Channels FlowControlChannels
}

func NewEnvironment() *Env {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	slog.SetDefault(logger)
	return &Env{
		Logger:   logger,
		Channels: NewFlowControlChannels(),
	}
}
