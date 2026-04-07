-- Composite index for dashboard filtering by type + status
CREATE INDEX IF NOT EXISTS idx_tasks_type_status ON tasks(type, status);

-- Composite index for filtering by priority range + status
CREATE INDEX IF NOT EXISTS idx_tasks_status_priority ON tasks(status, priority);

-- Composite index for filtering by creation date
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC);

-- Composite index for combined filters: status + type + created_at
CREATE INDEX IF NOT EXISTS idx_tasks_status_type_created ON tasks(status, type, created_at DESC);

-- Composite index for batch operations on status
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

-- Index for task dependencies (parent_id) - will be used after adding dependencies
-- Optimize the existing pending tasks query with priority
DROP INDEX IF EXISTS idx_tasks_priority_next_run;
CREATE INDEX idx_tasks_priority_next_run ON tasks(priority ASC, next_run_at ASC) WHERE status = 'pending';
