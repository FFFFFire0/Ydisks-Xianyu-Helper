-- +goose Up
-- 整数布尔列已满足应用读取要求，SQLite 无需结构变更；保留版本号与其他方言一致。
SELECT 1;

-- +goose Down
SELECT 1;
