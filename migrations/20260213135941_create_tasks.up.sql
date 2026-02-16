CREATE TABLE tasks IF NOT EXISTS (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    "description" TEXT
    "type" VARCHAR(50) NOT NULL,
    payload JSONB,
    cron_expr VARCHAR(50),
    "status" VARCHAR(20) NOT NULL DEFAULT "pending",
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    next_run_at TIMESTAMP WITH TIME ZONE 
)

CREATE TABLE task_runs IN NOT EXISTS (
    id UUID PRIMARY KEY,
    task_id INT NOT NULL,
    "status" VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    finished_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    error_desc TEXT 
) 

CREATE INDEX idx_tasks_next_run_at on tasks(next_run_at) WHERE status = "scheduled"