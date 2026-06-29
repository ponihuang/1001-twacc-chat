# 1001-twacc-chat

以 Go 為主體的 Web 聊天系統 MVP，同一個服務內包含聊天室前台、聊天系統管理後台與共用 API。產品主核心是聊天與群組通訊；外部系統接入僅作為統一 API 能力，不另為 ERP 或任何單一來源系統建立專屬產品分類或專屬界面。

分支責任：

- `main`：正式穩定版本
- `develop`：前台與後台功能整合分支
- `develop-nina`：聊天室前台開發分支
- `develop-admin`：聊天系統後台開發分支

## 目前狀態

目前已完成：

- Go HTTP server 入口與基本路由
- `GET /`、`GET /home`、`GET /login`、`GET /chat`、`GET /m/chat`、`GET /embed/chat`、`GET /healthz`
- `GET /api/conversations`
- `POST /api/conversations/direct`
- `POST /api/conversations/group`
- `GET /api/contacts`
- `POST /api/contacts`
- `GET /api/users/search`
- `GET /api/messages/search`
- `PATCH /api/conversations/{conversation_id}`
- `GET /api/conversations/{conversation_id}/members`
- `POST /api/conversations/{conversation_id}/members`
- `DELETE /api/conversations/{conversation_id}/members`
- `GET /api/conversations/{conversation_id}/messages`
- `POST /api/conversations/{conversation_id}/messages`
- `POST /api/conversations/{conversation_id}/read`
- `POST /api/conversations/{conversation_id}/messages/{message_id}/recall`
- `DELETE /api/conversations/{conversation_id}/messages/{message_id}`
- `GET /ws`
- `GET /uploads/{filename}`
- `/` 依登入狀態導向 `/login` 或 `/chat`
- `/home` 與 `/chat` 未登入或 session 過期會導向 `/login`
- `/home` 保留原入口導覽畫面
- `/login` 為獨立登入頁，登入後導向 `/chat`，登出後導向 `/login`
- 桌機版 Web 測試登入、對話列表、搜尋使用者與訊息
- 桌機版搜尋使用者時，點選使用者會先比對既有一對一對話；若已有對話會直接載入，若尚未建立則先開啟暫存聊天畫面，送出第一則訊息時才建立一對一對話
- 桌機版左側新增入口可建立多人群組：新增成員、輸入群組名稱後建立群組對話
- 聯絡人清單、加入聯絡人、群組資訊、群組成員新增與移除
- 群組名稱與描述更新
- 桌機版左欄下拉選單：第一列顯示登入帳號並進入設定頁
- 桌機版設定頁：個人資料、顯示名稱更新、聊天室分類建立與分類列顯示
- 文字訊息送出、`Enter` 快捷送出
- 圖片與檔案訊息上傳
- 桌機版訊息泡泡右鍵操作選單：自己或對方訊息都可引用回覆，收回僅限自己送出的訊息；收回以 `is_recalled` 隱藏，不會實體刪除訊息資料；回覆摘要與左側對話列表預覽只取該訊息本身內容
- MySQL 設定讀取、driver 初始化與 migration 自動執行
- 外部系統註冊 / 登入 API
- session token 簽發與 Bearer 驗證
- 顯示名稱更新 API
- system_admin 裝置查詢與 IP 白名單管理 API
- system_admin 使用者列表 API，支援帳號、暱稱、Email、狀態、來源、建立日期與排序篩選
- 管理後台登入頁與 `/office`、`/office/user`、`/office/conversations`、`/office/admins` 路由
- `/office/user` 使用者管理畫面、資料表列表與可設定欄位寬度
- 使用者列表支援總筆數、分頁及每頁 10、20、50、100 筆切換
- `/office/user/create` 新增使用者頁面與 `POST /api/system-admin/users`
- `/office/user/{user_id}/edit` 編輯使用者頁面與使用者查詢、更新 API；編輯時密碼留空會保留原密碼
- trusted device 核准 API
- 使用者、session、裝置、IP 白名單、聊天核心資料表 migration
- 裝置與 IP 白名單的基本驗證邏輯
- 基本單元測試
- 三種 Web 形態骨架已實作：桌機全畫面、手機全畫面、可嵌入其他系統的小浮動視窗

目前尚未完成：

