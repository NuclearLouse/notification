package notification

import "bytes"

type Notificator interface {
	SendMessage(message *bytes.Buffer, subject string) error
}
