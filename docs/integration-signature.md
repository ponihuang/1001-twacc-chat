# 外部系統簽章串接

本文件說明外部系統呼叫接入端點時，如何帶 `Authorization`、`timestamp` 與 `signature`。

目前適用端點：

- `POST /api/erp/register`
- `POST /api/erp/login`

## 1. 必要 Header

| Header | 必填 | 說明 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer <shared-token>` |
| `X-TWACC-Timestamp` | 是 | Unix timestamp 秒數，例如 `1776384000` |
| `X-TWACC-Signature` | 是 | 請求簽章，格式為 `hex(HMAC-SHA256(...))` |

## 2. 簽章字串

簽章原文格式：

```text
HTTP_METHOD + "\n" + REQUEST_PATH + "\n" + TIMESTAMP + "\n" + RAW_BODY
```

範例：

```text
POST
/api/erp/register
1776384000
{"source_system":"erp","external_user_id":"A12345","display_name":"王小明"}
```

注意：

- `REQUEST_PATH` 只包含 path，例如 `/api/erp/register`
- 不包含 domain、protocol、query string
- `RAW_BODY` 必須是實際送出的原始 JSON 字串，不能重新排序欄位後再簽

## 3. 演算法

- 演算法：`HMAC-SHA256`
- secret：`INTEGRATION_SIGNATURE_SECRET`
- 輸出格式：小寫 hex 字串

## 4. 時間戳規則

- `X-TWACC-Timestamp` 使用 Unix 秒數
- 預設可容許誤差：`5 分鐘`
- 伺服器可用 `INTEGRATION_TIMESTAMP_TOLERANCE` 調整

若請求時間超過容許範圍，伺服器會回：

- `401 INTEGRATION_TIMESTAMP_EXPIRED`

## 5. Register 範例

Request body：

```json
{
  "source_system": "erp",
  "external_user_id": "A12345",
  "display_name": "王小明"
}
```

Header：

```http
Authorization: Bearer your-shared-token
X-TWACC-Timestamp: 1776384000
X-TWACC-Signature: <calculated-signature>
Content-Type: application/json
```

## 6. Login 範例

Request body：

```json
{
  "source_system": "erp",
  "external_user_id": "A12345",
  "device_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

Header：

```http
Authorization: Bearer your-shared-token
X-TWACC-Timestamp: 1776384000
X-TWACC-Signature: <calculated-signature>
Content-Type: application/json
```

## 7. Go 範例

```go
payload := []byte(`{"source_system":"erp","external_user_id":"A12345","display_name":"王小明"}`)
timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
message := strings.Join([]string{
	"POST",
	"/api/erp/register",
	timestamp,
	string(payload),
}, "\n")

mac := hmac.New(sha256.New, []byte(signatureSecret))
mac.Write([]byte(message))
signature := hex.EncodeToString(mac.Sum(nil))
```

## 8. JavaScript 範例

```js
const payload = JSON.stringify({
  source_system: "erp",
  external_user_id: "A12345",
  display_name: "王小明",
});

const timestamp = Math.floor(Date.now() / 1000).toString();
const message = ["POST", "/api/erp/register", timestamp, payload].join("\n");
```

前端或 Node.js 若要實際產生 HMAC-SHA256，請使用伺服器端安全環境保存 `signature secret`，不要把 secret 暴露在瀏覽器。

## 9. 常見錯誤碼

| Code | 說明 |
| --- | --- |
| `INVALID_INTEGRATION_TOKEN` | `Authorization` 錯誤或缺失 |
| `INTEGRATION_TIMESTAMP_REQUIRED` | 缺少 `X-TWACC-Timestamp` |
| `INVALID_INTEGRATION_TIMESTAMP` | timestamp 格式錯誤 |
| `INTEGRATION_TIMESTAMP_EXPIRED` | timestamp 超出容許時間範圍 |
| `INTEGRATION_SIGNATURE_REQUIRED` | 缺少 `X-TWACC-Signature` |
| `INVALID_INTEGRATION_SIGNATURE` | 簽章不正確 |

## 10. 開發模式

若伺服器未設定：

- `INTEGRATION_SHARED_TOKEN`
- `INTEGRATION_SIGNATURE_SECRET`

則 `register/login` 目前仍允許不帶整合驗證 header 直接測試。
