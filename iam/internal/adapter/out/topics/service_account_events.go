package topics

import (
	"context"

	eventsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/events/v1"
	identityv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/identity/v1"
	identitymodel "github.com/m8platform/platform/iam/internal/module/iam/model"
	legacytopics "github.com/m8platform/platform/iam/internal/topics"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ServiceAccountEventPublisher struct {
	publisher *legacytopics.Publisher
	topic     string
}

func NewServiceAccountEventPublisher(publisher *legacytopics.Publisher, topic string) *ServiceAccountEventPublisher {
	return &ServiceAccountEventPublisher{
		publisher: publisher,
		topic:     topic,
	}
}

func (p *ServiceAccountEventPublisher) PublishServiceAccountCreated(ctx context.Context, event identitymodel.ServiceAccountCreatedEvent) error {
	if p == nil || p.publisher == nil {
		return nil
	}

	return p.publisher.PublishProto(ctx, p.topic, &eventsv1.ServiceAccountCreated{
		Meta: &eventsv1.EventMeta{
			EventId:       event.OperationID,
			OccurredAt:    timestamppb.New(event.OccurredAt.UTC()),
			CorrelationId: event.OperationID,
			TenantId:      event.Account.TenantID,
		},
		ServiceAccount: &identityv1.ServiceAccount{
			ServiceAccountId: event.Account.ID,
			TenantId:         event.Account.TenantID,
			DisplayName:      event.Account.DisplayName,
			Description:      event.Account.Description,
			Disabled:         event.Account.Disabled,
			KeycloakClientId: event.Account.KeycloakClientID,
			OperationId:      event.Account.OperationID,
			CreatedAt:        timestamppb.New(event.Account.CreatedAt.UTC()),
			UpdatedAt:        timestamppb.New(event.Account.UpdatedAt.UTC()),
		},
	})
}
