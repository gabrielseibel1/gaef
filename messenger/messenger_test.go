package messenger_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gabrielseibel1/gaef/messenger"
	"github.com/gabrielseibel1/gaef/types"
	amqp "github.com/rabbitmq/amqp091-go"
	"reflect"
	"testing"
)

type mockPublisher struct {
	ctx       context.Context
	exchange  string
	key       string
	mandatory bool
	immediate bool
	msg       amqp.Publishing
	err       error
}

func (m *mockPublisher) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	m.ctx = ctx
	m.exchange = exchange
	m.key = key
	m.mandatory = mandatory
	m.immediate = immediate
	m.msg = msg
	return m.err
}

var (
	dummyCtx             = context.TODO()
	dummyExchangeName    = "dummy-exchange-name"
	dummyID              = "dummy-id"
	dummyUser            = types.User{ID: dummyID, Name: "dummy-name"}
	dummyError           = errors.New("dummy-error")
	dummyBodyDeletion, _ = json.Marshal(dummyID)
	dummyBodyUpdate, _   = json.Marshal(dummyUser)
)

func TestMessenger_SendUserDeletedMessage(t *testing.T) {
	type fields struct {
		marshaller       messenger.Marshaller
		publisher        messenger.Publisher
		amqpExchangeName string
	}
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantErr       bool
		wantPublisher messenger.Publisher
	}{
		{
			name: "send user deleted message ok",
			fields: fields{
				marshaller:       json.Marshal,
				publisher:        &mockPublisher{},
				amqpExchangeName: dummyExchangeName,
			},
			args: args{
				ctx:    dummyCtx,
				userID: dummyID,
			},
			wantErr: false,
			wantPublisher: &mockPublisher{
				ctx:      dummyCtx,
				exchange: dummyExchangeName,
				msg: amqp.Publishing{
					Body: dummyBodyDeletion,
				},
			},
		},
		{
			name: "send user deleted message publisher error",
			fields: fields{
				marshaller:       json.Marshal,
				publisher:        &mockPublisher{err: dummyError},
				amqpExchangeName: dummyExchangeName,
			},
			args: args{
				ctx:    dummyCtx,
				userID: dummyID,
			},
			wantErr: true,
			wantPublisher: &mockPublisher{
				ctx:      dummyCtx,
				exchange: dummyExchangeName,
				msg: amqp.Publishing{
					Body: dummyBodyDeletion,
				},
				err: dummyError,
			},
		},
		{
			name: "send user deleted message marshal error",
			fields: fields{
				marshaller:       func(v any) ([]byte, error) { return nil, dummyError },
				publisher:        &mockPublisher{},
				amqpExchangeName: dummyExchangeName,
			},
			args: args{
				ctx:    dummyCtx,
				userID: dummyID,
			},
			wantErr:       true,
			wantPublisher: &mockPublisher{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := messenger.New(tt.fields.marshaller, "", tt.fields.amqpExchangeName, tt.fields.publisher)
			if err := m.SendUserDeletedMessage(tt.args.ctx, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("SendUserDeletedMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got, want := tt.fields.publisher, tt.wantPublisher; !reflect.DeepEqual(got, want) {
				t.Errorf("SendUserDeletedMessage() Publisher = %v, wantPublisher = %v", got, want)
			}
		})
	}
}

func TestMessenger_SendUserUpdatedMessage(t *testing.T) {
	type fields struct {
		marshaller       messenger.Marshaller
		publisher        messenger.Publisher
		amqpExchangeName string
	}
	type args struct {
		ctx  context.Context
		user types.User
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantErr       bool
		wantPublisher messenger.Publisher
	}{
		{
			name: "send user updated message ok",
			fields: fields{
				marshaller:       json.Marshal,
				publisher:        &mockPublisher{},
				amqpExchangeName: dummyExchangeName,
			},
			args: args{
				ctx:  dummyCtx,
				user: dummyUser,
			},
			wantErr: false,
			wantPublisher: &mockPublisher{
				ctx:      dummyCtx,
				exchange: dummyExchangeName,
				msg: amqp.Publishing{
					Body: dummyBodyUpdate,
				},
			},
		},
		{
			name: "send user updated message publisher error",
			fields: fields{
				marshaller:       json.Marshal,
				publisher:        &mockPublisher{err: dummyError},
				amqpExchangeName: dummyExchangeName,
			},
			args: args{
				ctx:  dummyCtx,
				user: dummyUser,
			},
			wantErr: true,
			wantPublisher: &mockPublisher{
				ctx:      dummyCtx,
				exchange: dummyExchangeName,
				msg: amqp.Publishing{
					Body: dummyBodyUpdate,
				},
				err: dummyError,
			},
		},
		{
			name: "send user updated message marshal error",
			fields: fields{
				marshaller:       func(v any) ([]byte, error) { return nil, dummyError },
				publisher:        &mockPublisher{},
				amqpExchangeName: dummyExchangeName,
			},
			args: args{
				ctx:  dummyCtx,
				user: dummyUser,
			},
			wantErr:       true,
			wantPublisher: &mockPublisher{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := messenger.New(tt.fields.marshaller, tt.fields.amqpExchangeName, "", tt.fields.publisher)
			if err := m.SendUserUpdatedMessage(tt.args.ctx, tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("SendUserUpdatedMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got, want := tt.fields.publisher, tt.wantPublisher; !reflect.DeepEqual(got, want) {
				t.Errorf("SendUserUpdatedMessage() Publisher = %v, wantPublisher = %v", got, want)
			}
		})
	}
}
