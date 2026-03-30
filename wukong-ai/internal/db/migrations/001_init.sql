-- Wukong-AI 数据库初始化
-- 版本: 001_init.sql

BEGIN;

-- 启用 pgvector 扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 1. 任务主表
CREATE TABLE IF NOT EXISTS tasks (
    id               VARCHAR(64)  PRIMARY KEY,
    status           VARCHAR(20)  NOT NULL DEFAULT 'pending',
    mode             VARCHAR(20)  NOT NULL DEFAULT 'flash',
    user_input       TEXT         NOT NULL,
    intention        TEXT,
    plan             TEXT,
    tasks_list       JSONB,
    sub_results      JSONB,
    final_output     TEXT,
    last_node        VARCHAR(100),
    retry_count      INTEGER      NOT NULL DEFAULT 0,
    error_msg        TEXT,
    thinking_enabled BOOLEAN      NOT NULL DEFAULT FALSE,
    plan_enabled     BOOLEAN      NOT NULL DEFAULT FALSE,
    subagent_enabled BOOLEAN      NOT NULL DEFAULT FALSE,
    create_time      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    finish_time      TIMESTAMPTZ
);

-- 2. 节点执行日志
CREATE TABLE IF NOT EXISTS node_execution_logs (
    id          BIGSERIAL    PRIMARY KEY,
    task_id     VARCHAR(64)  NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    node_name   VARCHAR(100) NOT NULL,
    status      VARCHAR(20)  NOT NULL,
    input       TEXT,
    output      TEXT,
    error_msg   TEXT,
    duration_ms INTEGER,
    start_time  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    end_time    TIMESTAMPTZ
);

-- 3. 长期记忆
CREATE TABLE IF NOT EXISTS memories (
    id          BIGSERIAL    PRIMARY KEY,
    task_id     VARCHAR(64)  REFERENCES tasks(id) ON DELETE SET NULL,
    session_id  VARCHAR(64),
    content     TEXT         NOT NULL,
    embedding   vector(1536),
    memory_type VARCHAR(20)  NOT NULL DEFAULT 'long_term',
    metadata    JSONB,
    create_time TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 4. 工具调用日志
CREATE TABLE IF NOT EXISTS tool_call_logs (
    id          BIGSERIAL    PRIMARY KEY,
    task_id     VARCHAR(64)  NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    node_name   VARCHAR(100) NOT NULL,
    tool_name   VARCHAR(100) NOT NULL,
    input       TEXT,
    output      TEXT,
    success     BOOLEAN      NOT NULL DEFAULT TRUE,
    error_msg   TEXT,
    duration_ms INTEGER,
    call_time   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 5. 子 Agent 记录
CREATE TABLE IF NOT EXISTS sub_agents (
    id          BIGSERIAL    PRIMARY KEY,
    task_id     VARCHAR(64)  NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    agent_index INTEGER      NOT NULL,
    sub_task    TEXT         NOT NULL,
    result      TEXT,
    status      VARCHAR(20)  NOT NULL DEFAULT 'pending',
    retry_count INTEGER      NOT NULL DEFAULT 0,
    error_msg   TEXT,
    start_time  TIMESTAMPTZ,
    end_time    TIMESTAMPTZ
);

-- 6. 技能注册
CREATE TABLE IF NOT EXISTS skills (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    skill_type  VARCHAR(20)  NOT NULL DEFAULT 'basic',
    config      JSONB,
    enabled     BOOLEAN      NOT NULL DEFAULT TRUE,
    create_time TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_tasks_status      ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_create_time ON tasks (create_time DESC);
CREATE INDEX IF NOT EXISTS idx_node_logs_task_id ON node_execution_logs (task_id);
CREATE INDEX IF NOT EXISTS idx_memories_task_id  ON memories (task_id);
CREATE INDEX IF NOT EXISTS idx_memories_embedding
    ON memories USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
CREATE INDEX IF NOT EXISTS idx_tool_logs_task_id ON tool_call_logs (task_id);
CREATE INDEX IF NOT EXISTS idx_sub_agents_task_id ON sub_agents (task_id);

COMMIT;