- 檔案留存策略
- 管理後台的使用者設備、IP 白名單與操作日誌完整 UI
- 管理後台的對話管理與管理員管理頁面，目前仍為占位畫面
- 手機版 / 嵌入式頁面的聊天資料串接
- 三語系完整文案

## 目錄結構

```text
.
├─ db/
│  └─ migrations/
├─ docs/
├─ chat_admin/
│  ├─ static/
│  └─ templates/
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

目前服務預設監聽 `http://0.0.0.0:8080`，可透過 `.env` 的 `LISTEN_HOST` 與 `PORT` 調整。若要直接由 Go server 啟用 HTTPS，需同時設定 `TLS_CERT_FILE` 與 `TLS_KEY_FILE`，程式會改用 `https://` 啟動；若只設定其中一個會啟動失敗以避免誤用。

開發時若想要自動重啟 server，可使用 `air`：

```bash
air -c .air.toml
```

或：

```bash
make dev
```

若本機尚未安裝 `air`，可先安裝後再使用；專案已提供 [`.air.toml`](.air.toml) 設定，會在 `.go`、`.html`、`.css`、`.js`、`.md`、`.toml` 變更時自動重建與重啟。

若要啟用 MySQL 與外部系統接入 / 管理 API：

```bash
LISTEN_HOST=0.0.0.0
PORT=8080
# TLS_CERT_FILE=certs/server.crt
# TLS_KEY_FILE=certs/server.key
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
REQUIRE_TRUSTED_DEVICE=false
```

建議把以上內容放進根目錄 `.env` 後直接執行 `go run .`。

`REQUIRE_TRUSTED_DEVICE=false` 會關閉新瀏覽器 / 新裝置登入核准，方便 MVP 測試；若要恢復「新裝置需由既有信任裝置或 system_admin 核准」流程，改成 `true`。

若未設定 `INTEGRATION_SHARED_TOKEN` 與 `INTEGRATION_SIGNATURE_SECRET`，目前程式會以開發模式允許未帶整合驗證 header 的外部系統 register / login 請求。文件語意已改為統一外部整合能力；路由命名後續建議由 `/api/erp/*` 收斂為更中性的 `/api/integrations/*` 或 `/api/external/*`。

## 桌機測試頁

`/chat` 目前已接上桌機版 Web 測試流程：

- 未登入或 session 過期時導向 `/login`
- 登入後載入對話列表
- 左側上方搜尋框可搜尋使用者與聊天記錄
- 搜尋使用者後，點選使用者會先比對既有一對一對話；若已有對話會直接載入，若尚未建立則先開啟暫存聊天畫面，送出第一則訊息時才建立一對一對話
- 左側右下角新增入口可建立多人群組：新增成員、輸入群組名稱後建立群組對話
- 左上 icon 可打開左欄下拉選單，第一列顯示登入帳號並進入設定頁
- 設定頁目前提供個人資料、顯示名稱更新、頭像預覽與聊天室分類前端原型
- 聊天室分類目前保存在瀏覽器 `localStorage`，並顯示在左側 `Conversations` 上方
- `ALL` 為預設分類，固定顯示全部對話
- 聊天室分類與新建資料夾介面已調整為接近 Telegram Web 的流程
- 傳送文字訊息與可點擊網址
- `Enter` 送出、`Shift + Enter` 換行
- 透過訊息輸入框右側附件按鈕選擇圖片或文件，選檔後可在預覽視窗輸入說明文字並送出；可一次選取多個檔案，送出後會保留在同一則訊息中；說明文字一開始顯示一行，最多顯示六行後可捲動查看
- 可將圖片或檔案拖放到對話頁面，拖入時顯示文件 / 照片兩個上傳區，放開後進入同一個預覽與送出流程；多檔會以同一則訊息送出
- 對任一訊息泡泡點右鍵會顯示接近 Telegram Web 的操作選單，目前支援自己或對方訊息引用回覆，收回僅限自己送出的訊息；收回會將訊息標記為 `is_recalled` 並在列表、搜尋與訊息區隱藏，不會實體刪除資料；回覆預覽會整合在同一個輸入框區塊內，且回覆已含引用的訊息時只帶入該訊息本身內容，不遞迴帶入更早引用鏈；左側對話列表顯示最後訊息時也只顯示回覆正文，不顯示引用前綴與引用內容
- WebSocket 即時刷新對話列表與目前訊息區

目前限制：

