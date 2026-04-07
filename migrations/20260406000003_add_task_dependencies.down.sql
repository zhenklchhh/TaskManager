DROP TABLE IF EXISTS task_dependencies;
ALTER TABLE tasks DROP COLUMN IF EXISTS parent_id;
