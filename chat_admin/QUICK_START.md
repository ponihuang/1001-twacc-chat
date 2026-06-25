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

管理員必須：

1. 已存在於 `users`
2. 已設定可登入密碼
3. 存在於 `system_admin_users`

建議先透過既有註冊流程建立使用者，再授予管理員角色：

```sql
SELECT id, source_system, external_user_id, display_name
FROM users
WHERE source_system = 'erp' AND external_user_id = 'admin';

INSERT INTO system_admin_users (user_id)
VALUES (<查到的 user id>);
```

若資料已存在，請勿重複新增。不要直接寫入明文密碼；密碼需由系統註冊流程產生合法的 `password_hash`。

## 3. 登入

開啟：

```text
http://127.0.0.1:8080/admin/login
```

目前後台登入表單需要：

- 帳號：對應 `users.external_user_id`
- 密碼：註冊時設定的密碼

登入代理目前固定使用 `source_system=erp`，成功後導向：

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

## 5. 目前限制

- 使用者列表的「編輯」可進入 `/office/user/{user_id}/edit` 查看不可修改的帳號，並修改密碼、暱稱、Email 與狀態；密碼留空時保留原密碼。
- `/office/conversations` 與 `/office/admins` 目前是占位頁。
- 設備與 IP 白名單 API 已存在，但新版後台尚未提供對應頁面。
- 使用者列表尚未分頁，目前後端最多回傳 500 筆。

## 6. 常見問題

### 登入後立即被導回登入頁

- 確認帳號存在於 `system_admin_users`
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
go test ./...
node --check chat_admin/static/admin-app.js
git diff --check
```
