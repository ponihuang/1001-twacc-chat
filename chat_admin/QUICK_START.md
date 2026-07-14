# 管理後台快速啟動

## 1. 啟動服務

在專案根目錄執行：

```bash
go run .
```

預設網址：

```text
http://127.0.0.1:8080
```

服務需要 MySQL 才能登入並載入使用者列表。環境變數設定請參考根目錄 [README.md](../README.md)。

## 2. 準備管理員帳號

第一次部署時，先用 CLI 建立第一個管理員帳號：

```bash
go run . seed-admin --account admin01 --password Admin123! --name 系統管理員
```

此指令會：

1. 讀取 `.env` 的 MySQL 設定
2. 先套用 migration
3. 建立 `admin_users` 後台管理員帳號
4. 後台登入 session 會寫入 `admin_auth_sessions`

指令可重複執行；若帳號已存在，不會覆蓋既有密碼，只會確認管理員帳號存在。不要直接用 SQL 寫入明文密碼，密碼需由系統產生合法的 `password_hash`。

## 3. 登入

開啟：

```text
http://127.0.0.1:8080/admin/login
```

目前後台登入表單需要：

- 帳號：對應 `admin_users.account`
- 密碼：註冊時設定的密碼

登入時只查後台管理員資料表，不會用聊天室玩家帳號登入。成功後導向：

```text
http://127.0.0.1:8080/office
```

## 4. 使用者管理

開啟：

```text
http://127.0.0.1:8080/office/user
```

頁面會呼叫：

```http
GET /api/system-admin/users
Authorization: Bearer <session-token>
```

列表資料來自 `users`，最近在線來自 `auth_sessions.last_used_at`。

可用篩選：

- 帳號
- 暱稱
- Email
- 狀態
- 來源
- 開始與結束日期
- 建立時間或最近在線排序

註冊邀請管理：

```text
http://127.0.0.1:8080/office/user/invitations
```

可單筆新增邀請、重送待註冊邀請，或點「批量新增」上傳 `.txt`。TXT 每行一個 Email，選檔後會先預檢並分類為可發送、檔案內重複、已註冊、已有有效邀請與格式錯誤；確認後才會批量發送。批量上傳限制為 request body 最大 `2MB`，有效且唯一 Email 最多 `1000` 筆。

## 5. 目前限制

- 使用者列表的「編輯」可進入 `/office/user/{user_id}/edit` 查看不可修改的帳號，並修改密碼、暱稱、Email 與狀態；密碼留空時保留原密碼。
- 註冊邀請批量發送會逐筆寄信，若 SMTP 回應較慢，大量發送時畫面會等待後端完成整批結果。
- `/office/conversations` 目前是占位頁；`/office/admins` 已提供管理員列表、新增與編輯。
- 設備與 IP 白名單 API 已存在，但新版後台尚未提供對應頁面。
- 使用者列表預設每頁顯示 10 筆，並可切換為 20、50 或 100 筆。

## 6. 常見問題

### 登入後立即被導回登入頁

- 確認帳號存在於 `admin_users`
- 確認瀏覽器 Local Storage 有 `admin_session_token`
- 確認 session 尚未過期

### 使用者列表顯示錯誤

- 確認 MySQL 已啟動
- 確認 `ENABLE_MYSQL=true`
- 確認 `users` 資料表有資料
- 在瀏覽器 Network 檢查 `/api/system-admin/users`

### 修改 CSS 或 JavaScript 後畫面未更新

- 重新整理頁面
- 必要時更新 `admin-dashboard.html` 的靜態資源 query version
- 清除瀏覽器快取後再測試

## 7. 驗證

```bash
go test ./internal/erp ./internal/store/mysql ./internal/httpx
node --check chat_admin/static/admin-app.js
git diff --check
```
