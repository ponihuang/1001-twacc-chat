package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	if err := c.call(ctx, http.MethodGet, "getMe", nil, &bot); err != nil {
		return bot, err
	}
	return bot, nil
}

// SetWebhook configures Telegram to send updates to webhookURL.
func (c *Client) SetWebhook(ctx context.Context, webhookURL string) error {
	webhookURL = strings.TrimSpace(webhookURL)
	if webhookURL == "" {
		return fmt.Errorf("telegram webhook URL is required")
	}

	form := url.Values{}
	form.Set("url", webhookURL)
	return c.call(ctx, http.MethodPost, "setWebhook", form, nil)
}

// DeleteWebhook removes the currently configured Telegram webhook.
func (c *Client) DeleteWebhook(ctx context.Context) error {
	return c.call(ctx, http.MethodPost, "deleteWebhook", nil, nil)
}

// GetWebhookInfo returns Telegram's current webhook status.
func (c *Client) GetWebhookInfo(ctx context.Context) (WebhookInfo, error) {
	var info WebhookInfo
	if err := c.call(ctx, http.MethodGet, "getWebhookInfo", nil, &info); err != nil {
		return info, err
	}
	return info, nil
}

func (c *Client) call(ctx context.Context, method, apiMethod string, form url.Values, result any) error {
	if !c.Configured() {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is not configured")
	}

	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, c.methodURL(apiMethod), body)
	if err != nil {
		return fmt.Errorf("build Telegram %s request: %w", apiMethod, err)
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("call Telegram %s: %s", apiMethod, safeHTTPError(err))
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read Telegram %s response: %w", apiMethod, err)
	}

	var payload apiResponse[json.RawMessage]
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return fmt.Errorf("decode Telegram %s response: %w", apiMethod, err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices || !payload.OK {
		if payload.Description != "" {
			return fmt.Errorf("Telegram %s failed: status=%d error_code=%d description=%s", apiMethod, resp.StatusCode, payload.ErrorCode, payload.Description)
		}
		return fmt.Errorf("Telegram %s failed: status=%d", apiMethod, resp.StatusCode)
	}

	if result == nil {
		return nil
	}
	if len(payload.Result) == 0 || string(payload.Result) == "true" {
		return nil
	}
	if err := json.Unmarshal(payload.Result, result); err != nil {
		return fmt.Errorf("decode Telegram %s result: %w", apiMethod, err)
	}

	return nil
}

func (c *Client) httpClient() *http.Client {
	if c.http != nil {
		return c.http
	}
	return http.DefaultClient
}

func (c *Client) methodURL(method string) string {
	baseURL := strings.TrimRight(c.baseURL, "/")
	return fmt.Sprintf("%s/bot%s/%s", baseURL, c.botToken, method)
}

func safeHTTPError(err error) string {
	if urlErr, ok := err.(*url.Error); ok {
		return urlErr.Err.Error()
	}
	return err.Error()
}
