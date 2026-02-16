package domain

import "errors"

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidCron  = errors.New("invalid cron expression")
	ErrValidation   = errors.New("invalid parameters")
)
