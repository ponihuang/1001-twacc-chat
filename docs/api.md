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

- `GET /chat`：桌機版全畫面聊天測試頁
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
- `password`
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
| `password` | `string` | 是 | 4-20 字元；只允許英文、數字、特殊符號；不可包含空白 |
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
  "password": "pass123!",
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
- `password`
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
| `password` | `string` | 是 | 規則同註冊 API |
| `device_id` | `string` | 是 | 1-128 字元，建議使用 UUID；不可為空白字串 |

請求草案：

```json
{
  "source_system": "erp",
  "external_user_id": "A12345",
  "password": "pass123!",
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

外部系統對接 header 規格：

| Header | 必填 | 規則 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer <shared-token>` 格式 |
| `X-TWACC-Timestamp` | 是 | Unix timestamp 秒數，例如 `1776384000` |
| `X-TWACC-Signature` | 是 | `hex(HMAC-SHA256(...))` |

簽章字串格式：

```text
HTTP_METHOD + "\n" + REQUEST_PATH + "\n" + TIMESTAMP + "\n" + RAW_BODY
```

補充說明請見 [integration-signature.md](/Users/elva/新公司/聊天系統/1001-twacc-chat-main/docs/integration-signature.md)。

## 3. 聊天

### `GET /api/conversations`

用途：

- 取得使用者的對話列表

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- 回應項目包含 `conversation_id`、`type`、`title`、`member_count`
- 若已有最後一則訊息，會附帶 `last_message_type`、`last_message_preview`、`last_message_at`
- `unread_count` 為一般未讀數；`has_unread_mention` 代表未讀訊息中有標註目前使用者或 `@ALL`

### `POST /api/conversations/direct`

用途：

- 建立或取得一對一對話

請求草案：

```json
{
  "source_system": "erp",
  "external_user_id": "user_b"
}
```

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- actor 需為一般聊天使用者，`system_admin` 不可建立聊天
- 若指定使用者不存在，回 `404 USER_NOT_FOUND`
- 若一對一對話已存在，回 `200 CONVERSATION_EXISTS`
- 若新建成功，回 `201 CONVERSATION_CREATED`
- 建立或載入成功後，WebSocket 會對對話成員推送 `conversation.ready` 事件，讓雙方可即時看到這個對話

### `POST /api/conversations/group`

用途：

- 建立多人群組對話

請求草案：

```json
{
  "name": "採購小組",
  "members": [
    {
      "source_system": "erp",
      "external_user_id": "user_b"
    },
    {
      "source_system": "erp",
      "external_user_id": "user_c"
    }
  ]
}
```

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- actor 需為一般聊天使用者，`system_admin` 不可建立聊天
- `name` 去除前後空白後不可為空
- `members` 至少需包含一位有效成員；actor 會自動加入並設為 `owner`
- 成員會以 `source_system` + `external_user_id` 查找 active 使用者
- 建立成功後寫入 `conversations.type = group` 與多筆 `conversation_members`
- 建立成功後，WebSocket 會對所有群組成員推送 `conversation.ready` 事件

成功回應草案：

```json
{
  "success": true,
  "code": "GROUP_CONVERSATION_CREATED",
  "message": "群組對話建立成功",
  "data": {
    "conversation_id": 123,
    "type": "group",
    "title": "採購小組"
  }
}
```

### `GET /api/conversations/{conversation_id}/members`

用途：

- 取得群組或對話的成員列表
- 桌機版點選群組對話上方標題區後，右側群組資訊面板會使用此 API 顯示成員

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- actor 必須是該對話成員
- 回傳成員依 `owner`、`admin`、一般成員排序

成功回應草案：

```json
{
  "success": true,
  "code": "CONVERSATION_MEMBERS_OK",
  "message": "成員列表讀取成功",
  "data": {
    "conversation_id": 123,
    "type": "group",
    "title": "採購小組",
    "description": "採購相關討論",
    "member_count": 3,
    "members": [
      {
        "user_id": 7,
        "source_system": "erp",
        "external_user_id": "user_a",
        "display_name": "王小明",
        "role": "owner"
      }
    ]
  }
}
```

### `PATCH /api/conversations/{conversation_id}`

用途：

- 更新群組資訊，目前支援群組名稱與描述
- 桌機版群組資訊面板右上角編輯按鈕會開啟編輯頁，送出後呼叫此 API

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- actor 必須是該群組的 `owner` 或 `admin`
- 僅支援 `group` 對話，一對一對話不可透過此 API 更新
- `name` 去除前後空白後不可為空，最長 255 字
- `description` 可空，最長 1000 字
- 更新成功後回傳更新後的群組資訊與成員列表

請求草案：

```json
{
  "name": "採購小組",
  "description": "採購相關討論"
}
```

