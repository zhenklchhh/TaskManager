DROP INDEX IF EXISTS idx_tasks_priority_next_run;

ALTER TABLE tasks DROP COLUMN IF EXISTS priority;
