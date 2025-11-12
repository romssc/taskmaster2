package kafkaa

import (
	"context"
	"errors"
	"taskmaster2/service2/internal/controller/kafkarouter"
	"taskmaster2/service2/internal/domain"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

type mockReader struct {
	c   mockClose
	cm  mockCommitMessages
	cf  mockConfig
	fm  mockFetchMessage
	l   mockLag
	o   mockOffset
	rl  mockReadLag
	rm  mockReadMessage
	so  mockSetOffset
	soa mockSetOffsetAt
	s   mockStats
}

type mockClose struct {
	err error
}

type mockCommitMessages struct {
	err error
}

type mockConfig struct {
	config kafka.ReaderConfig
}

type mockFetchMessage struct {
	message kafka.Message
	err     error
}

type mockLag struct {
	lag int64
}

type mockOffset struct {
	offset int64
}

type mockReadLag struct {
	lag int64
	err error
}

type mockReadMessage struct {
	message kafka.Message
	err     error
}

type mockSetOffset struct {
	err error
}

type mockSetOffsetAt struct {
	err error
}

type mockStats struct {
	stats kafka.ReaderStats
}

func (m *mockReader) Close() error {
	return m.c.err
}

func (m *mockReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return m.cm.err
}

func (m *mockReader) Config() kafka.ReaderConfig {
	return m.cf.config
}

func (m *mockReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return m.fm.message, m.fm.err
}

func (m *mockReader) Lag() int64 {
	return m.l.lag
}

func (m *mockReader) Offset() int64 {
	return m.o.offset
}

func (m *mockReader) ReadLag(ctx context.Context) (lag int64, err error) {
	return m.rl.lag, m.rl.err
}

func (m *mockReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return m.rm.message, m.rm.err
}

func (m *mockReader) SetOffset(offset int64) error {
	return m.so.err
}

func (m *mockReader) SetOffsetAt(ctx context.Context, t time.Time) error {
	return m.soa.err
}

func (m *mockReader) Stats() kafka.ReaderStats {
	return m.s.stats
}

type mockHandler struct {
	err error
}

func (m *mockHandler) Pick(ctx context.Context, action domain.Action, event domain.Event) error {
	return m.err
}

type mockEncoder struct {
	b   []byte
	err error
}

func (m *mockEncoder) Marshal(data any) ([]byte, error) {
	return m.b, m.err
}

type mockDecoder struct {
	err error
}

func (m *mockDecoder) Unmarshal(data []byte, v any) error {
	return m.err
}

type mockProccessor struct {
	err error
}

func (m *mockProccessor) Process(ctx context.Context) error {
	return m.err
}

func Test_Run_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		ctx      context.Context
		consumer *Consumer
		err      error
	}{
		{
			name: "context canceled mid proccessor call",
			ctx:  context.Background(),
			consumer: &Consumer{
				proccessor: &mockProccessor{err: ErrConsumerClosed},
			},
			err: ErrConsumerClosed,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			err := cs.consumer.Run(cs.ctx)
			assert.ErrorIs(t, err, cs.err)
		})
	}
}

func Test_Proccess_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		ctx        context.Context
		proccessor *Proccess
		err        error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			proccessor: &Proccess{
				reader: &mockReader{
					fm: mockFetchMessage{},
					cm: mockCommitMessages{},
				},
				handler: &mockHandler{},
				decoder: &mockDecoder{},
			},
			err: nil,
		},
		{
			name: "failed to fetch messages",
			ctx:  context.Background(),
			proccessor: &Proccess{
				reader: &mockReader{
					fm: mockFetchMessage{err: ErrFetchingMessages},
					cm: mockCommitMessages{},
				},
				handler: &mockHandler{},
				decoder: &mockDecoder{},
			},
			err: ErrFetchingMessages,
		},
		{
			name: "consumer closed mid fetching messages",
			ctx:  context.Background(),
			proccessor: &Proccess{
				reader: &mockReader{
					fm: mockFetchMessage{err: kafka.ErrGroupClosed},
					cm: mockCommitMessages{},
				},
				handler: &mockHandler{},
				decoder: &mockDecoder{},
			},
			err: ErrConsumerClosed,
		},
		{
			name: "context closed mid fetching messages",
			ctx:  context.Background(),
			proccessor: &Proccess{
				reader: &mockReader{
					fm: mockFetchMessage{err: context.Canceled},
					cm: mockCommitMessages{},
				},
				handler: &mockHandler{},
				decoder: &mockDecoder{},
			},
			err: ErrOperationCanceled,
		},
		{
			name: "failed to unmarshal message",
			ctx:  context.Background(),
			proccessor: &Proccess{
				reader: &mockReader{
					fm: mockFetchMessage{},
					cm: mockCommitMessages{},
				},
				handler: &mockHandler{},
				decoder: &mockDecoder{err: errors.New("")},
			},
			err: ErrUnmarshalingMessage,
		},
		{
			name: "failed to proccess event",
			ctx:  context.Background(),
			proccessor: &Proccess{
				reader: &mockReader{
					fm: mockFetchMessage{},
					cm: mockCommitMessages{},
				},
				handler: &mockHandler{err: kafkarouter.ErrFailedToProccess},
				decoder: &mockDecoder{},
			},
			err: ErrProcessingMessage,
		},
		{
			name: "context closed mid proccessing",
			ctx:  context.Background(),
			proccessor: &Proccess{
				reader: &mockReader{
					fm: mockFetchMessage{},
					cm: mockCommitMessages{},
				},
				handler: &mockHandler{err: kafkarouter.ErrOperationCanceled},
				decoder: &mockDecoder{},
			},
			err: ErrOperationCanceled,
		},
		{
			name: "failed to commit",
			ctx:  context.Background(),
			proccessor: &Proccess{
				reader: &mockReader{
					fm: mockFetchMessage{},
					cm: mockCommitMessages{err: errors.New("")},
				},
				handler: &mockHandler{},
				decoder: &mockDecoder{},
			},
			err: ErrCommitting,
		},
		{
			name: "consumer closed mid commiting",
			ctx:  context.Background(),
			proccessor: &Proccess{
				reader: &mockReader{
					fm: mockFetchMessage{},
					cm: mockCommitMessages{err: kafka.ErrGroupClosed},
				},
				handler: &mockHandler{},
				decoder: &mockDecoder{},
			},
			err: ErrConsumerClosed,
		},
		{
			name: "context closed mid commiting",
			ctx:  context.Background(),
			proccessor: &Proccess{
				reader: &mockReader{
					fm: mockFetchMessage{},
					cm: mockCommitMessages{err: context.Canceled},
				},
				handler: &mockHandler{},
				decoder: &mockDecoder{},
			},
			err: ErrOperationCanceled,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			err := cs.proccessor.Process(cs.ctx)
			assert.ErrorIs(t, err, cs.err)
		})
	}
}
