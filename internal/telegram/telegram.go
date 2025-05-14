package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var botToken string
var apiURL string

func Init() {
	botToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		panic("TELEGRAM_BOT_TOKEN environment variable is not set")
	}
	apiURL = fmt.Sprintf("https://api.telegram.org/bot%s", botToken)
}

func SendMessage(chatID string, message string) error {
	url := fmt.Sprintf("%s/sendMessage", apiURL)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-OK status: %s", resp.Status)
	}

	return nil
}
