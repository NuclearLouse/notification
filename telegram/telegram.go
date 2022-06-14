package telegram

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"encoding/json"
	"errors"
	"fmt"

	"redits.oculeus.com/asorokin/notification"
	"redits.oculeus.com/asorokin/request"
)

const (
	telegramProtocol = "https"
	requestMessage   = "sendMessage"
)

type Notificator struct {
	cfg *Config
}

func (n *Notificator) String() string {
	return "telegram"
}

type Config struct {
	Proto   string        `cfg:"proto"`
	Host    string        `cfg:"host"`
	Token   string        `cfg:"token"`
	Timeout time.Duration `cfg:"timeout"`
	// Addresses []int         `cfg:"addresses"`
}

func New(cfg *Config) *Notificator {
	if cfg.Proto == "" {
		cfg.Proto = telegramProtocol
	}
	return &Notificator{cfg}
}

func (n *Notificator) requestPath(request string) string {
	return fmt.Sprintf("/bot%s/%s", n.cfg.Token, request)
}

func (n *Notificator) SendMessage(message notification.Message, attachments ...notification.Attachment) error {

	//TODO: implement attacments for telegram

	body, err := io.ReadAll(message.Content)
	if err != nil {
		return err
	}
	for _, chat := range message.Addresses {
		chatid, err := strconv.Atoi(chat)
		if err != nil {
			return err
		}
		reqBody := struct {
			ChatId    int    `json:"chat_id"`
			Text      string `json:"text"`
			ParseMode string `json:"parse_mode"`
		}{
			ChatId:    chatid,
			Text:      string(body),
			ParseMode: "html",
		}
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(reqBody); err != nil {
			return fmt.Errorf("encode body JSON: %w", err)
		}

		res, err := request.Do(&request.Params{
			URL: request.NewAddress(n.cfg.Proto, n.cfg.Host).
				SetEndpoint(n.requestPath(requestMessage)).URL().String(),
			Body: buf,
			Header: map[string]string{
				"Content-Type": "application/json",
			},
			Client: &http.Client{
				Timeout: n.cfg.Timeout,
			},
		})
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			var response struct {
				OK          bool   `json:"ok"`
				ErrorCode   int    `json:"error_code"`
				Description string `json:"description"`
			}
			if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
				return fmt.Errorf("decode json errResponse: %w", err)
			}
			if response.OK {
				return fmt.Errorf("error:%d: %s", response.ErrorCode, response.Description)
			}
			return errors.New("unsupported telegram-api response")
		}
	}
	return nil
}
