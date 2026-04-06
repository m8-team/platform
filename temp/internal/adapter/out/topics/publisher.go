package topics

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Publisher struct {
	logger *zap.Logger
}

func NewPublisher(logger *zap.Logger) *Publisher {
	return &Publisher{logger: logger}
}

func (p *Publisher) PublishProto(_ context.Context, topic string, message proto.Message) error {
	payload, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(message)
	if err != nil {
		return err
	}
	p.logger.Info("publish domain event", zap.String("topic", topic), zap.ByteString("payload", payload))
	return nil
}
