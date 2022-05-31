package bitrix

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"redits.oculeus.com/asorokin/notification"
	"redits.oculeus.com/asorokin/request"
)

const (
	bitrixProtocol          = "https"
	requestMessage          = "im.message.add.json"
	requestDelete           = "im.message.delete"
	requestNotify           = "im.notify.system.add.json"
	requestUserList         = "im.chat.user.list"
	requestBotMessage       = "imbot.message.add.json"
	requestBotMessageDelete = "imbot.message.delete"
	requestBotUserList      = "imbot.chat.user.list.json"
	reqValueChatId          = "CHAT_ID"
	reqValueMessId          = "MESSAGE_ID"
	reqValueDialog          = "DIALOG_ID"
	reqValueUserId          = "USER_ID"
	reqValueMessage         = "MESSAGE"
	reqValueSystem          = "SYSTEM"
	reqValueBotId           = "BOT_ID"
	reqValueClientId        = "CLIENT_ID"
	reqValueComplete        = "COMPLETE"
	// reqValueAttach  = "ATTACH"
)

type notificator struct {
	cfg *Config
}

type Config struct {
	Proto           string        `cfg:"proto"`
	Host            string        `cfg:"host"`
	UserToken       string        `cfg:"user_token"`
	UserID          string        `cfg:"user_id"`
	BotID           string        `cfg:"bot_id"`
	BotCode         string        `cfg:"bot_code"`
	ClientID        string        `cfg:"client_id"`
	AdminID         string        `cfg:"admin_id"`
	AdminToken      string        `cfg:"admin_token"`
	Timeout         time.Duration `cfg:"timeout"`
	LifetimeMessage time.Duration `cfg:"lifetime_message"`
	UseNotification bool          `cfg:"use_notification"`
	// Addresses       []string      `cfg:"addresses"`
}

func New(cfg *Config) *notificator {
	if cfg.Proto == "" {
		cfg.Proto = bitrixProtocol
	}
	return &notificator{cfg}
}

func (n *notificator) requestPath(request string) string {
	return fmt.Sprintf("/rest/%s/%s/%s", n.cfg.UserID, n.cfg.UserToken, request)
}

func (n *notificator) requestPathAdmin(request string) string {
	return fmt.Sprintf("/rest/%s/%s/%s", n.cfg.AdminID, n.cfg.AdminToken, request)
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

func (n *notificator) urlForBotMessage(dialogId, message string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestBotMessage),
			// reqValueSystem, "Y",
			reqValueDialog, dialogId,
			reqValueMessage, message,
			reqValueBotId, n.cfg.BotID,
			reqValueClientId, n.cfg.ClientID,
		).URL().String()
}

func (n *notificator) urlForDeleteMessage(messageId string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestDelete),
			reqValueMessId, messageId,
		).URL().String()
}

func (n *notificator) urlForBotDeleteMessage(messageId string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestBotMessageDelete),
			reqValueMessId, messageId,
			reqValueBotId, n.cfg.BotID,
			reqValueClientId, n.cfg.ClientID,
			reqValueComplete, "Y",
		).URL().String()
}

func (n *notificator) urlForNotify(userId, message string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPathAdmin(requestNotify),
			reqValueUserId, userId,
			reqValueMessage, message,
		).URL().String()
}

func (n *notificator) urlForUserList(chatId string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestUserList),
			reqValueChatId, chatId,
		).URL().String()
}

func (n *notificator) urlForBotUserList(chatId string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestBotUserList),
			reqValueChatId, chatId,
			reqValueBotId, n.cfg.BotID,
			reqValueClientId, n.cfg.ClientID,
		).URL().String()
}

func (n *notificator) getUserListForNotificate(chats []string) []string {
	var users []string
	for _, chat := range chats {
		if strings.HasPrefix(chat, "chat") {
			ids := func() []int64 {
				res, err := request.Do(&request.Params{
					URL: n.urlForBotUserList(chat),
					Client: &http.Client{
						Timeout: n.cfg.Timeout,
					},
				})
				if err != nil {
					return nil
				}
				defer res.Body.Close()

				var result struct {
					Result []int64 `json:"result"`
					Time   struct {
						Start      float64   `json:"start"`
						Finish     float64   `json:"finish"`
						Duration   float64   `json:"duration"`
						Processing float64   `json:"processing"`
						DateStart  time.Time `json:"date_start"`
						DateFinish time.Time `json:"date_finish"`
					} `json:"time"`
				}
				if res.StatusCode != 200 {
					return nil
				}
				if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
					return nil
				}
				return result.Result
			}()

			if ids != nil {
				for _, id := range ids {
					users = append(users, fmt.Sprintf("%d", id))
				}
			}

		} else {
			users = append(users, chat)
		}
	}
	return users
}

func (n *notificator) SendMessage(message notification.Message, attachments ...notification.Attachment) error {

	if len(message.Addresses) == 0 {
		return errors.New("no addresses to send")
	}

	//TODO: implement attacments for Bitrix

	if n.cfg.UseNotification {
		for _, user := range n.getUserListForNotificate(message.Addresses) {
			url := n.urlForNotify(user, message.Subject)
			if _, err := n.send(url); err != nil {
				return err
			}
		}
	}

	body, err := io.ReadAll(message.Content)
	if err != nil {
		return err
	}
	for _, chat := range message.Addresses {
		url := n.urlForBotMessage(chat, string(body))
		mesId, err := n.send(url)
		if err != nil {
			return err
		}
		if n.cfg.LifetimeMessage > 0 {
			go func() {
				time.Sleep(n.cfg.LifetimeMessage)
				n.send(n.urlForBotDeleteMessage(fmt.Sprintf("%d", mesId)))
			}()
		}
	}

	return nil
}

func (n *notificator) send(url string) (int64, error) {

	res, err := request.Do(&request.Params{
		URL: url,
		Client: &http.Client{
			Timeout: n.cfg.Timeout,
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
