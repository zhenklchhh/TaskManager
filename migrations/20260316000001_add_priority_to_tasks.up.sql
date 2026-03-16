ALTER TABLE tasks ADD COLUMN priority INTEGER DEFAULT 5 NOT NULL;

CREATE INDEX idx_tasks_priority_next_run ON tasks(priority DESC, next_run_at) WHERE status = 'pending';

COMMENT ON COLUMN tasks.priority IS 'Task priority: 1 (highest) to 10 (lowest), default 5';
