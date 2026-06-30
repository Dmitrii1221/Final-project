package kafka

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *kafka.Reader
	handler *Handler
}

func New(brokers, topic, groupID, dlqTopic string, handler *Handler) *Consumer {
	handler.dlqWriter = &kafka.Writer{
		Addr:     kafka.TCP(brokers),
		Topic:    dlqTopic,
		Balancer: &kafka.Hash{},
	}

	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{brokers},
			Topic:     topic,
			GroupID:   groupID,
			MinBytes:  10,
			MaxBytes:  10e6,
		}),
		handler: handler,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("starting kafka consumer")

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		if err := c.handler.ProcessMessage(ctx, msg); err != nil {
			slog.Error("processing message failed, sending to DLQ", "err", err)

			if dlqErr := c.handler.sendToDLQ(ctx, msg, err.Error()); dlqErr != nil {
				slog.Error("failed to send to DLQ", "err", dlqErr)
			}
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("commit message failed", "err", err)
		}
	}
}
