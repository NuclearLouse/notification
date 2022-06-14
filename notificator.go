package notification

import "io"

//go:generate mockgen -destination=mock/notificator_mock.go redits.oculeus.com/asorokin/notification Notificator
type Notificator interface {
	SendMessage(message Message, attachments ...Attachment) error
	String() string
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
