package task

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/zhenklchhh/TaskManager/internal/domain"
	"gopkg.in/mail.v2"
)

type TaskHandler interface {
	Handle(ctx context.Context, task *domain.Task) error
}

const (
	SendEmailTask = "send_email"
)

type EmailCredentials struct {
	To      string `json:"to"`
	From string `json:"from"`
	Body    string `json:"body"`
}

type EmailTaskHandler struct {
	diler          *mail.Dialer
}

func NewEmailTaskHandler(host string, port int, username, password string) *EmailTaskHandler {
	return &EmailTaskHandler{
		diler: mail.NewDialer(host, port, username, password),
	}
}

func (h *EmailTaskHandler) Handle(ctx context.Context, task *domain.Task) error {
	message := mail.NewMessage()
	
	var mailCreds EmailCredentials
	err := json.Unmarshal(task.Payload, &mailCreds)
	if err != nil {
		return errors.New("email handler: invalid json")
	}
	
	message.SetHeader("From", mailCreds.From)
	message.SetHeader("To", mailCreds.To)
	
	message.SetBody("text/plain", mailCreds.Body)
	return h.diler.DialAndSend(message)
}
