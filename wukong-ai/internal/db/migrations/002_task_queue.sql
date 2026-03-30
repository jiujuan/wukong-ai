-- Wukong-AI 任务队列表
-- 版本: 002_task_queue.sql

BEGIN;

-- 任务队列表
CREATE TABLE IF NOT EXISTS task_queue (
    id             BIGSERIAL    PRIMARY KEY,
    task_id        VARCHAR(64)  NOT NULL UNIQUE REFERENCES tasks(id) ON DELETE CASCADE,
    status         VARCHAR(20)  NOT NULL DEFAULT 'queued',
    priority       INTEGER      NOT NULL DEFAULT 0,
    payload        JSONB        NOT NULL,
    retry_count    INTEGER      NOT NULL DEFAULT 0,
    max_retries    INTEGER      NOT NULL DEFAULT 3,
    worker_id      VARCHAR(64),
    enqueue_time   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    start_time     TIMESTAMPTZ,
    finish_time    TIMESTAMPTZ,
    next_retry_at  TIMESTAMPTZ
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_task_queue_status
    ON task_queue (status, priority DESC, enqueue_time ASC);
CREATE INDEX IF NOT EXISTS idx_task_queue_retry
    ON task_queue (status, next_retry_at)
    WHERE status = 'queued';

COMMIT;
