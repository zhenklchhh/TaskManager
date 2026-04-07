package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
	"gopkg.in/mail.v2"
)

type NotificationService struct {
	repo       repository.NotificationRepository
	dialer     *mail.Dialer
	httpClient *http.Client
}

func NewNotificationService(repo repository.NotificationRepository, dialer *mail.Dialer) *NotificationService {
	return &NotificationService{
		repo:   repo,
		dialer: dialer,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *NotificationService) CreateConfig(ctx context.Context, cfg *domain.NotificationConfig) error {
	cfg.ID = uuid.New()
	cfg.CreatedAt = time.Now()
	return s.repo.CreateConfig(ctx, cfg)
}

func (s *NotificationService) OnTaskStatusChanged(ctx context.Context, task *domain.Task, oldStatus, newStatus domain.TaskStatus) {
	var event domain.NotificationEvent
	switch newStatus {
	case domain.TaskStatusCompleted:
		event = domain.EventTaskCompleted
	case domain.TaskStatusFailed:
		event = domain.EventTaskFailed
	default:
		event = domain.EventStatusChanged
	}

	configs, err := s.repo.GetConfigsByTaskAndEvent(ctx, task.ID, event)
	if err != nil {
		slog.Error("notification: failed to get task configs", "task_id", task.ID, "error", err)
	}

	globalConfigs, err := s.repo.GetGlobalConfigsByEvent(ctx, event)
	if err != nil {
		slog.Error("notification: failed to get global configs", "error", err)
	}

	allConfigs := append(configs, globalConfigs...)

	for _, cfg := range allConfigs {
		log := &domain.NotificationLog{
			ID:          uuid.New(),
			ConfigID:    cfg.ID,
			TaskID:      task.ID,
			Event:       event,
			Status:      "pending",
			Attempts:    0,
			MaxAttempts: 3,
			CreatedAt:   time.Now(),
		}
		if err := s.repo.CreateLog(ctx, log); err != nil {
			slog.Error("notification: failed to create log", "config_id", cfg.ID, "error", err)
		}
	}
}

func (s *NotificationService) ProcessPendingNotifications(ctx context.Context) {
	logs, err := s.repo.GetPendingLogs(ctx, 50)
	if err != nil {
		slog.Error("notification: failed to get pending logs", "error", err)
		return
	}

	for _, log := range logs {
		cfg, err := s.repo.GetConfigByID(ctx, log.ConfigID)
		if err != nil {
			slog.Error("notification: failed to get config", "config_id", log.ConfigID, "error", err)
			continue
		}

		now := time.Now()
		log.Attempts++
		log.LastAttemptAt = &now

		var sendErr error
		switch cfg.Type {
		case domain.NotificationEmail:
			sendErr = s.sendEmail(cfg.Target, log)
		case domain.NotificationWebhook:
			sendErr = s.sendWebhook(cfg.Target, log)
		default:
			sendErr = fmt.Errorf("unknown notification type: %s", cfg.Type)
		}

		if sendErr != nil {
			log.LastError = sendErr.Error()
			if log.Attempts >= log.MaxAttempts {
				log.Status = "failed"
				slog.Error("notification: max attempts reached", "log_id", log.ID, "error", sendErr)
			} else {
				backoff := time.Duration(math.Pow(2, float64(log.Attempts))) * 30 * time.Second
				nextRetry := now.Add(backoff)
				log.NextRetryAt = &nextRetry
				slog.Warn("notification: retrying", "log_id", log.ID, "attempt", log.Attempts, "next_retry", nextRetry)
			}
		} else {
			log.Status = "sent"
			slog.Info("notification: sent successfully", "log_id", log.ID, "type", cfg.Type)
		}

		if err := s.repo.UpdateLog(ctx, log); err != nil {
			slog.Error("notification: failed to update log", "log_id", log.ID, "error", err)
		}
	}
}

func (s *NotificationService) sendEmail(target string, log *domain.NotificationLog) error {
	if s.dialer == nil {
		return fmt.Errorf("email dialer not configured")
	}

	msg := mail.NewMessage()
	msg.SetHeader("From", "taskmanager@notifications.local")
	msg.SetHeader("To", target)
	msg.SetHeader("Subject", fmt.Sprintf("Task %s: %s", log.Event, log.TaskID))
	msg.SetBody("text/plain", fmt.Sprintf(
		"Task ID: %s\nEvent: %s\nTime: %s",
		log.TaskID, log.Event, time.Now().Format(time.RFC3339),
	))

	return s.dialer.DialAndSend(msg)
}

func (s *NotificationService) sendWebhook(target string, log *domain.NotificationLog) error {
	payload := map[string]interface{}{
		"task_id":   log.TaskID.String(),
		"event":     string(log.Event),
		"timestamp": time.Now().Format(time.RFC3339),
		"attempt":   log.Attempts,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TaskManager-Event", string(log.Event))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook: unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
