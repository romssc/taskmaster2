package kafkaa

import (
	"context"
	"errors"
	"testing"

	"service1/internal/domain"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

type mockPublisher struct {
	c  mockClose
	s  mockStats
	wm mockWriteMessages
}

type mockClose struct {
	err error
}

type mockStats struct {
	stats kafka.WriterStats
}

type mockWriteMessages struct {
	err error
}

func (m *mockPublisher) Close() error {
	return m.c.err
}

func (m *mockPublisher) Stats() kafka.WriterStats {
	return m.s.stats
}

func (m *mockPublisher) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	return m.wm.err
}

type mockEncoder struct {
	b   []byte
	err error
}

func (m *mockEncoder) Marshal(data any) ([]byte, error) {
	return m.b, m.err
}

func Test_Close_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		producer *Producer
		err      error
	}{
		{
			name: "success",
			producer: &Producer{producer: &mockPublisher{
				c: mockClose{},
			}},
			err: nil,
		},
		{
			name: "failed to close",
			producer: &Producer{producer: &mockPublisher{
				c: mockClose{err: errors.New("")},
			}},
			err: ErrClosingConnection,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			err := cs.producer.Close()
			assert.ErrorIs(t, err, cs.err)
		})
	}
}

func Test_PublishEvent_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		ctx      context.Context
		event    domain.Event
		producer *Producer
		err      error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			event: domain.Event{
				Record: domain.Record{
					ID: 0,
				},
			},
			producer: &Producer{
				producer: &mockPublisher{
					wm: mockWriteMessages{}},
				encoder: &mockEncoder{}},
			err: nil,
		},
		{
			name: "failed to marshal event",
			ctx:  context.Background(),
			event: domain.Event{
				Record: domain.Record{
					ID: 0,
				},
			},
			producer: &Producer{
				producer: &mockPublisher{
					wm: mockWriteMessages{}},
				encoder: &mockEncoder{err: errors.New("")}},
			err: ErrMarshalingEvent,
		},
		{
			name: "group closed",
			ctx:  context.Background(),
			event: domain.Event{
				Record: domain.Record{
					ID: 0,
				},
			},
			producer: &Producer{
				producer: &mockPublisher{
					wm: mockWriteMessages{err: kafka.ErrGroupClosed}},
				encoder: &mockEncoder{}},
			err: ErrClosed,
		},
		{
			name: "failed to produce event",
			ctx:  context.Background(),
			event: domain.Event{
				Record: domain.Record{
					ID: 0,
				},
			},
			producer: &Producer{
				producer: &mockPublisher{
					wm: mockWriteMessages{err: errors.New("")}},
				encoder: &mockEncoder{}},
			err: ErrProducingEvent,
		},
		{
			name: "context closed mid kafka call",
			ctx:  context.Background(),
			event: domain.Event{
				Record: domain.Record{
					ID: 0,
				},
			},
			producer: &Producer{
				producer: &mockPublisher{
					wm: mockWriteMessages{err: context.Canceled}},
				encoder: &mockEncoder{}},
			err: ErrOperationCanceled,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			err := cs.producer.PublishEvent(cs.ctx, cs.event)
			assert.ErrorIs(t, err, cs.err)
		})
	}
}
