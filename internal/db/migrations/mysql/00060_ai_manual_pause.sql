-- +goose Up
-- 人工接管：账号级暂停时长与按会话持久化的暂停窗口。
ALTER TABLE ai_reply_settings ADD COLUMN manual_pause_minutes INT NOT NULL DEFAULT 10;
ALTER TABLE ai_reply_settings ADD COLUMN manual_pause_ai_only TINYINT(1) NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS ai_manual_pauses (
    cookie_id VARCHAR(255) NOT NULL,
    chat_id VARCHAR(255) NOT NULL,
    until_at BIGINT NOT NULL DEFAULT 0,
    ai_only TINYINT(1) NOT NULL DEFAULT 1,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (cookie_id, chat_id),
    INDEX idx_ai_manual_pauses_until (until_at),
    CONSTRAINT fk_ai_manual_pauses_cookie FOREIGN KEY (cookie_id) REFERENCES cookies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS ai_manual_pauses;
ALTER TABLE ai_reply_settings DROP COLUMN manual_pause_ai_only;
ALTER TABLE ai_reply_settings DROP COLUMN manual_pause_minutes;
