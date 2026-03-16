package domain

type TaskStats struct {
	TotalTasks      int            `json:"total_tasks"`
	PendingTasks    int            `json:"pending_tasks"`
	ScheduledTasks  int            `json:"scheduled_tasks"`
	RunningTasks    int            `json:"running_tasks"`
	CompletedTasks  int            `json:"completed_tasks"`
	FailedTasks     int            `json:"failed_tasks"`
	TasksByType     map[string]int `json:"tasks_by_type"`
	TasksByPriority map[int]int    `json:"tasks_by_priority"`
}
