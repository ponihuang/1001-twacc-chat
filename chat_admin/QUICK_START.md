# 快速啟動指南

## 系統要求

- Go 1.18+
- 現代 web 瀏覽器（Chrome、Firefox、Safari 等）
- MySQL 資料庫（可選，取決於後端配置）

## 啟動應用

### 1. 啟動後端服務

```bash
# 在專案根目錄執行
go run .
```

或使用 air 自動重啟：

```bash
air -c .air.toml
```

或使用 make 命令：

```bash
make dev
```

伺服器預設監聽 `http://0.0.0.0:8080`

### 2. 訪問管理後台

在瀏覽器中打開：

```
http://localhost:8080/admin/login
```

## 登入憑證

為了測試管理後台，您需要有 `system_admin` 角色的帳號。

### 建立測試帳號

1. **直接在資料庫中建立**（開發用）

```sql
-- 首先建立 user
INSERT INTO users (external_system, external_user_id, display_name) 
VALUES ('erp', 'admin', '系統管理員');

-- 取得剛建立的 user id
SELECT last_insert_id();

-- 將該 user 標記為 system_admin
INSERT INTO system_admin_users (user_id) 
VALUES (/*上面查出的 user_id*/);
```

2. **使用 API 呼叫**

```bash
# 第一步：外部系統註冊
curl -X POST http://localhost:8080/api/erp/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer shared-token" \
  -d '{
    "source_system": "erp",
    "external_user_id": "admin",
    "display_name": "系統管理員",
    "password": "admin123"
  }'

# 第二步：登入取得 session token
curl -X POST http://localhost:8080/api/erp/login \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer shared-token" \
  -d '{
    "source_system": "erp",
    "external_user_id": "admin",
    "password": "admin123"
  }'
```

### 登入表單填寫

在登入頁面輸入：

- **來源系統**：`erp`
- **外部系統使用者 ID**：`admin`（或您建立的帳號）
- **密碼**：`admin123`（或您設定的密碼）

## 主要功能演示

### 1. 查詢使用者設備

進入「設備管理」頁面：

1. 輸入使用者 ID（例如 `1`）
2. 點擊「搜尋設備」
3. 查看該使用者的登入設備列表
4. 核准待信任的設備

### 2. 設定 IP 白名單

進入「IP 白名單」頁面：

1. 輸入使用者 ID
2. 點擊「載入設定」
3. 選擇設定方式：
   - 全部允許：勾選「允許從所有 IP 登入」
   - 限制列表：新增 IP 規則
4. 點擊「保存設定」

支援的 IP 格式：
- IPv4：`192.168.1.0`
- CIDR：`10.0.0.0/24`
- IPv6：`2001:db8::0/32`

### 3. 查看操作日誌

進入「操作日誌」頁面查看所有管理員操作記錄。

### 4. 檢查系統狀態

進入「系統狀態」頁面檢查：
- API 服務狀態
- 資料庫連接狀態
- WebSocket 伺服器狀態
- 伺服器資訊

## 開發提示

### 本地開發

1. **開啟開發人員工具**
   - F12 或右鍵「檢查」開啟瀏覽器控制臺
   - 查看 Network 標籤監控 API 呼叫
   - 查看 Console 標籤檢查錯誤

2. **檢查存儲的資料**
   - Application > Local Storage > http://localhost:8080
   - 查看各項 admin_* 鍵值對

3. **測試 API**
   - 使用 cURL、Postman 或 curl 命令行測試 API
   - 確保 Bearer Token 正確

### 常見錯誤

| 錯誤 | 原因 | 解決方案 |
|-----|------|---------|
| 無法連接到伺服器 | 後端未啟動或埠不對 | 確認伺服器運行，檢查埠號 |
| 登入失敗「對接凭证错误」 | 認證 Token 錯誤 | 檢查 `.env` 中的 `INTEGRATION_SHARED_TOKEN` |
| 無法加載頁面 | 靜態資源路徑錯誤 | 確認 CSS/JS 檔案能正確載入 |
| 設備列表為空 | 沒有登入記錄 | 先用聊天應用登入以建立登入紀錄 |

## 下一步

1. **擴展功能**
   - 實作使用者搜尋 API
   - 新增批量操作功能
   - 實現登入日誌查詢

2. **改進 UI/UX**
   - 新增分頁功能
   - 實現表格排序
   - 優化響應式設計

3. **安全性增強**
   - 實現雙因子認證
   - 新增操作審批流程
   - 管理員活動追蹤記錄

4. **監控和報表**
   - 建立實時監控面板
   - 導出 Excel 報表
   - 設定警報規則

## 故障排查

### 靜態文件無法加載

如果看到 404 錯誤，確認文件位置：

```
chat_admin/
├── static/
│   ├── admin.css
│   ├── admin-auth.js
│   └── admin-app.js
└── templates/
    ├── admin-login.html
    └── admin-dashboard.html
```

### Session Token 過期

如果看到「登入已過期」訊息：

1. 點擊頁面上的「登出」按鈕
2. 清除瀏覽器 localStorage：
   ```javascript
   localStorage.clear(); // 在控制臺執行
   ```
3. 重新登入

### API 呼叫失敗

檢查瀏覽器控制臺（F12 > Network）：

1. 檢查請求 URL 是否正確
2. 確認 Authorization header 包含有效 token
3. 檢查回應狀態碼和錯誤訊息

## 聯絡支援

如有問題或建議，請：

1. 查閱 [README.md](README.md) 詳細文件
2. 檢查 [API 文件](../docs/api.md)
3. 查看 [授權政策](../docs/authorization.md)