成功回應草案同 `GET /api/conversations/{conversation_id}/members`。

### `POST /api/conversations/{conversation_id}/members`

用途：

- 將聯絡人加入既有群組對話
- 桌機版群組資訊面板右下角「新增成員」按鈕會開啟成員選擇頁，送出後呼叫此 API

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- actor 必須是該群組成員
- 目前先允許群組內成員新增其他 active 使用者，後續可再補 owner/admin 權限限制
- 重複加入既有成員會被忽略，不會重複寫入
- 新增成功後回傳更新後的成員列表

請求草案：

```json
{
  "members": [
    {
      "source_system": "erp",
      "external_user_id": "user_b"
    }
  ]
}
```

成功回應草案同 `GET /api/conversations/{conversation_id}/members`。

### `DELETE /api/conversations/{conversation_id}/members`

用途：

- 從既有群組對話移除一位成員
- 桌機版群組資訊面板內，對成員點選右鍵後可執行「從群組中移除」

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- actor 必須是該群組的 `owner` 或 `admin`
- 不允許移除群組擁有者，也不允許透過此操作移除自己
- 移除成功後回傳更新後的成員列表

請求草案：

```json
{
  "source_system": "erp",
  "external_user_id": "user_b"
}
```

成功回應草案同 `GET /api/conversations/{conversation_id}/members`。

### `POST /api/contacts`

用途：

- 將一位 active 使用者加入目前使用者的單向聯絡人清單
- 桌機版可從一對一對話右上角選單開啟「加入聯絡人」面板送出

請求草案：

```json
{
  "source_system": "erp",
  "external_user_id": "user_b",
  "alias_name": "王小明"
}
```

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- actor 需為一般聊天使用者，`system_admin` 不可加入聯絡人
- 聯絡人是單向關係，只新增 `owner_user_id -> contact_user_id`
- 不可將自己加入聯絡人
- 重複加入同一人會更新 `alias_name` 並將 `status` 恢復為 `active`

成功回應草案：

```json
{
  "success": true,
  "code": "CONTACT_ADDED",
  "message": "联系人已加入",
  "data": {
    "contact_user_id": 8,
    "source_system": "erp",
    "external_user_id": "user_b",
    "display_name": "王小明",
    "alias_name": "王小明",
    "status": "active"
  }
}
```

### `GET /api/contacts`

用途：

- 取得目前使用者已加入的 active 聯絡人清單
- 桌機版左上角選單的「聯絡人」頁面會使用此 API 顯示名單

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- actor 需為一般聊天使用者，`system_admin` 不可讀取聯絡人
- 只回傳 `user_contacts.status = active` 且對方 `users.status = active` 的資料

成功回應草案：

```json
{
  "success": true,
  "code": "CONTACTS_OK",
  "message": "联系人列表已载入",
  "data": {
    "contacts": [
      {
        "contact_user_id": 8,
        "source_system": "erp",
        "external_user_id": "user_b",
        "display_name": "王小明",
        "alias_name": "王小明",
        "status": "active"
      }
    ]
  }
}
```

### `GET /api/conversations/{conversation_id}/messages`

用途：

- 取得對話訊息列表
- 回傳前會依目前使用者的 `conversation_reads.last_read_message_id` 找出第一筆未讀訊息，放在 `first_unread_message_id`
- 回傳成功後不會自動標記整個對話已讀；前端可用 `first_unread_message_id` 將畫面定位到第一筆未讀訊息，並在該訊息前顯示「未讀訊息」分隔提示
- 前端應在使用者實際看到訊息後呼叫 `POST /api/conversations/{conversation_id}/read` 更新已讀位置

成功回應草案：

```json
{
  "success": true,
  "code": "MESSAGES_OK",
  "message": "訊息列表讀取成功",
  "data": {
    "conversation_id": 123,
    "type": "direct",
    "title": "王小明",
    "first_unread_message_id": 456,
    "messages": [
      {
        "message_id": 456,
        "sender_id": 8,
        "sender_name": "王小明",
        "message_type": "text",
        "content": "hello",
        "created_at": "2026-05-16T00:08:17+08:00"
      }
    ]
  }
}
```

目前限制：

- 訊息列表目前載入最近 `100` 筆；若第一筆未讀早於這批訊息，前端會找不到該訊息並退回定位到最新訊息。

### `POST /api/conversations/{conversation_id}/read`

用途：

- 更新目前使用者在對話中的已讀位置
- 建議由前端在訊息進入可視範圍後呼叫，避免只打開聊天室就全部已讀

請求草案：

```json
{
  "last_read_message_id": 456
}
```

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- actor 必須是該對話成員
- 後端只會往較新的訊息更新，不會把已讀位置往回退

