# agent.md

## 專案目標

本專案要建立一個以 Go 為主體的網頁聊天系統 MVP，需包含：

- 一對一聊天
- 群組聊天
- ERP 自動註冊與自動登入
- 裝置識別與登入紀錄
- 使用者個人 IP 白名單
- 電腦版與手機版 Web 介面

## 開發規則

- 後端以 Go 為主，不改用其他後端語言重寫。
- 前端頁面與 API 都維持在同一個專案中。
- `system_admin` 角色不可參與聊天功能。
- 需求以 `docs/requirements.md` 為最高優先依據。
- 開發時優先延續現有檔案結構與文件，不任意重組專案。

## 工作順序

1. 先閱讀 `docs/requirements.md`
2. 再閱讀 `docs/architecture.md`
3. 再閱讀 `docs/api.md`
4. 再閱讀 `docs/data-model.md`
5. 實作時同步更新 `TODO.md`

## 驗證要求

- 每次修改後都要保持 `go build ./...` 可通過
- 補功能時優先維持可執行狀態
- 涉及業務邏輯時要逐步補上測試
