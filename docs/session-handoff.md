# Session Handoff

這份文件用來保存「目前工作記憶」。

建議用法：

- 每次對話上下文接近上限前更新一次
- 每完成一個明確子任務後更新一次
- 開新對話時，先讓助手讀這份文件，再讀 `README.md`、`TODO.md` 與相關 `docs/*.md`

## 使用原則

- 只保留高訊息密度內容
- 寫清楚下一步，不寫長篇回顧
- 以 repo 內已落地的事實為準
- 若已有 commit，請記下 commit SHA 或簡短說明

## Handoff Template

```md
## Current Goal
- 目前要完成的單一目標

## Done
- 這一輪已完成的事項

## Files Changed
- 修改過的檔案路徑

## Next Step
- 下一步最直接要做的事

## Verification
- 跑了哪些測試
- 結果如何

## Risks
- 已知風險
- 尚未處理的邊界情況

## Notes
- 其他必要背景
```

## Current Handoff

## Current Goal
- 讓桌機 / 手機 / embed 三種前端骨架開始串接真實聊天資料與送訊息流程

## Done
- 已明確定義產品主核心為聊天系統，外部整合為統一 API 能力
- 已建立桌機全畫面、手機全畫面、嵌入式浮動視窗三種頁面骨架
- 已完成 `GET /api/conversations`
- 已完成 `GET /api/conversations/{conversation_id}/messages`
- 已完成 `POST /api/conversations/{conversation_id}/messages`
- 已補 conversation membership 驗證與 sender 權限檢查
- 已補送訊息 handler / service / repository 單元測試

## Files Changed
- `README.md`
- `TODO.md`
- `docs/api.md`
- `docs/session-handoff.md`
- `internal/chat/types.go`
- `internal/chat/service.go`
- `internal/chat/handler.go`
- `internal/chat/service_test.go`
- `internal/chat/handler_test.go`
- `internal/httpx/router.go`
- `internal/store/mysql/chat_repository.go`

## Next Step
- 先讓桌機版骨架串 `GET /api/conversations`、`GET /api/conversations/{conversation_id}/messages`、`POST /api/conversations/{conversation_id}/messages`
- 抽出桌機 / 手機 / embed 可共用的前端聊天資料載入與送訊息腳本
- 規劃建立對話 API，補一對一 / 群組聊天建立流程

## Verification
- `go test ./...` passed

## Risks
- `/api/erp/*` 仍是暫時相容命名，後續應收斂為較中性的 integration 路由
- embed 模式尚未實作 token / session 邊界
- 目前只支援文字訊息；圖片 / 檔案訊息尚未實作
- 目前尚未支援建立對話，仍需依賴既有 conversation 資料才能送訊息

## Notes
- 若開新對話，先讀這份文件，再讀 `README.md`、`TODO.md`、`docs/api.md`
- 送訊息 API 目前只接受 `type = text`，且會 trim `content`
