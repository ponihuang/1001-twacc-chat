# TWACC 聊天系統 - 管理後台

## 概述

系統管理後台（Admin Interface）是一個獨立的 web 應用程式，用於管理聊天系統的系統管理員相關操作，包括：

- 👥 **使用者管理** - 查詢使用者資訊
- 🖥️ **設備管理** - 查詢使用者登入設備、核准信任設備
- 🛡️ **IP 白名單** - 配置使用者允許登入的 IP 範圍
- 📋 **操作日誌** - 查看管理員操作記錄
- 🏥 **系統狀態** - 檢查系統健康狀態

## 目錄結構

```
chat_admin/
├── static/
│   ├── admin.css          # 管理後台樣式表
│   ├── admin-auth.js      # 認證模組
│   └── admin-app.js       # 主要應用邏輯
└── templates/
    ├── admin-login.html   # 登入頁面
    └── admin-dashboard.html # 儀表板頁面
```

## 使用方法

### 1. 訪問管理後台

在瀏覽器中進入：`http://localhost:8080/admin/login`

### 2. 登入

管理員需要提供：
- **來源系統** (source_system) - 例如：`erp`, `crm`
- **外部系統使用者 ID** (external_user_id) - 在外部系統中的唯一識別碼
- **密碼** - 登入密碼

登入會驗證：
- 認證資訊有效性
- 使用者是否為 `system_admin` 角色

成功登入後，頁面會跳轉到儀表板。

### 3. 儀表板功能

#### 使用者管理
- 查詢系統中的使用者
- 搜尋使用者 ID、名稱或 Email
- 查看使用者建立時間和最後登入時間

#### 設備管理
1. 輸入使用者 ID
2. 點擊「搜尋設備」查詢該使用者的登入設備列表
3. 檢視設備資訊：
   - Device ID
   - User Agent 資訊
   - 登入 IP 位址
   - 首次/最後登入時間
   - 信任狀態
4. 核准待信任的設備

**核准設備流程：**
```
新設備 → 待核准 (Pending) → 管理員核准 → 已信任 (Trusted)
```

#### IP 白名單管理
1. 輸入要設定的使用者 ID
2. 點擊「載入設定」取得該使用者的當前白名單
3. 選擇設定方式：
   - **允許所有 IP** - 開啟此選項允許從任何 IP 登入
   - **限制 IP 列表** - 新增特定 IP 或 IP 段
4. 支援的 IP 格式：
   - IPv4: `192.168.1.10`
   - IPv6: `2001:db8::10`
   - CIDR: `10.0.0.0/24`
5. 點擊「保存設定」更新設置

#### 操作日誌
查詢管理員的操作記錄，支援以下篩選：
- 設備核准 (device_approval)
- 白名單更新 (whitelist_update)
- 使用者查詢 (user_query)

#### 系統狀態
查看系統的健康狀態檢查結果：
- API 服務狀態
- 資料庫連接狀態
- WebSocket 伺服器狀態
- 伺服器版本、運行時間、記憶體使用量

## API 整合

管理後台與後端 API 的對應關係：

### 設備查詢
```
GET /api/system-admin/users/{user_id}/devices
Header: Authorization: Bearer {session_token}
```

回應格式：
```json
{
  "success": true,
  "code": "DEVICES_OK",
  "data": [
    {
      "device_id": "browser-uuid",
      "user_agent": "Mozilla/5.0...",
      "ip": "192.168.1.10",
      "first_login_at": "2026-04-01T08:00:00Z",
      "last_login_at": "2026-04-01T09:00:00Z",
      "is_trusted_device": true
    }
  ]
}
```

### IP 白名單更新
```
PUT /api/system-admin/users/{user_id}/ip-whitelist
Header: Authorization: Bearer {session_token}
Body: {
  "allow_all": false,
  "rules": ["192.168.1.10", "10.0.0.0/24"]
}
```

