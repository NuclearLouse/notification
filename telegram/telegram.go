package telegram

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"encoding/json"
	"errors"
	"fmt"

	"redits.oculeus.com/asorokin/request"
)

const (
	telegramProtocol = "https"
	requestMessage   = "sendMessage"
)

type notificator struct {
	cfg *Config
}

type Config struct {
	Proto     string
	Host      string
	Token     string
	Timeout   time.Duration
	Addresses []int
}

func New(cfg *Config) *notificator {
	if cfg.Proto == "" {
		cfg.Proto = telegramProtocol
	}
	return &notificator{cfg}
}

func (n *notificator) requestPath(request string) string {
	return fmt.Sprintf("/bot%s/%s", n.cfg.Token, request)
}

func (n *notificator) SendMessage(message io.Reader, subject string) error {
	body, err := io.ReadAll(message)
	if err != nil {
		return err
	}
	for _, chat := range n.cfg.Addresses {

		reqBody := struct {
			ChatId    int    `json:"chat_id"`
			Text      string `json:"text"`
			ParseMode string `json:"parse_mode"`
		}{
			ChatId:    chat,
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
				Timeout: time.Duration(n.cfg.Timeout) * time.Second,
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
