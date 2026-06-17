package ws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1 << 20 // 1 MiB (audio chunks)
	sendBuffer     = 256
)

// Client is one WebSocket connection in a room.
type Client struct {
	id   string
	name string
	conn *websocket.Conn
	send chan []byte
	room *Room
	hub  *Hub
}

// newClient builds a client bound to a connection.
func newClient(id, name string, conn *websocket.Conn, room *Room, hub *Hub) *Client {
	return &Client{
		id:   id,
		name: name,
		conn: conn,
		send: make(chan []byte, sendBuffer),
		room: room,
		hub:  hub,
	}
}

// trySend enqueues data without blocking; drops on overflow to protect the hub.
func (c *Client) trySend(data []byte) {
	select {
	case c.send <- data:
	default:
		c.hub.logger.Warn("ws send buffer full, dropping message",
			"participant_id", c.id, "meeting_id", c.room.slug)
	}
}

// readPump reads messages from the connection and dispatches them.
func (c *Client) readPump() {
	defer func() {
		c.room.remove(c.id)
		c.room.broadcast(encode(Envelope{
			Type:    TypeParticipantLeft,
			Payload: mustMarshal(PeerInfo{ID: c.id, Name: c.name}),
		}), c.id)
		if c.hub.metrics != nil {
			c.hub.metrics.WSConnectionsActive.Dec()
		}
		_ = c.conn.Close()
		close(c.send)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			c.sendError("bad_message", "invalid JSON envelope")
			continue
		}
		c.handle(env)
	}
}

// writePump pushes queued messages and keepalive pings to the connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendError(code, msg string) {
	c.trySend(encode(Envelope{
		Type:    TypeError,
		Payload: mustMarshal(ErrorPayload{Code: code, Message: msg}),
	}))
}

// handle dispatches a single inbound envelope.
func (c *Client) handle(env Envelope) {
	switch env.Type {
	case TypeSignalOffer, TypeSignalAnswer, TypeSignalICE:
		c.relaySignal(env)
	case TypeSpeechReceived:
		c.handleSpeech(env)
	default:
		c.sendError("unknown_type", "unsupported message type: "+env.Type)
	}
}

// relaySignal forwards a directed WebRTC signaling message to its target.
func (c *Client) relaySignal(env Envelope) {
	if env.To == "" {
		c.sendError("missing_to", env.Type+" requires a 'to' field")
		return
	}
	env.From = c.id // never trust client-supplied identity
	if !c.room.sendTo(env.To, encode(env)) {
		c.sendError("peer_not_found", "target participant not in room")
	}
}

// handleSpeech runs the audio chunk through STT then translation and broadcasts results.
func (c *Client) handleSpeech(env Envelope) {
	var p SpeechReceivedPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		c.sendError("bad_payload", "invalid speech.received payload")
		return
	}
	audio, err := base64.StdEncoding.DecodeString(p.Audio)
	if err != nil {
		c.sendError("bad_audio", "audio must be base64")
		return
	}

	lang := p.Lang
	if lang == "" {
		lang = "en"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	res, err := c.hub.stt.Transcribe(ctx, audio, lang)
	if c.hub.metrics != nil {
		c.hub.metrics.TranscriptionLatency.
			WithLabelValues(c.hub.stt.Name()).
			Observe(time.Since(start).Seconds())
	}
	if err != nil {
		c.hub.logger.Error("transcription failed", "error", err, "meeting_id", c.room.slug)
		if c.hub.metrics != nil {
			c.hub.metrics.ErrorsTotal.WithLabelValues("transcription").Inc()
		}
		c.sendError("stt_failed", "transcription failed")
		return
	}
	if res.Text == "" {
		return
	}

	c.room.broadcast(encode(Envelope{
		Type: TypeTranscriptUpdated,
		From: c.id,
		Payload: mustMarshal(TranscriptPayload{
			ParticipantID: c.id,
			Text:          res.Text,
			Lang:          res.Lang,
			IsFinal:       res.IsFinal,
			Seq:           p.Seq,
		}),
	}), "")

	target := c.hub.defaultTargetLang
	if target == "" || target == res.Lang {
		return
	}
	translated, err := c.hub.tr.Translate(ctx, res.Text, res.Lang, target)
	if err != nil {
		c.hub.logger.Error("translation failed", "error", err, "meeting_id", c.room.slug)
		if c.hub.metrics != nil {
			c.hub.metrics.ErrorsTotal.WithLabelValues("translation").Inc()
		}
		return
	}
	if translated == "" {
		return
	}
	c.room.broadcast(encode(Envelope{
		Type: TypeTranslationUpdated,
		From: c.id,
		Payload: mustMarshal(TranslationPayload{
			ParticipantID: c.id,
			Text:          translated,
			SourceLang:    res.Lang,
			TargetLang:    target,
			Seq:           p.Seq,
		}),
	}), "")
}
