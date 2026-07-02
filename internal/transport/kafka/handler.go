package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"final-project/internal/service"

	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
)

type SpendEvent struct {
	IdempotencyKey string `json:"idempotency_key"`
	BudgetID       int64  `json:"budget_id"`
	CurrencyCode   string `json:"currency_code"`
	Amount         string `json:"amount"`
	SpentAt        string `json:"spent_at"`
	UserID         int64  `json:"user_id"`
	Source         string `json:"source"`
}

type Handler struct {
	service  *service.SpendingService
	dlqWriter *kafka.Writer
}

func NewHandler(svc *service.SpendingService) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	var event SpendEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	if event.BudgetID == 0 || event.CurrencyCode == "" || event.Amount == "" || event.SpentAt == "" || event.UserID == 0 {
		return fmt.Errorf("missing required fields")
	}

	spentAt, err := time.Parse(time.RFC3339, event.SpentAt)
	if err != nil {
		return fmt.Errorf("invalid spent_at: %w", err)
	}

	amount, err := decimal.NewFromString(event.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	_, err = h.service.Process(ctx, service.SpendCommand{
		BudgetID:       event.BudgetID,
		CurrencyCode:   event.CurrencyCode,
		Amount:         amount,
		SpentAt:        spentAt,
		UserID:         event.UserID,
		IdempotencyKey: event.IdempotencyKey,
		Source:         event.Source,
	})
	if err != nil {
		return err
	}

	slog.Info("spending processed", "budget", event.BudgetID)
	return nil
}

func (h *Handler) sendToDLQ(ctx context.Context, original kafka.Message, reason string) error {
	dlqMsg := struct {
		Original  json.RawMessage `json:"original"`
		Reason    string          `json:"error_reason"`
		Error     string          `json:"error_message"`
		FailedAt  string          `json:"failed_at"`
		Instance  string          `json:"consumer_instance"`
	}{
		Original: original.Value,
		Reason:   reason,
		Error:    reason,
		FailedAt: time.Now().UTC().Format(time.RFC3339),
		Instance: "budget-consumer-1",
	}

	body, err := json.Marshal(dlqMsg)
	if err != nil {
		return fmt.Errorf("marshal dlq: %w", err)
	}

	return h.dlqWriter.WriteMessages(ctx, kafka.Message{
		Key:   original.Key,
		Value: body,
	})
}