- 設定頁與聊天室分類仍屬前端原型，尚未接正式設定 API
- 聊天室分類目前保存在瀏覽器 `localStorage`，屬於 MVP 階段的前端暫存實作
- `ALL` 與自訂分類已可影響左側對話列表顯示，但資料仍未持久化到後端
- 頭像上傳仍為前端原型，尚未接正式後端儲存
- 若要進正式版，仍需補設定資料表、設定 API 與前端 API 串接

附件格式目前支援：

- 圖片：`.jpg`、`.jpeg`、`.png`、`.webp`、`.gif`、`.heic`、`.heif`
- 文件：`.doc`、`.docx`、`.xls`、`.xlsx`、`.ppt`、`.pptx`、`.pdf`
- 文字與資料：`.txt`、`.csv`、`.rtf`、`.md`
- OpenDocument：`.odt`、`.ods`、`.odp`
- iWork：`.pages`、`.numbers`、`.key`
- 壓縮檔：`.zip`、`.rar`、`.7z`

單檔上限 `40MB`。

## 目前可用端點

- `GET /`
- `GET /home`
- `GET /login`
- `GET /chat`
- `GET /m/chat`
- `GET /embed/chat`
- `GET /healthz`
- `GET /ws`
- `GET /api/conversations`
- `POST /api/conversations/direct`
- `POST /api/conversations/group`
- `GET /api/contacts`
- `POST /api/contacts`
- `GET /api/users/search`
- `GET /api/messages/search`
- `PATCH /api/conversations/{conversation_id}`
- `GET /api/conversations/{conversation_id}/members`
- `POST /api/conversations/{conversation_id}/members`
- `DELETE /api/conversations/{conversation_id}/members`
- `GET /api/conversations/{conversation_id}/messages`
- `POST /api/conversations/{conversation_id}/messages`
- `POST /api/conversations/{conversation_id}/read`
- `POST /api/conversations/{conversation_id}/messages/{message_id}/recall`
- `DELETE /api/conversations/{conversation_id}/messages/{message_id}`
- `GET /uploads/{filename}`
- `POST /api/erp/register`
- `POST /api/erp/login`
- `PATCH /api/users/me/profile`
- `GET /api/system-admin/users`
- `GET /api/system-admin/users/{user_id}/devices`
- `PUT /api/system-admin/users/{user_id}/ip-whitelist`
- `POST /api/users/{user_id}/devices/{device_id}/approve`

## 管理後台

- 登入入口：`/admin/login`
- 後台首頁：`/office`
- 使用者管理：`/office/user`
- 對話管理：`/office/conversations`（目前為占位畫面）
- 管理員管理：`/office/admins`（目前為占位畫面）
- 舊路由 `/admin/dashboard` 會重新導向 `/office`

目前 `/office/user` 會從 `users` 資料表載入使用者，並以 `auth_sessions.last_used_at` 顯示最近在線時間。詳細說明請見 [`chat_admin/README.md`](chat_admin/README.md)。

## Web 介面形態

系統需自帶三種前端形態，且共用同一套聊天能力、登入狀態與權限規則：

- 桌機版全畫面：以對話列表 + 內容區為主的雙欄工作區
- 手機版全畫面：以單欄頁面切換為主，優先處理對話列表與訊息頁切換
- 嵌入式小浮動視窗：可插在其他系統頁面中的小型聊天入口與浮動對話窗

目前骨架路由：

- `/home`：登入後入口導覽頁
- `/login`：獨立登入頁
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

完整對接範例請見 [`docs/integration-signature.md`](docs/integration-signature.md)。

登入成功後，`/api/erp/login` 會回傳 session token。這個端點目前是外部系統接入的第一個相容入口；後續聊天 API、管理 API 與裝置核准 API 使用：

- `Authorization: Bearer <session-token>`

說明：

- `user` 可使用聊天 API 與後續聊天頁面
- `system_admin` 可使用管理 API，但不屬於聊天主流程參與者，聊天 API 應回 `403 SYSTEM_ADMIN_CANNOT_CHAT`

Web 測試頁目前是以「帳號輸入」方式包裝既有外部 API：

- 頁面仍然呼叫 `POST /api/erp/login`
- 僅允許既有使用者登入，不會因帳號不存在而由前端自動註冊
- 成功後把 session token 保存在瀏覽器，供後續聊天 API 與 WebSocket 使用
- 密碼需為 4 到 20 碼，只允許英文、數字與特殊符號

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
