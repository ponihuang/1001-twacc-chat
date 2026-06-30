# TWACC 聊天系統管理後台

`chat_admin/` 是聊天室系統的管理後台前端，與聊天室前台 `web/` 共用同一個 Go server 與 MySQL，但後台管理員帳號使用 `admin_users`，聊天室玩家使用 `users`。

## 目前狀態

已完成：

- 管理員登入頁 `/admin/login`
- 登入成功導向 `/office`
- 舊 `/admin/dashboard` 導向 `/office`
- `/office/user` 使用者管理頁
- `/office/user/create` 新增使用者頁
- `/office/user/{user_id}/edit` 編輯使用者頁
- 從 `users` 資料表載入使用者列表
- 以 `auth_sessions.last_used_at` 顯示最近在線
- 帳號、暱稱、Email、狀態、來源、建立日期與排序篩選
- 使用者表格欄位設定 `USER_COLUMNS`
- 暖色系後台版面與共用側欄

尚未完成：

- 對話管理 `/office/conversations`
- 新版介面中的設備、IP 白名單與操作日誌頁面
- 批次操作、稽核記錄與完整多語系
- 依路由拆分模板與頁面 JavaScript

## 目錄

```text
chat_admin/
├── README.md
├── QUICK_START.md
├── static/
│   ├── admin.css
│   ├── admin-auth.js
│   └── admin-app.js
└── templates/
    ├── admin-login.html
    └── admin-dashboard.html
```

目前 `admin-dashboard.html` 是 `/office` 系列路由共用模板。後續功能增加時，建議拆成 layout 與各功能頁模板。

## 路由

| 路由 | 狀態 |
|---|---|
| `/admin/login` | 管理員登入 |
| `/admin/dashboard` | 重新導向 `/office` |
| `/office` | 後台首頁 |
| `/office/user` | 使用者管理，已串接資料 |
| `/office/user/create` | 新增一般使用者 |
| `/office/user/{user_id}/edit` | 編輯一般使用者 |
| `/office/conversations` | 對話管理占位頁 |
| `/office/admins` | 管理員管理，已串接資料 |

## 使用者列表 API

```http
GET /api/system-admin/users
Authorization: Bearer <session-token>
```

可用 query parameters：

- `external_user_id`
- `display_name`
- `email`
- `status`
- `source_system`
- `created_from`，格式 `YYYY-MM-DD`
- `created_to`，格式 `YYYY-MM-DD`
- `sort`：`created_at_desc`、`created_at_asc`、`last_online_desc`
- `page`：頁碼，預設 `1`
- `per_page`：每頁筆數，可用 `10`、`20`、`50`、`100`，預設 `10`

回應包含 `items`、`total`、`page`、`per_page` 與 `total_pages`。

建立使用者：

```http
POST /api/system-admin/users
Authorization: Bearer <session-token>
Content-Type: application/json
```

新增頁欄位依序為帳號、密碼、暱稱、Email 與狀態。頁面固定使用 `source_system=office` 與 `language=zh-Hant`；狀態可選啟用或停用，預設為啟用。

取得與更新單一使用者：

```http
GET /api/system-admin/users/{user_id}
PATCH /api/system-admin/users/{user_id}
Authorization: Bearer <session-token>
```

帳號由 `source_system + external_user_id` 識別，編輯頁僅供查看且不可修改。密碼欄位固定顯示空白；PATCH 時密碼有輸入才更新，留空則保留原密碼。

其他既有管理 API：

```text
GET /api/system-admin/users/{user_id}/devices
PUT /api/system-admin/users/{user_id}/ip-whitelist
POST /api/users/{user_id}/devices/{device_id}/approve
```

上述 API 已存在，但設備與白名單 UI 尚未重新整合進新版 `/office` 畫面。

## 前端結構

- `admin-auth.js`：session token、登入保護與登出
- `admin-app.js`：路由切換、使用者查詢與表格渲染
- `admin.css`：登入頁與後台共用樣式
- `USER_COLUMNS`：使用者表格的欄位名稱、順序與寬度設定

調整使用者表格欄寬時，修改 `admin-app.js`：

```javascript
const USER_COLUMNS = [
  { key: 'index', title: '#', width: 60 },
  { key: 'account', title: '帳號', minWidth: 160 },
  { key: 'email', title: 'Email', minWidth: 260 }
];
```

## 權限

- 管理 API 需帶 Bearer session token。
- 後端會確認操作者存在於 `admin_users` 且角色為 `system_admin`。
- `system_admin` 不可使用聊天 API。
- 前端權限檢查只作為使用體驗防呆，實際權限以後端為準。

## 相關文件

- [主 README](../README.md)
- [快速啟動](QUICK_START.md)
- [API 文件](../docs/api.md)
- [權限文件](../docs/authorization.md)
- [需求文件](../docs/requirements.md)
