package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultAPIBaseURL = "https://api.telegram.org"

// Client holds Telegram Bot API configuration.
type Client struct {
	botToken string
	baseURL  string
	http     *http.Client
}

type apiResponse[T any] struct {
	OK          bool   `json:"ok"`
	Result      T      `json:"result"`
	Description string `json:"description"`
	ErrorCode   int    `json:"error_code"`
}

// NewClient builds a Telegram Bot API client from the configured bot token.
func NewClient(botToken string) *Client {
	return &Client{
		botToken: strings.TrimSpace(botToken),
		baseURL:  defaultAPIBaseURL,
		http:     http.DefaultClient,
	}
}

// Configured returns true when the bot token is available.
func (c *Client) Configured() bool {
	return c != nil && c.botToken != ""
}

// GetMe calls Telegram Bot API getMe and returns the bot identity.
func (c *Client) GetMe(ctx context.Context) (BotUser, error) {
	var bot BotUser
	if !c.Configured() {
		return bot, fmt.Errorf("TELEGRAM_BOT_TOKEN is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.methodURL("getMe"), nil)
	if err != nil {
		return bot, fmt.Errorf("build Telegram getMe request: %w", err)
	}

	httpClient := c.http
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return bot, fmt.Errorf("call Telegram getMe: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return bot, fmt.Errorf("read Telegram getMe response: %w", err)
	}

	var payload apiResponse[BotUser]
	if err := json.Unmarshal(body, &payload); err != nil {
		return bot, fmt.Errorf("decode Telegram getMe response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices || !payload.OK {
		if payload.Description != "" {
			return bot, fmt.Errorf("Telegram getMe failed: status=%d error_code=%d description=%s", resp.StatusCode, payload.ErrorCode, payload.Description)
		}
		return bot, fmt.Errorf("Telegram getMe failed: status=%d", resp.StatusCode)
	}

	return payload.Result, nil
}

func (c *Client) methodURL(method string) string {
	baseURL := strings.TrimRight(c.baseURL, "/")
	return fmt.Sprintf("%s/bot%s/%s", baseURL, c.botToken, method)
}
