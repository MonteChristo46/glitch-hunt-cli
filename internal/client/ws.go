package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type EventCallback func(eventType string, payload json.RawMessage)

type WSClient struct {
	apiEndpoint string
	apiKey      string
	eventTypes  []string
	conn        *websocket.Conn
	callback    EventCallback
}

func NewWSClient(apiEndpoint, apiKey string, eventTypes []string) *WSClient {
	return &WSClient{
		apiEndpoint: apiEndpoint,
		apiKey:      apiKey,
		eventTypes:  eventTypes,
	}
}

func (w *WSClient) OnEvent(cb EventCallback) {
	w.callback = cb
}

func (w *WSClient) Connect(ctx context.Context) error {
	u, err := url.Parse(w.apiEndpoint)
	if err != nil {
		return fmt.Errorf("parse endpoint: %w", err)
	}

	scheme := "ws"
	if u.Scheme == "https" {
		scheme = "wss"
	}

	wsURL := fmt.Sprintf("%s://%s/ws/cli?api_key=%s", scheme, u.Host, w.apiKey)
	for _, et := range w.eventTypes {
		wsURL += fmt.Sprintf("&event_type=%s", et)
	}

	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			fmt.Printf("\r⚠ Connection failed: %v (retrying in %v)...\n", err, backoff)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}
		backoff = 1 * time.Second
		w.conn = conn
		return nil
	}
}

func (w *WSClient) Listen(ctx context.Context) error {
	if w.conn == nil {
		return fmt.Errorf("not connected")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		w.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		_, message, err := w.conn.ReadMessage()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return fmt.Errorf("read message: %w", err)
		}

		var raw json.RawMessage
		if err := json.Unmarshal(message, &raw); err != nil {
			continue
		}

		var event struct {
			EventMetadata *struct {
				EventType string `json:"event_type"`
			} `json:"event_metadata"`
		}

		eventType := ""
		if err := json.Unmarshal(message, &event); err == nil && event.EventMetadata != nil {
			eventType = event.EventMetadata.EventType
		}

		if w.callback != nil {
			w.callback(eventType, raw)
		}
	}
}

func (w *WSClient) Close() error {
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}