### 設備核准
```
POST /api/users/{user_id}/devices/{device_id}/approve
Header: Authorization: Bearer {session_token}
```

## 登入流程

1. **前端登入表單**
   - 使用者輸入來源系統、外部使用者 ID 和密碼

2. **後端驗證**
   - 調用 `POST /api/erp/login` 端點
   - 驗證認證資訊
   - 檢查使用者是否為 system_admin
   - 簽發 session token

3. **Token 儲存**
   - 前端將 session token 儲存到 localStorage
   - 後續 API 請求都會攜帶此 token

4. **自動重導**
   - 未登入使用者訪問受保護頁面會自動導向登入頁
   - Session 過期時自動登出

## 安全性設計

### 認證機制
- 使用 Bearer Token 驗證
- Token 儲存在瀏覽器 localStorage 中
- 每個 API 請求都需要帶有有效的 session token

### 角色檢查
- 只有標記為 `system_admin` 的使用者能訪問管理後台
- API 層面進行權限驗證
- 前端層面進行防呆檢查

### 操作限制
- System Admin 無法使用聊天功能
- System Admin 無法加入對話
- Manager 操作會被記錄到日誌中

## 開發指南

### 新增功能

1. **在 HTML 中添加 UI 元素**（admin-dashboard.html）
2. **在 CSS 中添加樣式**（admin.css）
3. **在 JS 中實現邏輯**（admin-app.js / admin-auth.js）

### 例子：新增使用者功能

```javascript
// 在 admin-app.js 中
async function createUser(userData) {
  try {
    const response = await makeAuthenticatedRequest(
      '/api/system-admin/users',
      {
        method: 'POST',
        body: JSON.stringify(userData)
      }
    );
    
    const data = await response.json();
    if (data.success) {
      showSuccess('使用者建立成功');
      loadUsers();
    }
  } catch (err) {
    showError('建立失敗: ' + err.message);
  }
}
```

### API 錯誤處理

常見的錯誤碼：
- `INSUFFICIENT_ROLE` - 帳號沒有足夠權限
- `INVALID_REQUEST` - 請求格式錯誤
- `IP_NOT_ALLOWED` - 登入 IP 不在白名單
- `DEVICE_NOT_TRUSTED` - 設備未被信任
- `SERVICE_UNAVAILABLE` - 服務暫時不可用

## 部署注意事項

1. **靜態資源路由**
   - 確保 `/admin/static/` 路由正確指向 `chat_admin/static/` 目錄
   - CSS 和 JS 文件應該能正確載入

2. **模板位置**
   - 管理後台模板位於 `chat_admin/templates/` 目錄
   - 與主應用的 `web/templates/` 目錄分開

3. **環境變數**
   - 確保 `INTEGRATION_SHARED_TOKEN` 正確配置
   - 這是登入時使用的認證憑證

4. **CORS 和安全頭部**
   - 如果前後端分開部署，需要配置 CORS
   - 建議配置 CSP、X-Frame-Options 等安全頭部

## 未來功能計劃

- [ ] 使用者搜尋和篩選
- [ ] 批量操作（批量設定白名單等）
- [ ] 自訂報表
- [ ] 審計日誌導出
- [ ] 登入歷史分析
- [ ] 實時通知和警報
- [ ] 多語言支援

## 故障排查

### 無法登入
1. 檢查來源系統和使用者 ID 是否正確
2. 確認帳號確實是 system_admin 角色
3. 檢查瀏覽器控制臺是否有錯誤訊息
4. 驗證後端 API 是否正常運行

### 設備無法查詢
1. 確認使用者 ID 格式正確
2. 檢查該使用者是否真的有登入記錄
3. 驗證管理員權限

### 白名單設定失敗
1. 檢查 IP 格式是否有效
2. 確認至少設定了一條規則或啟用「允許所有 IP」
3. 檢查使用者 ID 是否存在

## 相關文件

- [API 文件](../docs/api.md)
- [授權政策](../docs/authorization.md)
- [主應用 README](../README.md)
