package logging

import (
	"go.uber.org/zap"
)

type Logger = zap.Logger

func New(development bool) (*zap.Logger, error) {
	if development {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
