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

## 2. `system_admin_users`

欄位草案：

- `id`
- `user_id`
- `role`
- `created_by`
- `created_at`

說明：

- `system_admin` 角色與聊天使用者角色需明確區分

## 3. `auth_sessions`

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
- `user_id` 對應登入後的 actor
- `device_id` 對應登入裝置，用於後續 trusted device 判斷

## 4. `devices`

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

## 5. `ip_whitelists`

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
- `created_at`

說明：

- `message_type` 可為 `text`、`image`、`file`

## 9. `attachments`

欄位草案：

- `id`
- `message_id`
- `original_name`
- `storage_path`
- `mime_type`
- `size_bytes`
- `created_at`

## 10. 留存策略

MVP 需要支援訊息與檔案保留 3 個月，因此後續應補：

- 清理排程
- 封存策略
- 稽核紀錄
