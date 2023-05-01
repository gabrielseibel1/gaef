package messenger

import (
	"context"
	"github.com/gabrielseibel1/gaef/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Messenger struct {
	marshal                Marshaller
	publisher              Publisher
	amqpUserUpdateExchange string
	amqpUserDeleteExchange string
}

func New(marshaller Marshaller, amqpUserUpdateExchange, amqpUserDeleteExchange string, publisher Publisher) Messenger {
	return Messenger{
		marshal:                marshaller,
		publisher:              publisher,
		amqpUserUpdateExchange: amqpUserUpdateExchange,
		amqpUserDeleteExchange: amqpUserDeleteExchange,
	}
}

type Publisher interface {
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

type Marshaller func(v any) ([]byte, error)

func (m Messenger) sendMessage(ctx context.Context, exchange string, msg any) error {
	body, err := m.marshal(msg)
	if err != nil {
		return err
	}

	err = m.publisher.PublishWithContext(
		ctx,
		exchange,
		"",
		false,
		false,
		amqp.Publishing{
			Body: body,
		},
	)
	return err
}

func (m Messenger) SendUserDeletedMessage(ctx context.Context, userID string) error {
	return m.sendMessage(ctx, m.amqpUserDeleteExchange, userID)
}

func (m Messenger) SendUserUpdatedMessage(ctx context.Context, user types.User) error {
	return m.sendMessage(ctx, m.amqpUserUpdateExchange, user)
}
