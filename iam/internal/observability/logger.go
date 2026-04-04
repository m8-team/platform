package observability

import "go.uber.org/zap"

func NewLogger(development bool) (*zap.Logger, error) {
	if development {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
