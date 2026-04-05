package logging

import (
	"github.com/m8platform/platform/iam/internal/observability"
	"go.uber.org/zap"
)

type Logger = zap.Logger

func New(development bool) (*zap.Logger, error) {
	return observability.NewLogger(development)
}
