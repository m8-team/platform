package topic

import (
	legacytopic "github.com/m8platform/platform/iam/internal/adapter/in/topic"
	tenantuc "github.com/m8platform/platform/iam/internal/usecase/tenant"
)

type SupportAccessConsumer struct {
	*legacytopic.SupportAccessConsumer
}

func NewSupportAccessConsumer(useCase *tenantuc.SupportAccessUseCase) *SupportAccessConsumer {
	return &SupportAccessConsumer{
		SupportAccessConsumer: legacytopic.NewSupportAccessConsumer(useCase),
	}
}
