# API 草案

本文件整理 MVP 階段目前已實作與後續預定的主要端點。聊天是產品主核心；外部系統 API 是提供其他網站或業務系統接入本對談系統的統一能力，而不是為 ERP 或任何單一來源系統獨立設計的產品分類。

## 1. 健康檢查

### `GET /healthz`

回應：

```json
{
  "status": "ok"
}
```

## 1.1 Web 入口

- `GET /chat`：桌機版全畫面聊天骨架
- `GET /m/chat`：手機版全畫面聊天骨架
- `GET /embed/chat`：嵌入式浮動視窗骨架

## 2. 外部系統整合

本系統未來可能接入 ERP、CRM、HR 或其他網站，因此外部整合應統一收斂在共通 API 能力下。現階段路由仍採 `/api/erp/*` 命名，但這只是暫時相容名稱，語意上應視為外部系統整合入口。

- `source_system`：外部來源系統代碼，例如 `erp`、`crm`、`hr`
- `external_user_id`：該來源系統中的唯一使用者識別

兩者組合需能唯一定位使用者。

### `POST /api/erp/register`

用途：

- 外部系統自動註冊使用者

最小必填欄位：

- `source_system`
- `external_user_id`
- `display_name`

非強制欄位：

- `email`
- `language`
- `whatsapp_account`
- `telegram_account`

欄位規格：

| 欄位 | 型別 | 必填 | 規則 |
| --- | --- | --- | --- |
| `source_system` | `string` | 是 | 2-50 字元，僅允許英數字、底線、連字號，建議小寫 |
| `external_user_id` | `string` | 是 | 1-100 字元，不可為空白字串；由外部系統保證唯一 |
| `display_name` | `string` | 是 | 1-100 字元，去除前後空白後不可為空 |
| `email` | `string` | 否 | 最長 255 字元 |
| `language` | `string` | 否 | 建議值：`zh-Hant`、`zh-Hans`、`en` |
| `whatsapp_account` | `string` | 否 | 最長 50 字元 |
| `telegram_account` | `string` | 否 | 1-64 字元 |

請求草案：

```json
{
  "source_system": "erp",
  "external_user_id": "A12345",
  "display_name": "王小明",
  "email": "user@example.com",
  "language": "zh-Hant",
  "whatsapp_account": "+886912345678",
  "telegram_account": "@mingwang"
}
```

回應草案：

```json
{
  "success": true,
  "code": "REGISTERED",
  "message": "user registered"
}
```

### `POST /api/erp/login`

用途：

- 外部系統自動登入使用者

最小必填欄位：

- `source_system`
- `external_user_id`
- `device_id`

說明：

- `device_id` 用於裝置識別、登入紀錄與信任裝置流程
- 登入成功後回傳的是可供後續 API 使用的 session token
- `user` 與 `system_admin` 都可取得 session token，但後續可用 API 仍需依角色限制

欄位規格：

| 欄位 | 型別 | 必填 | 規則 |
| --- | --- | --- | --- |
| `source_system` | `string` | 是 | 規則同註冊 API |
| `external_user_id` | `string` | 是 | 規則同註冊 API |
| `device_id` | `string` | 是 | 1-128 字元，建議使用 UUID；不可為空白字串 |

請求草案：

```json
{
  "source_system": "erp",
  "external_user_id": "A12345",
  "device_id": "uuid-value"
}
```

回應草案：

```json
{
  "success": true,
  "code": "LOGIN_OK",
  "message": "login success",
  "token": "session-token",
  "expires_at": "2026-03-27T12:00:00Z",
  "data": {
    "role": "user"
  }
}
```

外部系統對接 header 規格草案：

