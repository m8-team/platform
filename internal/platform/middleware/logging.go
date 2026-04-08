package middleware

import (
	"context"
	"log/slog"
)

func Logger(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if logger == nil {
		return slog.Default()
	}
	if correlationID := FromGRPCContext(ctx).CorrelationID; correlationID != "" {
		return logger.With("correlation_id", correlationID)
	}
	return logger
}
