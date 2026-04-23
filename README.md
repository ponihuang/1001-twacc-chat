# 1001-twacc-chat

以 Go 為主體的 Web 聊天系統 MVP。產品主核心是聊天與群組通訊；外部系統接入僅作為統一 API 能力，讓其他網站或業務系統可以透過 API 與本對談系統溝通，不另為 ERP 或任何單一來源系統建立專屬產品分類或專屬界面。現階段已建立可執行的 HTTP server、首頁導覽頁、MySQL 連線初始化、migration 自動執行、外部系統註冊 / 登入 API、session/token 驗證、聊天列表 / 訊息列表 / 一對一對話 API、WebSocket 即時更新、圖片與檔案訊息上傳、本機靜態附件提供、IP 白名單管理 API，以及 trusted device 核准 API。

## 目前狀態

目前已完成：

- Go HTTP server 入口與基本路由
- `GET /`、`GET /chat`、`GET /m/chat`、`GET /embed/chat`、`GET /healthz`
- `GET /api/conversations`
- `POST /api/conversations/direct`
- `GET /api/conversations/{conversation_id}/messages`
- `POST /api/conversations/{conversation_id}/messages`
- `GET /ws`
- `GET /uploads/{filename}`
- 桌機版 Web 測試登入、對話列表、一對一對話建立
- 桌機版左欄下拉選單與子頁切換原型（登入 / 登入狀態 / 建立一對一對話 / 設定）
- 桌機版設定頁前端原型：頭像預覽、聊天室分類建立與分類列顯示
- 文字訊息送出、`Enter` 快捷送出
- 圖片與檔案訊息上傳
- MySQL 設定讀取、driver 初始化與 migration 自動執行
- 外部系統註冊 / 登入 API
- session token 簽發與 Bearer 驗證
- system_admin 裝置查詢與 IP 白名單管理 API
- trusted device 核准 API
- 使用者、session、裝置、IP 白名單、聊天核心資料表 migration
- 裝置與 IP 白名單的基本驗證邏輯
- 基本單元測試
- 三種 Web 形態骨架已實作：桌機全畫面、手機全畫面、可嵌入其他系統的小浮動視窗

目前尚未完成：

- 群組聊天建立流程
- 檔案留存策略
- `system_admin` 管理頁面
- 手機版 / 嵌入式頁面的聊天資料串接
- 三語系完整文案

## 目錄結構

```text
.
├─ db/
│  └─ migrations/
├─ docs/
├─ internal/
│  ├─ app/
│  ├─ auth/
│  ├─ chat/
│  ├─ config/
│  ├─ erp/          # 後續建議重命名為 integration/，作為外部系統接入入口
│  ├─ httpx/
│  └─ store/
├─ web/
│  ├─ static/
│  └─ templates/
├─ main.go
└─ TODO.md
```

## 啟動方式

```bash
go run .
```

程式啟動時會先嘗試讀取專案根目錄的 `.env`，再載入系統環境變數；若兩邊都有同名變數，系統環境變數優先。

預設啟動於 `http://localhost:8080`。

開發時若想要自動重啟 server，可使用 `air`：

```bash
air -c .air.toml
```

或：

```bash
make dev
```

若本機尚未安裝 `air`，可先安裝後再使用；專案已提供 [`.air.toml`](/Users/elva/新公司/聊天系統/1001-twacc-chat-main/.air.toml) 設定，會在 `.go`、`.html`、`.css`、`.js`、`.md`、`.toml` 變更時自動重建與重啟。

若要啟用 MySQL 與外部系統接入 / 管理 API：

```bash
ENABLE_MYSQL=true
AUTO_MIGRATE=true
MIGRATIONS_DIR=db/migrations
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_DATABASE=twacc_chat
MYSQL_USER=root
MYSQL_PASSWORD=secret
INTEGRATION_SHARED_TOKEN=your-shared-token
INTEGRATION_SIGNATURE_SECRET=your-signature-secret
INTEGRATION_TIMESTAMP_TOLERANCE=5m
```

建議把以上內容放進根目錄 `.env` 後直接執行 `go run .`。

若未設定 `INTEGRATION_SHARED_TOKEN` 與 `INTEGRATION_SIGNATURE_SECRET`，目前程式會以開發模式允許未帶整合驗證 header 的外部系統 register / login 請求。文件語意已改為統一外部整合能力；路由命名後續建議由 `/api/erp/*` 收斂為更中性的 `/api/integrations/*` 或 `/api/external/*`。

## 桌機測試頁

`/chat` 目前已接上桌機版 Web 測試流程：

