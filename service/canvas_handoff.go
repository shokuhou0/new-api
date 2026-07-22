package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/go-redis/redis/v8"
)

const canvasHandoffRedisPrefix = "canvas_handoff:"

var ErrCanvasHandoffTicketInvalid = errors.New("canvas handoff ticket is invalid or expired")

type CanvasHandoffTicketPayload struct {
	UserID    int   `json:"user_id"`
	TokenID   int   `json:"token_id"`
	ExpiresAt int64 `json:"expires_at"`
}

type canvasHandoffMemoryEntry struct {
	Payload   string
	ExpiresAt time.Time
}

var canvasHandoffMemory = struct {
	sync.Mutex
	Tickets map[string]canvasHandoffMemoryEntry
}{Tickets: make(map[string]canvasHandoffMemoryEntry)}

const consumeCanvasHandoffTicketScript = `
local value = redis.call("GET", KEYS[1])
if not value then
  return nil
end
redis.call("DEL", KEYS[1])
return value
`

func CreateCanvasHandoffTicket(userID, tokenID int, ttl time.Duration) (string, int64, error) {
	if userID <= 0 || tokenID <= 0 || ttl <= 0 {
		return "", 0, fmt.Errorf("invalid canvas handoff ticket parameters")
	}

	ticket, err := common.GenerateRandomCharsKey(48)
	if err != nil {
		return "", 0, err
	}
	expiresAt := time.Now().Add(ttl)
	payload := CanvasHandoffTicketPayload{
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt.Unix(),
	}
	encodedPayload, err := common.Marshal(payload)
	if err != nil {
		return "", 0, err
	}

	key := canvasHandoffCacheKey(ticket)
	if common.RedisEnabled && common.RDB != nil {
		if err := common.RedisSet(key, string(encodedPayload), ttl); err != nil {
			return "", 0, err
		}
	} else {
		canvasHandoffMemory.Lock()
		for cachedKey, entry := range canvasHandoffMemory.Tickets {
			if !entry.ExpiresAt.After(time.Now()) {
				delete(canvasHandoffMemory.Tickets, cachedKey)
			}
		}
		canvasHandoffMemory.Tickets[key] = canvasHandoffMemoryEntry{
			Payload:   string(encodedPayload),
			ExpiresAt: expiresAt,
		}
		canvasHandoffMemory.Unlock()
	}

	return ticket, expiresAt.Unix(), nil
}

func ConsumeCanvasHandoffTicket(ticket string) (CanvasHandoffTicketPayload, error) {
	var payload CanvasHandoffTicketPayload
	ticket = strings.TrimSpace(ticket)
	if ticket == "" {
		return payload, ErrCanvasHandoffTicketInvalid
	}

	key := canvasHandoffCacheKey(ticket)
	var encodedPayload string
	if common.RedisEnabled && common.RDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		result, err := common.RDB.Eval(ctx, consumeCanvasHandoffTicketScript, []string{key}).Text()
		if errors.Is(err, redis.Nil) {
			return payload, ErrCanvasHandoffTicketInvalid
		}
		if err != nil {
			return payload, err
		}
		encodedPayload = result
	} else {
		canvasHandoffMemory.Lock()
		entry, ok := canvasHandoffMemory.Tickets[key]
		delete(canvasHandoffMemory.Tickets, key)
		canvasHandoffMemory.Unlock()
		if !ok || !entry.ExpiresAt.After(time.Now()) {
			return payload, ErrCanvasHandoffTicketInvalid
		}
		encodedPayload = entry.Payload
	}

	if err := common.Unmarshal([]byte(encodedPayload), &payload); err != nil {
		return payload, err
	}
	if payload.UserID <= 0 || payload.TokenID <= 0 || payload.ExpiresAt < time.Now().Unix() {
		return CanvasHandoffTicketPayload{}, ErrCanvasHandoffTicketInvalid
	}
	return payload, nil
}

func canvasHandoffCacheKey(ticket string) string {
	digest := sha256.Sum256([]byte(ticket))
	return canvasHandoffRedisPrefix + hex.EncodeToString(digest[:])
}
