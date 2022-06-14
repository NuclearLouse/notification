package email

import (
	"errors"
	"io"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/jordan-wright/email"

	"redits.oculeus.com/asorokin/notification"
)

type Notificator struct {
	cfg *Config
}

func (n *Notificator) String() string {
	return "email"
}

type Config struct {
	SmtpUser    string        `cfg:"smtp_user"`
	SmtpPass    string        `cfg:"smtp_pass"`
	SmtpHost    string        `cfg:"smtp_host"`
	SmtpPort    string        `cfg:"smtp_port"`
	VisibleName string        `cfg:"visible_name"`
	Timeout     time.Duration `cfg:"timeout"`
	WithoutAuth bool          `cfg:"without_auth"`
	// TemplHTML   bool
}

func New(cfg *Config) *Notificator {
	return &Notificator{cfg}
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

func (n *Notificator) SendMessage(message notification.Message, attachments ...notification.Attachment) error {
	if len(message.Addresses) == 0 {
		return errors.New("no addresses to send")
	}

	from := mail.Address{
		Address: n.cfg.SmtpUser,
		Name:    n.cfg.VisibleName,
	}
	body, err := io.ReadAll(message.Content)
	if err != nil {
		return err
	}
	m := &email.Email{
		From:    from.String(),
		To:      message.Addresses,
		Subject: message.Subject,
		HTML:    body,
	}
	if attachments != nil {
		for _, a := range attachments {
			if a.Content == nil {
				continue
			}
			if _, err := m.Attach(a.Content, a.Filename, a.ContentType); err != nil {
				return err
			}
		}
	}

	var auth smtp.Auth = nil
	if !n.cfg.WithoutAuth {
		auth = loginAuth(n.cfg.SmtpUser, n.cfg.SmtpPass)
	}

	var sendChannel chan error
	if n.cfg.Timeout > 0 {
		sendChannel = make(chan error, 1)
		go func() {
			sendChannel <- m.Send(n.cfg.SmtpHost+":"+n.cfg.SmtpPort, auth)
		}()
	} else {
		return m.Send(n.cfg.SmtpHost+":"+n.cfg.SmtpPort, auth)
	}
	select {
	case sendErr := <-sendChannel:
		return sendErr
	case <-time.After(n.cfg.Timeout):
		return errors.New("email sending timed out")
	}
}
