package email

import (
	"errors"
	"io"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"time"

	"github.com/jordan-wright/email"
)

type notificator struct {
	cfg *Config
}

type Config struct {
	SmtpUser    string
	SmtpPass    string
	SmtpHost    string
	SmtpPort    string
	VisibleName string
	Timeout     int
	Addresses   []string
	WithoutAuth bool
	// TemplHTML   bool
}

func New(cfg *Config) *notificator {
	return &notificator{cfg}
}

type userinfo struct {
	username string
	password string
}

func loginAuth(username, password string) smtp.Auth {
	return &userinfo{username, password}
}

func (u *userinfo) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(u.username), nil
}

func (u *userinfo) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(u.username), nil
		case "Password:":
			return []byte(u.password), nil
		default:
			return nil, errors.New("unknown for server")
		}
	}
	return nil, nil
}

func (n *notificator) SendMessage(message io.Reader, subject string) error {
	if len(n.cfg.Addresses) == 0 {
		return errors.New("no addresses to send")
	}

	from := mail.Address{
		Address: n.cfg.SmtpUser,
		Name:    n.cfg.VisibleName,
	}
	body, err := io.ReadAll(message)
	if err != nil {
		return err
	}
	m := &email.Email{
		From:    from.String(),
		To:      n.cfg.Addresses,
		Subject: subject,
		HTML:    body,
		Headers: textproto.MIMEHeader{},
	}

	var auth smtp.Auth = nil
	if !n.cfg.WithoutAuth {
		auth = loginAuth(n.cfg.SmtpUser, n.cfg.SmtpPass)
	}

	pool, err := email.NewPool(n.cfg.SmtpHost+":"+n.cfg.SmtpPort, 1, auth)
	if err != nil {
		return err
	}
	defer pool.Close()

	return pool.Send(m, time.Duration(n.cfg.Timeout)*time.Second)
}
