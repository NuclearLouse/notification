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
	requestUserInfo         = "im.user.get"
	reqValueChatId          = "CHAT_ID"
	reqValueMessId          = "MESSAGE_ID"
	reqValueDialog          = "DIALOG_ID"
	reqValueUserId          = "USER_ID"
	reqValueMessage         = "MESSAGE"
	reqValueSystem          = "SYSTEM"
	reqValueBotId           = "BOT_ID"
	reqValueClientId        = "CLIENT_ID"
	reqValueComplete        = "COMPLETE"
	reqValueID              = "ID"
	// reqValueAttach  = "ATTACH"
)

type Notificator struct {
	cfg *Config
}

func (n *Notificator) String() string {
	return "bitrix"
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

func New(cfg *Config) *Notificator {
	if cfg.Proto == "" {
		cfg.Proto = bitrixProtocol
	}
	return &Notificator{cfg}
}

func (n *Notificator) requestPath(request string) string {
	return fmt.Sprintf("/rest/%s/%s/%s", n.cfg.UserID, n.cfg.UserToken, request)
}

func (n *Notificator) requestPathAdmin(request string) string {
	return fmt.Sprintf("/rest/%s/%s/%s", n.cfg.AdminID, n.cfg.AdminToken, request)
}

/*

{"log": "[ERROR] 2022/06/05 06:11:41.675974 ReporterWorker0: prepare: COPY raterresult.\"CarrierSum202204\"(\"Datum\",\"Year\",\"Month\",\"RType\",\"Carrierid\",\"OutRouteSum\",\"InRouteSu
m\",\"RateID\",\"OutRouteSumOrg\",\"InRouteSumOrg\",\"WaehrungIn\",\"WaehrungOut\") FROM STDIN -> pq: column \"OutRouteSumOrg\" of relation \"CarrierSum202204\" does not exist", "date": 165
4402305.704499}
{"log": "[ERROR] 2022/06/05 06:11:53.291055 ReporterWorker1: rows: capratedb.export_at_report -> pq: relation \"mtcarrierdbret.LpDestCountry\" does not exist", "date": 1654402315.7006679}

{"log": "[ERROR] 2022-06-14 18:46:57.035532+02 pgAgent job \"pgagent_capturang_check_routing_cache_nightly\"[2] status: \"f\" last run: 2022-06-14 01:01:00.022999+02 msg: ERROR:  duplicate
 key value violates unique constraint \"tmp_over_carr_pkey\"\nDETAIL:  Key (destid, tz, lock_pos)=(1834, 1, 1) already exists.\nCONTEXT:  SQL statement \"INSERT INTO tmp_over_carr \r\n\t\tS
ELECT \"Destinationid\", \"Timezone\", \"RankPos\", \"OverPos\" \r\n\t\t\tFROM cache_db.\"cache_2_2022-06-14\" WHERE \"OverPos\" > 0\"\nPL/pgSQL function cache_db.mid_cache_over(integer,dat
e) line 42 at EXECUTE\nSQL statement \"SELECT cache_db.mid_cache_over(i_rtype, i_day)\"\nPL/pgSQL function cache_db.check_cached_rates_cost(integer,date) line 160 at PERFORM\nSQL statement
\"SELECT cache_db.check_cached_rates_cost(i_rtypeid, i_date)\"\nPL/pgSQL function cache_db.check_cached_rates(integer,date) line 4 at PERFORM\nSQL statement \"SELECT cache_db.check_cached_r
ates(rtype, reference_date + I)\"\nPL/pgSQL function cache_db.f__cache_auto_check(date,integer,boolean,boolean,integer) line 48 at PERFORM", "date": 1655225217.035532}

url := b.urlForMessage(chat, body.String())
		mesId, err := b.send(url)
		if err != nil {
			return err
		}

func (b *Bitrixer) urlForMessage(dialogId, message string) string {
	return request.MethodApiURL(&request.MethodApi{
		Host: b.RestUrl,
		Path: b.requestPath(requestMessage),
		Params: map[string]string{
			reqParamSystem:  "Y",
			reqParamDialog:  dialogId,
			reqParamMessage: message,
		},
	})
}

type MethodApi struct {
	Scheme string
	Host   string
	Path   string
	Params map[string]string
}

func MethodApiURL(m *MethodApi) string {
	val := url.Values{}
	for key, value := range m.Params {
		val.Set(key, value)
	}
	if m.Scheme == "" {
		m.Scheme = "https"
	}
	u := url.URL{
		Scheme:   m.Scheme,
		Host:     m.Host,
		Path:     m.Path,
		RawQuery: val.Encode(),
	}
	return u.String()
}
*/

func (n *Notificator) urlForMessage(dialogId, message string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestMessage),
			reqValueSystem, "Y",
			reqValueDialog, dialogId,
			reqValueMessage, message,
		)
}

