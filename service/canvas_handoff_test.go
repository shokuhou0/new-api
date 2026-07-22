package service

import (
	"errors"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
)

func TestCanvasHandoffTicketCanOnlyBeConsumedOnce(t *testing.T) {
	previousRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	canvasHandoffMemory.Lock()
	canvasHandoffMemory.Tickets = make(map[string]canvasHandoffMemoryEntry)
	canvasHandoffMemory.Unlock()
	t.Cleanup(func() {
		common.RedisEnabled = previousRedisEnabled
		canvasHandoffMemory.Lock()
		canvasHandoffMemory.Tickets = make(map[string]canvasHandoffMemoryEntry)
		canvasHandoffMemory.Unlock()
	})

	ticket, expiresAt, err := CreateCanvasHandoffTicket(17, 23, time.Minute)
	if err != nil {
		t.Fatalf("CreateCanvasHandoffTicket() error = %v", err)
	}
	if ticket == "" || expiresAt <= time.Now().Unix() {
		t.Fatalf("CreateCanvasHandoffTicket() returned invalid ticket metadata")
	}

	payload, err := ConsumeCanvasHandoffTicket(ticket)
	if err != nil {
		t.Fatalf("ConsumeCanvasHandoffTicket() error = %v", err)
	}
	if payload.UserID != 17 || payload.TokenID != 23 {
		t.Fatalf("unexpected payload: %+v", payload)
	}

	if _, err := ConsumeCanvasHandoffTicket(ticket); !errors.Is(err, ErrCanvasHandoffTicketInvalid) {
		t.Fatalf("second consume error = %v, want ErrCanvasHandoffTicketInvalid", err)
	}
}

func TestCanvasHandoffTicketRejectsExpiredMemoryEntry(t *testing.T) {
	previousRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	t.Cleanup(func() { common.RedisEnabled = previousRedisEnabled })

	ticket := "expired-ticket"
	canvasHandoffMemory.Lock()
	canvasHandoffMemory.Tickets[canvasHandoffCacheKey(ticket)] = canvasHandoffMemoryEntry{
		Payload:   `{"user_id":1,"token_id":2,"expires_at":1}`,
		ExpiresAt: time.Now().Add(-time.Second),
	}
	canvasHandoffMemory.Unlock()

	if _, err := ConsumeCanvasHandoffTicket(ticket); !errors.Is(err, ErrCanvasHandoffTicketInvalid) {
		t.Fatalf("ConsumeCanvasHandoffTicket() error = %v, want ErrCanvasHandoffTicketInvalid", err)
	}
}
