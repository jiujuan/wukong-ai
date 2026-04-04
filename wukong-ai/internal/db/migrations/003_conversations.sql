-- Wukong-AI v0.9 Migration: 多轮对话
BEGIN;

-- 对话会话表
CREATE TABLE IF NOT EXISTS conversations (
    id           VARCHAR(64)  PRIMARY KEY,
    title        VARCHAR(255) NOT NULL DEFAULT '',       -- 对话标题（首轮输入自动截取）
    summary      TEXT,                                    -- 滚动压缩的早期历史摘要
    turn_count   INTEGER      NOT NULL DEFAULT 0,
    create_time  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversations_update ON conversations (update_time DESC);

-- 对话轮次表
CREATE TABLE IF NOT EXISTS conversation_turns (
    id              BIGSERIAL    PRIMARY KEY,
    conversation_id VARCHAR(64)  NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    task_id         VARCHAR(64)  REFERENCES tasks(id) ON DELETE SET NULL,
    turn_index      INTEGER      NOT NULL,       -- 第几轮（从 0 开始，单调递增）
    role            VARCHAR(20)  NOT NULL,        -- user / assistant
    content         TEXT         NOT NULL,        -- 用户输入 或 输出摘要（≤ 200 字）
    full_output     TEXT,                         -- 完整输出（仅 assistant 轮）
    create_time     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    UNIQUE (conversation_id, turn_index)
);

CREATE INDEX IF NOT EXISTS idx_turns_conversation ON conversation_turns (conversation_id, turn_index ASC);

COMMIT;
