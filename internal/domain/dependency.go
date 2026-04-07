package domain

import (
	"time"

	"github.com/google/uuid"
)

type DependencyCondition string

const (
	ConditionCompleted DependencyCondition = "completed"
	ConditionFailed    DependencyCondition = "failed"
	ConditionAny       DependencyCondition = "any"
)

type TaskDependency struct {
	ID          uuid.UUID
	TaskID      uuid.UUID
	DependsOnID uuid.UUID
	Condition   DependencyCondition
	CreatedAt   time.Time
}
