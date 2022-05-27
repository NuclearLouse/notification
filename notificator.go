package notification

import (
	"io"
)

//addresses []string, message io.Reader, subject string

type Notificator interface {
	SendMessage(message Message, attachments ...Attachment) error
}

type Message struct {
	Addresses []string
	Content   io.Reader
	Subject   string
}

type Attachment struct {
	Filename    string
	Content     io.Reader
	ContentType string
}
