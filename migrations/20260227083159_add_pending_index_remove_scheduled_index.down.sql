DROP INDEX IF EXISTS idx_tasks_next_run_at;
CREATE INDEX idx_tasks_next_run_at on tasks(next_run_at) where status = 'scheduled';