func (n *Notificator) urlForBotMessage(dialogId, message string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestBotMessage),
			reqValueDialog, dialogId,
			reqValueMessage, message,
			reqValueBotId, n.cfg.BotID,
			reqValueClientId, n.cfg.ClientID,
		)
}

func (n *Notificator) urlForDeleteMessage(messageId string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestDelete),
			reqValueMessId, messageId,
		)
}

func (n *Notificator) urlForBotDeleteMessage(messageId string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestBotMessageDelete),
			reqValueMessId, messageId,
			reqValueBotId, n.cfg.BotID,
			reqValueClientId, n.cfg.ClientID,
			reqValueComplete, "Y",
		)
}

func (n *Notificator) urlForNotify(userId, message string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPathAdmin(requestNotify),
			reqValueUserId, userId,
			reqValueMessage, message,
		)
}

func (n *Notificator) urlForUserList(chatId string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestUserList),
			reqValueChatId, chatId,
		)
}

func (n *Notificator) urlForBotUserList(chatId string) string {
	return request.NewAddress(n.cfg.Proto, n.cfg.Host).
		SetEndpoint(
			n.requestPath(requestBotUserList),
			reqValueDialog, chatId,
			reqValueBotId, n.cfg.BotID,
			reqValueClientId, n.cfg.ClientID,
		)
}

func (n *Notificator) getUserListForNotificate(chats []string) []string {
	//TODO: переделать под универсальный ответ
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

func (n *Notificator) SendMessage(message notification.Message, attachments ...notification.Attachment) error {

	if len(message.Addresses) == 0 {
		return errors.New("no addresses to send")
	}

	//TODO: implement attacments for Bitrix

	if n.cfg.UseNotification {
		for _, u := range n.getUserListForNotificate(message.Addresses) {
			//TODO: в рутинах и без чтения ошибок
			user := u
			go func() {
				url := n.urlForNotify(user, message.Subject)
				n.send(url)
			}()
			// if _, err := n.send(url); err != nil {
			// 	return err
			// }
		}
	}

	body, err := io.ReadAll(message.Content)
	if err != nil {
		return err
	}
	for _, chat := range message.Addresses {
		url := n.urlForBotMessage(chat, string(body))
		res, err := n.send(url)
		if err != nil {
			return err // error by DoRequest or decode response json
		}

		if res.StatusCode != 200 {
			return fmt.Errorf("status: %s error: %s : %s", res.Status, res.Error, res.Description)
		}

		switch result := res.Result.(type) {
		case int, float64:
			if n.cfg.LifetimeMessage > 0 {
				go func() {
					time.Sleep(n.cfg.LifetimeMessage)
					n.send(n.urlForBotDeleteMessage(fmt.Sprintf("%v", result)))
				}()
			}
			// case bool:
			//пока не надо никак обрабатывать
		}

	}

	return nil
}

type (
	response struct {
		Status     string
		StatusCode int
		Result     interface{} `json:"result"`
		Time       struct {
			Start      float64   `json:"start"`
			Finish     float64   `json:"finish"`
			Duration   float64   `json:"duration"`
			Processing float64   `json:"processing"`
			DateStart  time.Time `json:"date_start"`
			DateFinish time.Time `json:"date_finish"`
		} `json:"time"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
)

func (n *Notificator) send(url string) (*response, error) {

	res, err := request.Do(&request.Params{
		URL: url,
		Client: &http.Client{
			Timeout: n.cfg.Timeout,
		},
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	result := &response{
		StatusCode: res.StatusCode,
		Status:     res.Status,
	}

	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		return result, fmt.Errorf("decode response json: %w", err)
	}

	return result, nil
}
