package agent

import (
	"context"

	"github.com/ckken/ralphx/internal/contracts"
)

type Request struct {
	Workdir          string
	Prompt           string
	OutputSchemaPath string
	RawLogPath       string
	ExtraArgs        []string
	SessionID        string
}

type Response struct {
	RawOutput []byte
	Parsed    contracts.RoundResult
	SessionID string
}

type Agent interface {
	Run(ctx context.Context, req Request) (Response, error)
}