成功回應草案：

```json
{
  "success": true,
  "code": "CONVERSATION_READ_OK",
  "message": "已讀狀態已更新"
}
```

### `POST /api/conversations/{conversation_id}/messages`

用途：

- 傳送文字、圖片或一般檔案訊息

文字訊息請求草案：

```json
{
  "type": "text",
  "content": "hello"
}
```

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- JSON 請求可傳 `type = text`
- 文字訊息支援 `@ALL` 與 `@使用者帳號` 標註；後端會依目前對話成員建立標註提醒
- `multipart/form-data` 可上傳圖片或一般檔案，欄位使用 `file` 與可選的 `content`
- 目前支援圖片 `.jpg`、`.jpeg`、`.png`、`.webp`、`.gif`、`.heic`、`.heif`
- 目前支援文件 `.doc`、`.docx`、`.xls`、`.xlsx`、`.ppt`、`.pptx`、`.pdf`、`.txt`、`.csv`、`.rtf`、`.md`、`.odt`、`.ods`、`.odp`、`.pages`、`.numbers`、`.key`
- 目前支援壓縮檔 `.zip`、`.rar`、`.7z`
- 單檔大小上限 `40MB`
- 純文字訊息的 `content` 去除前後空白後不可為空
- 圖片 / 檔案訊息若未輸入文字，可只送附件
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

附件訊息回應草案：

```json
{
  "success": true,
  "code": "MESSAGE_SENT",
  "message": "讯息发送成功",
  "data": {
    "conversation_id": 123,
    "message": {
      "message_id": 789,
      "sender_id": 7,
      "sender_name": "王小明",
      "message_type": "image",
      "content": "",
      "created_at": "2026-04-07T03:05:00Z",
      "attachment": {
        "original_name": "sample.png",
        "url": "/uploads/sample.png",
        "mime_type": "image/png",
        "size_bytes": 12345
      }
    }
  }
}
```

### `POST /api/conversations/{conversation_id}/messages/{message_id}/recall`

用途：

- 收回自己送出的訊息

目前實作補充：

- 需帶 `Authorization: Bearer <session-token>`
- actor 必須是對話成員
- 只能收回自己送出的訊息
- 收回是軟隱藏：資料庫會將 `messages.is_recalled` 設為 `TRUE`，不實體刪除 `messages` 或 `attachments`
- 訊息列表、對話列表最後訊息、未讀數與訊息搜尋都會排除 `is_recalled = TRUE` 的訊息
- 舊版 `DELETE /api/conversations/{conversation_id}/messages/{message_id}` 目前保留相容，但前端收回流程使用此 `POST .../recall` 端點

成功回應草案：

```json
{
  "success": true,
  "code": "MESSAGE_RECALLED",
  "message": "訊息已收回",
  "data": {
    "conversation_id": 123,
    "message_id": 456
  }
}
```

補充原則：

- 對外整合不另開 ERP 專屬前端視窗
- 其他系統若需要頁面內聊天能力，應使用嵌入式浮動視窗或直接呼叫 API
- 桌機、手機、嵌入式三種前端形態共用同一套聊天 API

### `GET /ws`

用途：

- 建立登入後的 WebSocket 即時通知連線

目前實作補充：

- 連線方式：`ws://host/ws?token=<session-token>` 或 `wss://host/ws?token=<session-token>`
- token 使用既有 `/api/erp/login` 取得的 session token
- WebSocket 目前只負責推送事件通知；建立對話與送訊息仍使用既有 REST API
- 目前事件類型：
  - `conversation.ready`
  - `message.created`
  - `message.recalled`

事件草案：

```json
{
  "event_type": "conversation.ready",
  "conversation_id": 123,
  "conversation": {
    "conversation_id": 123,
    "type": "direct",
    "title": "user_b"
  }
}
```

```json
{
  "event_type": "message.created",
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
```

```json
{
  "event_type": "message.recalled",
  "conversation_id": 123,
  "message_id": 456
}
```

### `GET /uploads/{filename}`

用途：

- 讀取已上傳的圖片或檔案

目前實作補充：

- 後端會把附件存到 `web/uploads/`
- 訊息回應中的 `attachment.url` 會直接指向 `/uploads/{filename}`

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
- 允許圖片 `.jpg`、`.jpeg`、`.png`、`.webp`、`.gif`、`.heic`、`.heif`
- 允許文件 `.doc`、`.docx`、`.xls`、`.xlsx`、`.ppt`、`.pptx`、`.pdf`、`.txt`、`.csv`、`.rtf`、`.md`、`.odt`、`.ods`、`.odp`、`.pages`、`.numbers`、`.key`
- 允許壓縮檔 `.zip`、`.rar`、`.7z`

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
