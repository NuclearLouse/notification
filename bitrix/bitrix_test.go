package bitrix

import (
	"testing"
)

func testNotificator() *notificator {
	return &notificator{
		cfg: &Config{
			Proto:     bitrixProtocol,
			Host:      "company-name.bitrix24.eu",
			Token:     "777token666",
			UserId:    "1234",
			Addresses: []string{"999"},
		},
	}
}
func Test_notificator_urlForMessage(t *testing.T) {
	type args struct {
		dialogId string
		message  string
	}
	tests := []struct {
		name string
		n    *notificator
		args args
		want string
	}{
		{
			name: "Эндпоинт для отправки сообщения",
			n:    testNotificator(),
			args: args{
				dialogId: "chat987",
				message:  "test phrase",
			},
			want: "https://company-name.bitrix24.eu/rest/1234/777token666/im.message.add.json?DIALOG_ID=chat987&MESSAGE=test+phrase&SYSTEM=Y",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.urlForMessage(tt.args.dialogId, tt.args.message); got != tt.want {
				t.Errorf("notificator.urlForMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_notificator_urlForDelete(t *testing.T) {
	type args struct {
		messageId string
	}
	tests := []struct {
		name string
		n    *notificator
		args args
		want string
	}{
		{
			name: "Эндпоинт для удаления сообщения",
			n:    testNotificator(),
			args: args{
				messageId: "987654321",
			},
			want: "https://company-name.bitrix24.eu/rest/1234/777token666/im.message.delete?MESSAGE_ID=987654321",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.urlForDelete(tt.args.messageId); got != tt.want {
				t.Errorf("notificator.urlForDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_notificator_urlForNotify(t *testing.T) {
	type args struct {
		userId  string
		message string
	}
	tests := []struct {
		name string
		n    *notificator
		args args
		want string
	}{
		{
			name: "Эндпоинт для уведомления юзера",
			n:    testNotificator(),
			args: args{
				userId:  "1234",
				message: "test phrase",
			},
			want: "https://company-name.bitrix24.eu/rest/1234/777token666/im.notify.system.add.json?MESSAGE=test+phrase&USER_ID=1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.urlForNotify(tt.args.userId, tt.args.message); got != tt.want {
				t.Errorf("notificator.urlForNotify() = %v, want %v", got, tt.want)
			}
		})
	}
}
