package notification

import (
	"io"
)

type Notificator interface {
	SendMessage(message io.Reader, subject string) error
}
