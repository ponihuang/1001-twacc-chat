# 系統架構草案

## 1. 目標

本文件定義 MVP 階段的系統模組切分。聊天與群組通訊是產品主核心；外部系統接入僅負責提供統一 API 能力，讓其他網站或業務系統可以透過 API 與本對談系統溝通，不另針對 ERP 或任何單一來源系統建立專屬產品分類或專屬界面。後續可在同一套 Go 專案中同時承載：

- Web 前端
- 對外 API
- 帳號與權限邏輯
- 聊天與檔案流程

## 2. 模組切分

### 2.1 `main`

- 啟動應用程式
- 載入設定
- 啟動 HTTP server

### 2.2 `internal/app`

- 組裝 server
- 注入設定與相依元件
- 管理啟動流程

### 2.3 `internal/httpx`

- 設定路由
- 處理中介層
- 提供首頁、健康檢查與 API 掛載點

### 2.4 後續預定模組

- `internal/auth`：登入、憑證、角色判斷
- `internal/integration`：外部系統接入入口，處理對外系統呼叫聊天系統 API 的共通能力
- `internal/chat`：一對一聊天與訊息流
- `internal/group`：群組建立、成員管理、角色權限
- `internal/upload`：圖片與檔案上傳限制
- `internal/device`：裝置識別與信任裝置流程
- `internal/whitelist`：使用者 IP 白名單檢查
- `internal/systemadmin`：`system_admin` 管理功能
- `internal/store`：資料存取層
- `internal/i18n`：三語系字串管理

## 3. 請求流程

### 3.1 Web 使用者

1. 瀏覽器進入 Web 頁面。
2. 前端讀取或建立 `device_id`。
3. 使用者登入。
4. 後端驗證帳號、裝置資訊、IP 白名單。
5. 驗證成功後取得聊天清單與訊息內容。

### 3.2 外部系統接入

1. 外部系統呼叫註冊、登入或後續聊天相關 API。
2. 外部系統先帶 `Authorization: Bearer <shared-token>`。
3. 外部系統再帶 `X-TWACC-Timestamp` 與 `X-TWACC-Signature`。
4. 後端先驗證 shared token、timestamp 與 HMAC-SHA256 簽章。
5. 驗證通過後，後端依 `source_system` 與外部帳號建立或查找使用者。
6. 後端回傳成功或失敗結果與原因碼。

### 3.3 三種 Web 介面

1. 桌機版全畫面提供雙欄式工作區。
2. 手機版全畫面提供單欄式頁面切換。
3. 嵌入式小浮動視窗可放入其他系統頁面中，作為輕量聊天入口。
4. 三種介面共用同一套聊天 API、登入狀態與權限邏輯。
5. 目前建議路由為 `/chat`、`/m/chat`、`/embed/chat`。

## 4. 前端結構

MVP 先採傳統模板加靜態資源方式，後續可視需要再拆成 SPA。

- `web/templates`：HTML 模板
- `web/static`：CSS、JavaScript、圖片素材
- `web/templates/desktop`：桌機版全畫面頁面
- `web/templates/mobile`：手機版全畫面頁面
- `web/templates/embed`：嵌入式浮動視窗頁面
- 目前程式已先以 `desktop.html`、`mobile.html`、`embed.html` 實作骨架模板
- 桌機版設定頁中的頭像與聊天室分類目前仍屬前端原型；聊天室分類資料暫存於瀏覽器 `localStorage`，正式版需改由後端設定 API 與資料表承接

## 5. 資料層方向

目前先以 MySQL 作為主要資料庫，後續若有需要再評估其他資料庫：

- MySQL（MVP 預設）
- PostgreSQL（後續可評估）
- SQLite（僅限極簡開發驗證，不列為正式方向）

## 6. 測試方向

- handler 單元測試
- service 單元測試
- API 整合測試
- 外部系統簽章 / timestamp 驗證測試
- 上傳限制測試
- 白名單與權限測試
- 桌機 / 手機 / 嵌入式三種前端形態驗證
