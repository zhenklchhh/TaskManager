DROP INDEX IF EXISTS idx_tasks_type_status;
DROP INDEX IF EXISTS idx_tasks_status_priority;
DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_status_type_created;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_priority_next_run;
CREATE INDEX idx_tasks_priority_next_run ON tasks(priority DESC, next_run_at) WHERE status = 'pending';