- 直接輸入帳號登入
- 登入後載入對話列表
- 透過外部使用者編號建立 / 載入一對一對話
- 左上 icon 可打開左欄下拉選單，切換登入 / 登入狀態 / 建立一對一對話 / 設定子頁
- 設定頁目前提供頭像預覽與聊天室分類前端原型
- 聊天室分類目前保存在瀏覽器 `localStorage`，並顯示在左側 `Conversations` 上方
- `ALL` 為預設分類，固定顯示全部對話
- 傳送文字訊息
- `Enter` 送出、`Shift + Enter` 換行
- 上傳圖片或檔案訊息
- WebSocket 即時刷新對話列表與目前訊息區

目前限制：

- 設定頁與聊天室分類仍屬前端原型，尚未接正式設定 API
- 聊天室分類目前只完成前端保存與顯示，尚未真正套用到對話篩選邏輯

附件格式目前支援：

- `.jpg`
- `.jpeg`
- `.png`
- `.doc`
- `.pdf`
- `.xlsx`

單檔上限 `40MB`。

## 目前可用端點

- `GET /`
- `GET /chat`
- `GET /m/chat`
- `GET /embed/chat`
- `GET /healthz`
- `GET /ws`
- `GET /api/conversations`
- `POST /api/conversations/direct`
- `GET /api/conversations/{conversation_id}/messages`
- `POST /api/conversations/{conversation_id}/messages`
- `GET /uploads/{filename}`
- `POST /api/erp/register`
- `POST /api/erp/login`
- `GET /api/system-admin/users/{user_id}/devices`
- `PUT /api/system-admin/users/{user_id}/ip-whitelist`
- `POST /api/users/{user_id}/devices/{device_id}/approve`

## Web 介面形態

系統需自帶三種前端形態，且共用同一套聊天能力、登入狀態與權限規則：

- 桌機版全畫面：以對話列表 + 內容區為主的雙欄工作區
- 手機版全畫面：以單欄頁面切換為主，優先處理對話列表與訊息頁切換
- 嵌入式小浮動視窗：可插在其他系統頁面中的小型聊天入口與浮動對話窗

目前骨架路由：

- `/chat`：桌機版全畫面
- `/m/chat`：手機版全畫面
- `/embed/chat`：嵌入式浮動視窗

原則：

- 不為 ERP 或任何單一外部系統另開專屬聊天介面
- 外部系統只需透過 API 或嵌入式浮動視窗能力接入
- 三種前端形態應盡量共用同一套聊天元件與 API 契約

目前只有 `/chat` 已完成實際資料串接；`/m/chat` 與 `/embed/chat` 仍為骨架頁面。
`/chat` 另包含桌機版導覽 / 設定原型，但部分設定互動仍屬前端暫存實作。

## 認證方式

外部系統接入端點：

- `POST /api/erp/register`
- `POST /api/erp/login`

目前使用：

- `Authorization: Bearer <shared-token>`
- `X-TWACC-Timestamp: <unix-seconds>`
- `X-TWACC-Signature: <hex-hmac-sha256>`

簽章字串格式：

```text
HTTP_METHOD + "\n" + REQUEST_PATH + "\n" + TIMESTAMP + "\n" + RAW_BODY
```

預設 timestamp 容許誤差為 `5m`，可由 `INTEGRATION_TIMESTAMP_TOLERANCE` 調整。

完整對接範例請見 [docs/integration-signature.md](/Users/elva/新公司/聊天系統/1001-twacc-chat-main/docs/integration-signature.md)。

登入成功後，`/api/erp/login` 會回傳 session token。這個端點目前是外部系統接入的第一個相容入口；後續聊天 API、管理 API 與裝置核准 API 使用：

- `Authorization: Bearer <session-token>`

說明：

- `user` 可使用聊天 API 與後續聊天頁面
- `system_admin` 可使用管理 API，但不屬於聊天主流程參與者，聊天 API 應回 `403 SYSTEM_ADMIN_CANNOT_CHAT`

Web 測試頁目前是以「帳號輸入」方式包裝既有外部 API：

- 頁面仍然呼叫 `POST /api/erp/login`
- 若使用者不存在，前端會先呼叫 `POST /api/erp/register` 再重新登入
- 成功後把 session token 保存在瀏覽器，供後續聊天 API 與 WebSocket 使用

## 初始化資料庫

若 `AUTO_MIGRATE=true`，啟動時會自動讀取 `MIGRATIONS_DIR` 並套用尚未執行過的 `.sql` migration。

若要手動初始化：

1. 建立 MySQL database，例如 `twacc_chat`
2. 保留 `db/migrations/001_init.sql`、`db/migrations/002_auth_sessions.sql` 供檢視或手動套用
3. 設定環境變數後啟動服務

## 驗證

```bash
go build ./...
go test ./...
```
