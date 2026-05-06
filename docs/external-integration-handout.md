# 聊天系統測試機串接範例

本文件提供外部系統工程師直接測試 `register` / `login` 用。

測試環境請替換：

- `https://chat-test.example.com`
- `your-shared-token`
- `your-signature-secret`

## 1. Register API

### Request

```http
POST /api/erp/register
Authorization: Bearer your-shared-token
Content-Type: application/json
X-TWACC-Timestamp: <unix-seconds>
X-TWACC-Signature: <hex-hmac-sha256>
```

```json
{
  "source_system": "erp",
  "external_user_id": "A12345",
  "password": "pass123!",
  "display_name": "王小明"
}
```

### curl 範例

```bash
TIMESTAMP=$(date +%s)
BODY='{"source_system":"erp","external_user_id":"A12345","password":"pass123!","display_name":"王小明"}'
SIGNATURE=$(printf "POST\n/api/erp/register\n%s\n%s" "$TIMESTAMP" "$BODY" \
  | openssl dgst -sha256 -hmac "your-signature-secret" -binary \
  | xxd -p -c 256)

curl -X POST "https://chat-test.example.com/api/erp/register" \
  -H "Authorization: Bearer your-shared-token" \
  -H "Content-Type: application/json" \
  -H "X-TWACC-Timestamp: $TIMESTAMP" \
  -H "X-TWACC-Signature: $SIGNATURE" \
  -d "$BODY"
```

## 2. Login API

### Request

```http
POST /api/erp/login
Authorization: Bearer your-shared-token
Content-Type: application/json
X-TWACC-Timestamp: <unix-seconds>
X-TWACC-Signature: <hex-hmac-sha256>
```

```json
{
  "source_system": "erp",
  "external_user_id": "A12345",
  "password": "pass123!",
  "device_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### curl 範例

```bash
TIMESTAMP=$(date +%s)
BODY='{"source_system":"erp","external_user_id":"A12345","password":"pass123!","device_id":"550e8400-e29b-41d4-a716-446655440000"}'
SIGNATURE=$(printf "POST\n/api/erp/login\n%s\n%s" "$TIMESTAMP" "$BODY" \
  | openssl dgst -sha256 -hmac "your-signature-secret" -binary \
  | xxd -p -c 256)

curl -X POST "https://chat-test.example.com/api/erp/login" \
  -H "Authorization: Bearer your-shared-token" \
  -H "Content-Type: application/json" \
  -H "X-TWACC-Timestamp: $TIMESTAMP" \
  -H "X-TWACC-Signature: $SIGNATURE" \
  -d "$BODY"
```

## 3. Postman 前置腳本

將下列 script 放在 Postman Pre-request Script：

```javascript
const crypto = require('crypto');

const method = pm.request.method;
const path = pm.request.url.getPath();
const timestamp = Math.floor(Date.now() / 1000).toString();
const body = pm.request.body ? pm.request.body.raw : '';
const secret = pm.environment.get('integration_signature_secret');

const message = [method, path, timestamp, body].join('\n');
const signature = crypto
  .createHmac('sha256', secret)
  .update(message)
  .digest('hex');

pm.request.headers.upsert({
  key: 'X-TWACC-Timestamp',
  value: timestamp,
});

pm.request.headers.upsert({
  key: 'X-TWACC-Signature',
  value: signature,
});
```

另外固定帶：

```http
Authorization: Bearer {{integration_shared_token}}
Content-Type: application/json
```

Environment 變數：

- `integration_shared_token`
- `integration_signature_secret`

## 4. 簽章規則

簽章原文格式：

```text
HTTP_METHOD + "\n" + REQUEST_PATH + "\n" + TIMESTAMP + "\n" + RAW_BODY
```

範例：

```text
POST
/api/erp/register
1776384000
{"source_system":"erp","external_user_id":"A12345","password":"pass123!","display_name":"王小明"}
```

規則說明：

- `REQUEST_PATH` 只取 path，例如 `/api/erp/register`
- 不含 domain、protocol、query string
- `RAW_BODY` 必須與實際送出的 JSON 完全一致
- timestamp 使用 Unix 秒數
- 伺服器預設容許誤差 `5 分鐘`

## 5. 常見錯誤碼

| Code | 說明 |
| --- | --- |
| `INVALID_INTEGRATION_TOKEN` | `Authorization` 錯誤或缺少 |
| `INTEGRATION_TIMESTAMP_REQUIRED` | 缺少 `X-TWACC-Timestamp` |
| `INVALID_INTEGRATION_TIMESTAMP` | timestamp 格式錯誤 |
| `INTEGRATION_TIMESTAMP_EXPIRED` | timestamp 過期 |
| `INTEGRATION_SIGNATURE_REQUIRED` | 缺少 `X-TWACC-Signature` |
| `INVALID_INTEGRATION_SIGNATURE` | 簽章錯誤 |

## 6. 成功後下一步

1. 先呼叫 `POST /api/erp/register`
2. 再呼叫 `POST /api/erp/login`
3. 取得 `token`
4. 後續聊天 API 使用：

```http
Authorization: Bearer <session-token>
```

可再接：

- `GET /api/conversations`
- `POST /api/conversations/direct`
- `GET /api/conversations/{conversation_id}/messages`
- `POST /api/conversations/{conversation_id}/messages`
