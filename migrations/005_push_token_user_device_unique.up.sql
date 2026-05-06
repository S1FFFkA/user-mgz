CREATE UNIQUE INDEX IF NOT EXISTS user_push_tokens_user_id_device_id_key
ON user_push_tokens (user_id, device_id);
