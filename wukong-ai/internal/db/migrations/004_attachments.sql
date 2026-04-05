-- Wukong-AI v1.1 Migration: 文件上传与多模态处理
BEGIN;

-- 附件元信息表
CREATE TABLE IF NOT EXISTS task_attachments (
    id              BIGSERIAL     PRIMARY KEY,
    task_id         VARCHAR(64)   NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    file_name       VARCHAR(255)  NOT NULL,
    file_path       TEXT          NOT NULL,
    mime_type       VARCHAR(100)  NOT NULL,
    file_size       BIGINT        NOT NULL DEFAULT 0,
    extract_status  VARCHAR(20)   NOT NULL DEFAULT 'pending',
    is_image        BOOLEAN       NOT NULL DEFAULT FALSE,
    chunk_count     INTEGER       NOT NULL DEFAULT 0,
    error_msg       TEXT,
    upload_time     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    extract_time    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_attachments_task_id
    ON task_attachments (task_id);

CREATE INDEX IF NOT EXISTS idx_attachments_status
    ON task_attachments (extract_status)
    WHERE extract_status IN ('pending', 'extracting');

-- 扩展 memories 表，关联附件
ALTER TABLE memories
    ADD COLUMN IF NOT EXISTS attachment_id BIGINT
        REFERENCES task_attachments(id) ON DELETE CASCADE;

ALTER TABLE memories
    ADD COLUMN IF NOT EXISTS chunk_index INTEGER;

COMMIT;
