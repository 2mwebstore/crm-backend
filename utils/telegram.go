package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// telegramHTTPClient has a short, explicit timeout — a notification
// failing/hanging must never block or slow down the actual Deposit/
// Withdrawal request that triggered it.
var telegramHTTPClient = &http.Client{Timeout: 5 * time.Second}

type telegramSendMessageRequest struct {
	ChatID          string `json:"chat_id"`
	Text            string `json:"text"`
	ParseMode       string `json:"parse_mode,omitempty"`
	MessageThreadID *int   `json:"message_thread_id,omitempty"`
}

// SendTelegramMessage posts text to a Telegram chat (optionally a specific
// forum topic within it) via the Bot API's sendMessage endpoint. Intended
// to be called via `go utils.SendTelegramMessage(...)` (fire-and-forget) —
// it never returns an error to the caller; failures are only logged, since
// a notification going missing should never fail or roll back the
// Deposit/Withdrawal that triggered it.
func SendTelegramMessage(botToken, chatID string, topicID *int, text string) {
	if botToken == "" || chatID == "" {
		return // not configured for this branch — nothing to do
	}

	body, err := json.Marshal(telegramSendMessageRequest{
		ChatID:          chatID,
		Text:            text,
		ParseMode:       "HTML",
		MessageThreadID: topicID,
	})
	if err != nil {
		log.Printf("[telegram] failed to encode message: %v", err)
		return
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, err := telegramHTTPClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[telegram] failed to send message: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[telegram] sendMessage returned status %d for chat %s", resp.StatusCode, chatID)
	}
}
