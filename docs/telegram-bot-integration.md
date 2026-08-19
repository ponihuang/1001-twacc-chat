# Telegram Bot 串接測試流程

本文件記錄 Telegram Bot 串接第一階段的本機測試流程。

目前範圍：

- Go server 提供 `POST /api/telegram/webhook`
- `TELEGRAM_BOT_TOKEN` 由環境變數或 `.env` 讀取
- webhook 接收 Telegram Update
- 收到文字訊息時先輸出 log
- 暫不寫入資料庫
- 暫不接入既有聊天室訊息流程

## 1. 環境變數

在本機 `.env` 或部署環境設定：

```bash
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
```

不要將實際 token commit 到程式碼或文件。

Cloudflare Tunnel 測試時可另外設定：

```bash
TELEGRAM_WEBHOOK_URL=https://your-tunnel.trycloudflare.com/api/telegram/webhook
```

## 2. 啟動 Go server

```bash
go run .
```

確認服務正常：

```bash
curl -i http://127.0.0.1:8080/healthz
```

預期回應：

```http
HTTP/1.1 200 OK
```

```json
{"status":"ok"}
```

## 3. 建立 Cloudflare Tunnel

本機測試 Telegram webhook 需要公開 HTTPS URL。

```bash
cloudflared tunnel --url http://127.0.0.1:8080
```

啟動後會看到類似：

```text
https://your-tunnel.trycloudflare.com
```

測試 tunnel 是否能連到本機：

```bash
curl -i https://your-tunnel.trycloudflare.com/healthz
```

預期回應為 `200 OK`。

## 4. 設定 Telegram Webhook

將 webhook URL 設為：

```text
https://your-tunnel.trycloudflare.com/api/telegram/webhook
```

使用 curl 設定：

```bash
TOKEN=$(awk -F= '$1=="TELEGRAM_BOT_TOKEN" {print substr($0, index($0, "=")+1); exit}' .env)
TELEGRAM_WEBHOOK_URL="https://your-tunnel.trycloudflare.com/api/telegram/webhook"

curl -sS -X POST "https://api.telegram.org/bot${TOKEN}/setWebhook" \
  --data-urlencode "url=${TELEGRAM_WEBHOOK_URL}"
```

成功時 Telegram 會回：

```json
{"ok":true,"result":true,"description":"Webhook was set"}
```

## 5. 查詢 Webhook 狀態

```bash
TOKEN=$(awk -F= '$1=="TELEGRAM_BOT_TOKEN" {print substr($0, index($0, "=")+1); exit}' .env)

curl -sS "https://api.telegram.org/bot${TOKEN}/getWebhookInfo"
```

重點檢查：

- `url` 是否為目前 tunnel 的 webhook URL
- `pending_update_count` 是否持續累積
- `last_error_message` 是否有錯誤

## 6. 實際傳訊息測試

1. 在 Telegram 找到測試 bot
2. 傳送 `/start`
3. 傳送 `你好`
4. 回到 `go run .` 的 terminal 檢查 log

預期會看到類似：

```text
telegram text message received chat_id=... user_id=... username="..." first_name="..." message_id=... text="你好"
```

## 7. 清除 Webhook

若本機 tunnel 已關閉，建議清除 webhook，避免 Telegram 持續送到失效 URL。

```bash
TOKEN=$(awk -F= '$1=="TELEGRAM_BOT_TOKEN" {print substr($0, index($0, "=")+1); exit}' .env)

curl -sS -X POST "https://api.telegram.org/bot${TOKEN}/deleteWebhook"
```

成功時 Telegram 會回：

```json
{"ok":true,"result":true}
```

## 8. 驗收狀態

目前已驗證：

- Bot Token 可用
- Cloudflare Tunnel 可轉發到 `localhost:8080`
- Telegram webhook 可設定成功
- Telegram 實際文字訊息可進到 Go server
- Go terminal 可看到文字訊息 log

目前尚未實作：

- Telegram 訊息寫入資料庫
- Telegram user 與系統 user 對應
- Telegram 訊息接入既有聊天室流程
