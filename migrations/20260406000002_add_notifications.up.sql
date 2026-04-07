CREATE TABLE IF NOT EXISTS notification_configs (
    id UUID PRIMARY KEY,
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    event VARCHAR(30) NOT NULL,
    target TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notification_logs (
    id UUID PRIMARY KEY,
    config_id UUID NOT NULL REFERENCES notification_configs(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    event VARCHAR(30) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    last_error TEXT,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_attempt_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_notification_configs_task_id ON notification_configs(task_id);
CREATE INDEX idx_notification_configs_event ON notification_configs(event);
CREATE INDEX idx_notification_logs_status ON notification_logs(status) WHERE status = 'pending';
CREATE INDEX idx_notification_logs_next_retry ON notification_logs(next_retry_at) WHERE status = 'pending';
