-- +goose Up
-- 统一布尔列为整数：SQLite/MySQL 用整数存储布尔值，Postgres 需要对齐，驱动才能按整数扫描。
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'ai_reply_settings' AND column_name = 'manual_pause_ai_only' AND data_type = 'boolean') THEN
        ALTER TABLE ai_reply_settings ALTER COLUMN manual_pause_ai_only DROP DEFAULT;
        ALTER TABLE ai_reply_settings ALTER COLUMN manual_pause_ai_only TYPE INTEGER USING (CASE WHEN manual_pause_ai_only THEN 1 ELSE 0 END);
        ALTER TABLE ai_reply_settings ALTER COLUMN manual_pause_ai_only SET DEFAULT 1;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'ai_manual_pauses' AND column_name = 'ai_only' AND data_type = 'boolean') THEN
        ALTER TABLE ai_manual_pauses ALTER COLUMN ai_only DROP DEFAULT;
        ALTER TABLE ai_manual_pauses ALTER COLUMN ai_only TYPE INTEGER USING (CASE WHEN ai_only THEN 1 ELSE 0 END);
        ALTER TABLE ai_manual_pauses ALTER COLUMN ai_only SET DEFAULT 1;
    END IF;
END $$;

-- +goose Down
-- 保持整数存储，不做回退转换。
SELECT 1;
