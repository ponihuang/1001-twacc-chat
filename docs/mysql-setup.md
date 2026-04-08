# MySQL 啟動草案

## 環境變數

- `ENABLE_MYSQL`：是否在啟動時初始化 MySQL，預設 `false`
- `AUTO_MIGRATE`：是否在啟動時自動套用 migration，預設 `true`
- `MIGRATIONS_DIR`：migration 目錄，預設 `db/migrations`
- `MYSQL_DSN`：完整 DSN，若提供則優先使用
- `MYSQL_HOST`：預設 `127.0.0.1`
- `MYSQL_PORT`：預設 `3306`
- `MYSQL_DATABASE`
- `MYSQL_USER`
- `MYSQL_PASSWORD`
- `MYSQL_PARAMS`：預設 `parseTime=true&charset=utf8mb4&loc=Local`
- `MYSQL_MAX_OPEN_CONNS`：預設 `10`
- `MYSQL_MAX_IDLE_CONNS`：預設 `5`
- `MYSQL_CONN_MAX_LIFETIME`：預設 `30m`

## 初始化步驟

1. 建立 MySQL database，例如 `twacc_chat`
2. 設定環境變數後啟動服務
3. 若 `AUTO_MIGRATE=true`，服務啟動時會自動套用 `MIGRATIONS_DIR` 內尚未執行的 `.sql` 檔案

## 目前範圍

- 已建立 Go 側設定讀取骨架
- 已串接 MySQL driver dependency
- 已建立 migration 自動執行流程
- 已建立 `schema_migrations` 紀錄表
- 已建立使用者、裝置、IP 白名單與 `auth_sessions` 資料表初始化 SQL

## 注意事項

- migration runner 目前以檔名排序執行 `.sql` 檔案
- migration 內容需以分號結尾
- 目前 parser 支援一般多行 DDL / DML，不適合包含複雜 stored procedure delimiter 的 SQL

## 建議下一步

1. 補 ERP `Authorization` 以外的簽章 / timestamp 驗證
2. 規劃 session 續期、登出與失效策略
3. 補聊天、群組與上傳 API
