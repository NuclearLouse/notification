package bitrix

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"redits.oculeus.com/asorokin/request"
)

const (
	bitrixProtocol  = "https"
	requestMessage  = "im.message.add.json"
	requestDelete   = "im.message.delete"
	requestNotify   = "im.notify.system.add.json"
	reqValueMessId  = "MESSAGE_ID"
	reqValueDialog  = "DIALOG_ID"
	reqValueUserId  = "USER_ID"
	reqValueMessage = "MESSAGE"
	reqValueSystem  = "SYSTEM"
	// reqValueAttach  = "ATTACH"
)

type notificator struct {
	cfg *Config
}

type Config struct {
	Proto                 string
	Host                  string
	Token                 string
	UserId                string
	GlobalChat            string
	Timeout               int
	LifetimeMessage       int
	Addresses             []string
	GlobalChatUsers       []string
	UseGlobalNotification bool
}

func New(cfg *Config) *notificator {
	if cfg.Proto == "" {
		cfg.Proto = bitrixProtocol
	}
	return &notificator{cfg}
}

func (n *notificator) requestPath(request string) string {
	return fmt.Sprintf("/rest/%s/%s/%s", n.cfg.UserId, n.cfg.Token, request)
}

func (n *notificator) urlForMessage(dialogId, message string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestMessage),
			reqValueSystem, "Y",
			reqValueDialog, dialogId,
			reqValueMessage, message,
		).URL().String()
}

func (n *notificator) urlForDelete(messageId string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestDelete),
			reqValueMessId, messageId,
		).URL().String()
}

func (n *notificator) urlForNotify(userId, message string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestNotify),
			reqValueUserId, userId,
			reqValueMessage, message,
		).URL().String()

}

func (n *notificator) SendMessage(message *bytes.Buffer, subject string) error {

	if len(n.cfg.Addresses) == 0 {
		return errors.New("no addresses to send")
	}

	users := n.cfg.Addresses
	addresses := n.cfg.Addresses

	if n.cfg.UseGlobalNotification {
		addresses = []string{n.cfg.GlobalChat}
		users = n.cfg.GlobalChatUsers
	}

	for _, user := range users {
		url := n.urlForNotify(user, subject)
		if _, err := n.send(url); err != nil {
			return err
		}
	}

	for _, chat := range addresses {
		url := n.urlForMessage(chat, message.String())
		mesId, err := n.send(url)
		if err != nil {
			return err
		}
		go func() {
			time.Sleep(time.Duration(n.cfg.LifetimeMessage) * time.Hour)
			n.send(n.urlForDelete(fmt.Sprintf("%d", mesId)))
		}()
	}

	return nil
}

func (n *notificator) send(url string) (int64, error) {

	res, err := request.Do(&request.Params{
		URL: url,
		Client: &http.Client{
			Timeout: time.Duration(n.cfg.Timeout) * time.Second,
		},
	})
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	var result struct {
		Result int64 `json:"result"`
		Time   struct {
			Start      float64   `json:"start"`
			Finish     float64   `json:"finish"`
			Duration   float64   `json:"duration"`
			Processing float64   `json:"processing"`
			DateStart  time.Time `json:"date_start"`
			DateFinish time.Time `json:"date_finish"`
		} `json:"time"`
	}

	if res.StatusCode == 200 {
		//Success:
		// при удалении сообщения "result": 868197, type int64
		// при удалении сообщения "result": true, type bool - не обрабатываю результат
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			return result.Result, fmt.Errorf("decode result json: %w", err)
		}

		if result.Result == 0 {
			return result.Result, errors.New("no result in the response")
		}

		return result.Result, nil
	}
	//Error:
	var resultErr struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resultErr); err != nil {
		return result.Result, fmt.Errorf("decode result json: %w", err)
	}
	if resultErr.Error != "" {
		return result.Result, fmt.Errorf("%s: %s", resultErr.Error, resultErr.ErrorDescription)
	}
	return 0, errors.New("unsupported bitrix-api response")
}
