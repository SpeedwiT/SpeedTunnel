package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SpeedwiT/SpeedTunnel/internal/config"
	"github.com/SpeedwiT/SpeedTunnel/internal/utils"
)

type Bot struct {
	token   string
	adminID int64
	client  *http.Client
}

func New(token string, adminID int64) *Bot {
	return &Bot{token: token, adminID: adminID, client: &http.Client{Timeout: 10 * time.Second}}
}

func (b *Bot) api(method string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.token, method)
}

func (b *Bot) sendMessage(chatID int64, text string) error {
	v := url.Values{}
	v.Set("chat_id", fmt.Sprint(chatID))
	v.Set("text", text)
	v.Set("parse_mode", "HTML")
	_, err := b.client.PostForm(b.api("sendMessage"), v)
	return err
}

func (b *Bot) handleCommand(chatID int64, text string) {
	text = strings.TrimSpace(text)
	switch {
	case strings.HasPrefix(text, "/status"):
		cfg, _ := config.Load()
		if len(cfg.Tunnels) == 0 {
			b.sendMessage(chatID, "⭕️ هیچ تانلی یافت نشد.")
			return
		}
		var sb strings.Builder
		sb.WriteString("📊 <b>وضعیت تانل‌های Speed Tunnel</b>\n\n")
		for _, t := range cfg.Tunnels {
			status := "🔴 خاموش"
			if t.Enabled {
				status = "🟢 فعال"
			}
			ping, err := utils.TCPPing(fmt.Sprintf("127.0.0.1:%d", t.ListenPort), 2*time.Second)
			pingStr := "—"
			if err == nil {
				pingStr = ping.Round(time.Millisecond).String()
			}
			sb.WriteString(fmt.Sprintf("▫️ <b>%s</b> (%s)\n   ID: <code>%s</code>\n   Transport: %s | SNI: %s\n   %s:%d → %s:%d\n   وضعیت: %s | Ping: %s\n\n",
				t.Name, t.Role, t.ID, t.Transport, t.SNI, "0.0.0.0", t.ListenPort, t.RemoteAddr, t.RemotePort, status, pingStr))
		}
		b.sendMessage(chatID, sb.String())
	case strings.HasPrefix(text, "/start"), strings.HasPrefix(text, "/help"):
		b.sendMessage(chatID, "🚀 <b>Speed Tunnel Bot</b>\n\n دستورات:\n /status - وضعیت تانل‌ها\n /help - راهنما\n\n کانال: @Speedw_IT\n پشتیبانی: @SpeedwIT")
	default:
		b.sendMessage(chatID, "❓ دستور نامشخص. /help را بزنید.")
	}
}

func (b *Bot) Start() {
	if b.token == "" {
		log.Println("[bot] token empty, bot disabled")
		return
	}
	log.Println("[bot] starting...")
	var offset int64
	for {
		v := url.Values{}
		v.Set("timeout", "30")
		v.Set("offset", fmt.Sprint(offset+1))
		resp, err := b.client.Get(b.api("getUpdates") + "?" + v.Encode())
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		var result struct {
			OK     bool `json:"ok"`
			Result []struct {
				UpdateID int64 `json:"update_id"`
				Message  *struct {
					Chat struct {
						ID int64 `json:"id"`
					} `json:"chat"`
					Text string `json:"text"`
					From struct {
						ID int64 `json:"id"`
					} `json:"from"`
				} `json:"message"`
			} `json:"result"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		resp.Body.Close()
		for _, u := range result.Result {
			offset = u.UpdateID
			if u.Message == nil {
				continue
			}
			if b.adminID != 0 && u.Message.From.ID != b.adminID && u.Message.Chat.ID != b.adminID {
				continue
			}
			b.handleCommand(u.Message.Chat.ID, u.Message.Text)
		}
	}
}

func Notify(token string, adminID int64, text string) error {
	if token == "" || adminID == 0 {
		return nil
	}
	b := New(token, adminID)
	return b.sendMessage(adminID, text)
}
