# 裝置辨識建議流程

## 目標

在 MVP 階段先用成本最低、邏輯清楚的方式完成 `device_id` 與信任裝置流程。

## 建議流程

1. 前端首次進站即產生 UUID 格式的 `device_id`
2. 前端將 `device_id` 儲存在 `localStorage`
3. 外部系統登入 API 或 Web 登入流程一併帶上 `device_id`
4. 後端記錄 `device_id`、`user_agent`、`ip`、`first_login_at`、`last_login_at`
5. 若使用者目前沒有任何 trusted device，第一台登入成功裝置可直接標記 trusted
6. 若使用者已有 trusted device，新裝置先記錄為 untrusted
7. 未信任裝置登入時，回 `NEW_DEVICE_APPROVAL_REQUIRED`
8. 新裝置只能由既有 trusted device 或 `system_admin` 核准後改為 trusted
9. 使用者清除瀏覽器資料或更換瀏覽器時，視為新裝置

## 理由

- 前端直接產生 `device_id`，實作最簡單
- `localStorage` 足夠支撐目前 Web MVP
- trusted / untrusted 二分法容易落地
- 流程能和目前的外部系統登入 API 以及自有 Web 介面共用
