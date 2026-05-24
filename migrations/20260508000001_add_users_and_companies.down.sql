-- Remove indexes on tasks
DROP INDEX IF EXISTS idx_tasks_company_id;
DROP INDEX IF EXISTS idx_tasks_group_id;
DROP INDEX IF EXISTS idx_tasks_created_by;
DROP INDEX IF EXISTS idx_tasks_assigned_to;

-- Remove columns from tasks
ALTER TABLE tasks DROP COLUMN IF EXISTS assigned_to;
ALTER TABLE tasks DROP COLUMN IF EXISTS created_by;
ALTER TABLE tasks DROP COLUMN IF EXISTS group_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS company_id;

-- Drop foreign key from users to companies
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_company;

-- Drop tables in reverse order
DROP TABLE IF EXISTS invites;
DROP TABLE IF EXISTS project_groups;
DROP TABLE IF EXISTS company_members;
DROP TABLE IF EXISTS companies;

DROP INDEX IF EXISTS idx_users_oauth;
DROP TABLE IF EXISTS users;
