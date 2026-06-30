# 資料模型草案

本文件先整理 MVP 需要的主要資料實體，供後續建表與 Go struct 設計。

## 1. `users`

欄位草案：

- `id`
- `source_system`
- `external_user_id`
- `display_name`
- `password_hash`
- `email`
- `language`
- `whatsapp_account`
- `telegram_account`
- `status`
- `created_at`
- `updated_at`

說明：

- 聊天使用者主表
- `(source_system, external_user_id)` 需唯一
- `whatsapp_account`、`telegram_account` 為非強制欄位，僅供通知用途
- `whatsapp_account`、`telegram_account` 不作為登入識別

## 2. `admin_users`

欄位草案：

- `id`
- `account`
- `display_name`
- `password_hash`
- `role`
- `status`
- `last_login_at`
- `last_login_ip`
- `created_by`
- `created_at`
- `updated_at`

說明：

- 後台管理員主表，與聊天室玩家 `users` 分離
- `account` 需唯一
- `system_admin` 角色不可直接參與聊天

## 3. `admin_auth_sessions`

欄位草案：

- `id`
- `token_hash`
- `admin_user_id`
- `device_id`
- `expires_at`
- `created_at`
- `last_used_at`
- `revoked_at`

說明：

- 保存後台管理員 Bearer session token 的雜湊值
- `admin_user_id` 對應 `admin_users.id`

## 4. `auth_sessions`

欄位草案：

- `id`
- `token_hash`
- `user_id`
- `device_id`
- `expires_at`
- `created_at`
- `last_used_at`
- `revoked_at`

說明：

- 保存 Bearer session token 的雜湊值，不直接保存明文 token
- `user_id` 對應聊天室玩家登入後的 actor
- `device_id` 對應登入裝置，用於後續 trusted device 判斷

## 5. `devices`

欄位草案：

- `id`
- `user_id`
- `device_id`
- `user_agent`
- `ip`
- `first_login_at`
- `last_login_at`
- `is_trusted_device`
- `trusted_by`

## 6. `ip_whitelists`

欄位草案：

- `id`
- `user_id`
- `rule_type`
- `ip_or_cidr`
- `enabled`
- `created_at`
- `updated_at`

說明：

- `rule_type` 可區分 `ipv4`、`ipv6`、`cidr`

## 5.1 `user_security_settings`

欄位草案：

- `user_id`
- `allow_all_ips`
- `created_at`
- `updated_at`

說明：

- `allow_all_ips = true` 時，略過 IP 白名單比對

## 6. `conversations`

欄位草案：

- `id`
- `type`
- `name`
- `created_by`
- `created_at`

說明：

- `type` 可為 `direct` 或 `group`

## 7. `conversation_members`

欄位草案：

- `id`
- `conversation_id`
- `user_id`
- `role`
- `joined_at`

說明：

- 群組角色可為 `owner`、`admin`、`member`

## 8. `messages`

欄位草案：

- `id`
- `conversation_id`
- `sender_id`
- `message_type`
- `content`
- `is_recalled`
- `created_at`

說明：

- `message_type` 可為 `text`、`image`、`file`
- `is_recalled` 表示訊息已收回。收回屬於軟隱藏，訊息與附件資料仍保留，前端列表、訊息清單、搜尋與未讀計算需排除已收回訊息。

## 9. `attachments`

欄位草案：

- `id`
- `message_id`
- `original_name`
- `storage_path`
- `mime_type`
- `size_bytes`
- `created_at`

## 10. `conversation_reads`

欄位草案：

- `conversation_id`
- `user_id`
- `last_read_message_id`
- `last_read_at`

說明：

- 用於計算每個使用者在各對話的未讀數。
- 訊息列表 API 會依此表找出第一筆未讀訊息並回傳 `first_unread_message_id`，讓前端可定位到未讀起點。
- 前端會在使用者實際看到訊息後呼叫已讀 API 更新 `last_read_message_id`。

## 11. `message_mentions`

欄位草案：

- `id`
- `message_id`
- `conversation_id`
- `mentioned_user_id`
- `mention_type`
- `created_at`

說明：

- 用於記錄文字訊息中的 `@使用者` 與 `@ALL` 標註目標。
- `mention_type` 可為 `user` 或 `all`。
- 對話列表會搭配 `conversation_reads.last_read_message_id` 判斷目前使用者是否有未讀標註，並回傳 `has_unread_mention`。

## 12. `user_contacts`

欄位草案：

- `id`
- `owner_user_id`
- `contact_user_id`
- `alias_name`
- `status`
- `created_at`
- `updated_at`

說明：

- 用於 Telegram-like 的單向聯絡人關係。
- `owner_user_id` 是誰的聯絡人清單，`contact_user_id` 是被加入的人。
- `alias_name` 是目前使用者給聯絡人的顯示名稱 / 備註名。
- `status` 目前使用 `active`；移除或封鎖聯絡人時可擴充為 `deleted`、`blocked`。

## 12. 留存策略

MVP 需要支援訊息與檔案保留 3 個月，因此後續應補：

- 清理排程
- 封存策略
- 稽核紀錄
