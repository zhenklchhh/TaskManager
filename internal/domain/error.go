package domain

import "errors"

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrInvalidCron        = errors.New("invalid cron expression")
	ErrValidation         = errors.New("invalid parameters")
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
	ErrBatchEmpty         = errors.New("batch is empty")
	ErrBatchTooLarge      = errors.New("batch exceeds maximum size")

	ErrUserNotFound      = errors.New("user not found")
	ErrEmailExists       = errors.New("email already registered")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrCompanyNotFound   = errors.New("company not found")
	ErrAlreadyInCompany  = errors.New("user already belongs to a company")
	ErrInviteNotFound    = errors.New("invite not found")
	ErrInviteExpired     = errors.New("invite has expired")
	ErrInviteUsed        = errors.New("invite already used")
	ErrGroupNotFound      = errors.New("project group not found")
	ErrOAuthAccountExists = errors.New("oauth account already linked")
	ErrTaskNotCancellable = errors.New("task cannot be cancelled in current state")
	ErrTaskDeleted        = errors.New("task has been deleted")
)
