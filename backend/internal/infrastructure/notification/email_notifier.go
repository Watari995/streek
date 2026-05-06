package notification

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type EmailNotifier struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailNotifier(host string, port string, username string, password string, from string) *EmailNotifier {
	return &EmailNotifier{host: host, port: port, username: username, password: password, from: from}
}

func (n *EmailNotifier) Notify(ctx context.Context, to valueobject.Email, subject string, body string) error {
	auth := smtp.PlainAuth("", n.username, n.password, n.host)
	addr := fmt.Sprintf("%s:%s", n.host, n.port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", n.from, to, subject, body)
	return smtp.SendMail(addr, auth, n.from, []string{to.String()}, []byte(msg))
}
