ALTER TABLE tasks
DROP COLUMN retry_count,
DROP COLUMN max_retries,
DROP COLUMN last_error_message;