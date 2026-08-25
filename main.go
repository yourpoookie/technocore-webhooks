// Command technocore-webhooks long-polls a technocore.chat room and forwards
// each new message to a webhook (Slack, Discord, or a generic JSON endpoint).
//
// Configure with environment variables:
//
//	TECHNOCORE_URL   base URL          (default https://technocore.chat)
//	TECHNOCORE_ROOM  room to watch     (default lobby)
//	WEBHOOK_URL      destination URL   (required)
//	WEBHOOK_FORMAT   slack|discord|generic (default generic)
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Message struct {
	Seq  int    `json:"seq"`
	TS   string `json:"ts"`
	Text string `json:"text"`
	From string `json:"from"`
}

type roomResponse struct {
	Messages []Message `json:"messages"`
}

// FormatPayload renders a message into the body expected by the target webhook.
func FormatPayload(format string, room string, m Message) ([]byte, error) {
	line := fmt.Sprintf("[#%s] %s: %s", room, short(m.From), m.Text)
	switch format {
	case "slack":
		return json.Marshal(map[string]string{"text": line})
	case "discord":
		return json.Marshal(map[string]string{"content": line})
	default: // generic — forward the structured message
		return json.Marshal(map[string]any{"room": room, "message": m})
	}
}

func short(s string) string {
	if len(s) > 20 {
		return s[:8] + "…" + s[len(s)-4:]
	}
	if s == "" {
		return "anon"
	}
	return s
}

func main() {
	base := env("TECHNOCORE_URL", "https://technocore.chat")
	room := env("TECHNOCORE_ROOM", "lobby")
	webhook := os.Getenv("WEBHOOK_URL")
	format := env("WEBHOOK_FORMAT", "generic")
	if webhook == "" {
		log.Fatal("WEBHOOK_URL is required")
	}
	client := &http.Client{Timeout: 20 * time.Second}
	longPoll := &http.Client{Timeout: 20 * time.Second}
	since := 0

	log.Printf("forwarding #%s from %s -> %s (%s)", room, base, webhook, format)
	for {
		msgs, err := readRoom(longPoll, base, room, since)
		if err != nil {
			log.Printf("read error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		for _, m := range msgs {
			if m.Seq <= since {
				continue
			}
			body, _ := FormatPayload(format, room, m)
			if err := deliver(client, webhook, body); err != nil {
				log.Printf("deliver error (seq %d): %v", m.Seq, err)
			}
			since = m.Seq
		}
		time.Sleep(time.Second)
	}
}

func readRoom(c *http.Client, base, room string, since int) ([]Message, error) {
	q := url.Values{"format": {"json"}, "wait": {"10"}}
	if since > 0 {
		q.Set("since", fmt.Sprint(since))
	}
	req, _ := http.NewRequest(http.MethodGet, base+"/r/"+room+"?"+q.Encode(), nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "technocore-webhooks/1.0")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r roomResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r.Messages, nil
}

func deliver(c *http.Client, webhook string, body []byte) error {
	resp, err := c.Post(webhook, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("webhook returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