| Header | 必填 | 規則 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer <shared-token>` 格式；MVP 先採固定 shared token |
| `X-Request-Timestamp` | 否 | 後續可用於防重放 |
| `X-Signature` | 否 | 後續可用於驗證請求完整性 |

## 3. 聊天

### `GET /api/conversations`

用途：

- 取得使用者的對話列表

### `GET /api/conversations/{conversation_id}/messages`

用途：

- 取得對話訊息列表

### `POST /api/conversations/{conversation_id}/messages`

用途：

- 傳送文字訊息

請求草案：

```json
{
  "type": "text",
  "content": "hello"
}
```

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- 目前只接受 `type = text`
- `content` 去除前後空白後不可為空
- actor 必須是對話成員
- actor 若為 `system_admin`，需回 `403 SYSTEM_ADMIN_CANNOT_CHAT`

成功回應草案：

```json
{
  "success": true,
  "code": "MESSAGE_SENT",
  "message": "讯息发送成功",
  "data": {
    "conversation_id": 123,
    "message": {
      "message_id": 456,
      "sender_id": 7,
      "sender_name": "王小明",
      "message_type": "text",
      "content": "hello",
      "created_at": "2026-04-07T03:00:00Z"
    }
  }
}
```

補充原則：

- 對外整合不另開 ERP 專屬前端視窗
- 其他系統若需要頁面內聊天能力，應使用嵌入式浮動視窗或直接呼叫 API
- 桌機、手機、嵌入式三種前端形態共用同一套聊天 API

## 4. 群組

### `POST /api/groups`

用途：

- 建立群組

### `POST /api/groups/{group_id}/members`

用途：

- 邀請成員加入群組

## 5. 檔案上傳

### `POST /api/uploads`

用途：

- 上傳圖片或一般檔案

規則草案：

- 單檔上限 `40MB`
- 允許 `.jpg`、`.png`、`.doc`、`.pdf`、`.xlsx`

## 6. 管理與安全功能

### `GET /api/system-admin/users/{user_id}/devices`

用途：

- 查詢使用者近期裝置紀錄

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- actor 需先透過 `POST /api/erp/login` 取得 session token
- actor 必須是 `system_admin`

回應草案：

```json
{
  "success": true,
  "code": "DEVICES_OK",
  "message": "装置列表读取成功",
  "data": [
    {
      "device_id": "browser-uuid",
      "user_agent": "Mozilla/5.0",
      "ip": "192.168.1.10",
      "first_login_at": "2026-04-01T08:00:00Z",
      "last_login_at": "2026-04-01T09:00:00Z",
      "is_trusted_device": true
    }
  ]
}
```

### `PUT /api/system-admin/users/{user_id}/ip-whitelist`

用途：

- 更新指定使用者的 IP 白名單規則

規則補充：

- 支援單一 IP 逐筆輸入
- 支援 IPv4、IPv6 與 CIDR
- 需提供 `allow_all` 開關；開啟時直接放行，不比對規則
- actor 必須是 `system_admin`
- 認證方式為 `Authorization: Bearer <session-token>`

請求草案：

```json
{
  "allow_all": false,
  "rules": [
    "192.168.1.10",
    "2001:db8::10",
    "10.0.0.0/24"
  ]
}
```

回應草案：

```json
{
  "success": true,
  "code": "IP_WHITELIST_UPDATED",
  "message": "IP 白名单更新成功",
  "data": {
    "allow_all": false,
    "rules": [
      "192.168.1.10",
      "2001:db8::10",
      "10.0.0.0/24"
    ]
  }
}
```

### `POST /api/users/{user_id}/devices/{device_id}/approve`

用途：

- 將指定裝置核准為 trusted device

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- 若 actor 是 `system_admin`，可直接核准
- 若 actor 是使用者本人，該 session 對應的登入裝置必須已是 trusted device，且不可核准自己當前這台待核准裝置

回應草案：

```json
{
  "success": true,
  "code": "DEVICE_APPROVED",
  "message": "设备已设为信任设备"
}
```

## 7. 錯誤碼表

### 7.1 通用錯誤

| HTTP Status | Code | 說明 |
| --- | --- | --- |
| `400` | `INVALID_REQUEST` | 請求格式錯誤或必要欄位缺失 |
| `401` | `UNAUTHORIZED` | 未提供合法驗證資訊 |
| `401` | `SESSION_EXPIRED` | session 已過期，需要重新登入 |
| `403` | `FORBIDDEN` | 已驗證但無權限執行此操作 |
| `404` | `NOT_FOUND` | 指定資源不存在 |
| `409` | `CONFLICT` | 資源狀態衝突，無法完成操作 |
| `413` | `FILE_TOO_LARGE` | 上傳檔案超過大小限制 |
| `415` | `UNSUPPORTED_FILE_TYPE` | 檔案類型不被允許 |
| `429` | `RATE_LIMITED` | 請求頻率過高 |
| `500` | `INTERNAL_ERROR` | 系統內部錯誤 |

### 7.2 外部系統註冊與登入

| HTTP Status | Code | 說明 |
| --- | --- | --- |
| `400` | `INVALID_SOURCE_SYSTEM` | `source_system` 不合法或未開通 |
| `400` | `INVALID_EXTERNAL_USER_ID` | `external_user_id` 格式錯誤 |
| `404` | `USER_NOT_FOUND` | 登入時找不到對應使用者 |
| `409` | `USER_ALREADY_EXISTS` | 註冊時該外部使用者已存在 |
| `401` | `INVALID_INTEGRATION_TOKEN` | 外部系統對接憑證錯誤 |
| `401` | `INVALID_SIGNATURE` | 請求簽章驗證失敗 |
| `401` | `REQUEST_EXPIRED` | 請求時間戳過期或不可接受 |
| `200` | `REGISTERED` | 註冊成功 |
| `200` | `LOGIN_OK` | 登入成功 |

### 7.3 裝置與登入限制

| HTTP Status | Code | 說明 |
| --- | --- | --- |
| `400` | `DEVICE_ID_REQUIRED` | 缺少 `device_id` |
| `403` | `DEVICE_NOT_TRUSTED` | 目前裝置未被信任且需要額外授權 |
| `403` | `NEW_DEVICE_APPROVAL_REQUIRED` | 新裝置需要既有信任裝置或 `system_admin` 核准 |
| `403` | `IP_NOT_ALLOWED` | 來源 IP 不在使用者白名單中 |
| `403` | `LOGIN_BLOCKED` | 因安全規則限制而拒絕登入 |

### 7.4 群組與權限

| HTTP Status | Code | 說明 |
| --- | --- | --- |
| `403` | `SYSTEM_ADMIN_CANNOT_CHAT` | `system_admin` 角色不可直接參與聊天 |
| `403` | `INSUFFICIENT_ROLE` | 目前角色不足以執行該操作 |
| `404` | `CONVERSATION_NOT_FOUND` | 找不到指定對話 |
| `404` | `GROUP_NOT_FOUND` | 找不到指定群組 |
| `409` | `MEMBER_ALREADY_EXISTS` | 成員已在群組中 |

### 7.5 建議錯誤回應格式

```json
{
  "success": false,
  "code": "IP_NOT_ALLOWED",
  "message": "目前登入 IP 不在白名單中，請聯絡管理員或確認白名單設定"
}
```

