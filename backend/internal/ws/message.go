package ws

import "encoding/json"

// Envelope is the universal WebSocket message format (see docs/api-design.md).
type Envelope struct {
	Type    string          `json:"type"`
	From    string          `json:"from,omitempty"`
	To      string          `json:"to,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Event type constants.
const (
	TypeWelcome            = "room.welcome"
	TypeParticipantJoined  = "participant.joined"
	TypeParticipantLeft    = "participant.left"
	TypeSignalOffer        = "signal.offer"
	TypeSignalAnswer       = "signal.answer"
	TypeSignalICE          = "signal.ice"
	TypeSpeechReceived     = "speech.received"
	TypeTranscriptUpdated  = "transcript.updated"
	TypeTranslationUpdated = "translation.updated"
	TypeChatMessage        = "chat.message"
	TypeChatNew            = "chat.new"
	TypeError              = "error"
)

// PeerInfo identifies a participant in presence events.
type PeerInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// WelcomePayload is sent to a client right after it joins.
type WelcomePayload struct {
	SelfID       string     `json:"selfId"`
	Participants []PeerInfo `json:"participants"`
}

// SpeechReceivedPayload is sent by a client with an audio/text chunk.
type SpeechReceivedPayload struct {
	Audio      string `json:"audio"` // base64-encoded chunk
	Seq        int    `json:"seq"`
	Lang       string `json:"lang"`
	TargetLang string `json:"targetLang,omitempty"`
}

// TranscriptPayload is broadcast after STT.
type TranscriptPayload struct {
	ParticipantID string `json:"participantId"`
	Text          string `json:"text"`
	Lang          string `json:"lang"`
	IsFinal       bool   `json:"isFinal"`
	Seq           int    `json:"seq"`
}

// TranslationPayload is broadcast after translation.
type TranslationPayload struct {
	ParticipantID string `json:"participantId"`
	Text          string `json:"text"`
	SourceLang    string `json:"sourceLang"`
	TargetLang    string `json:"targetLang"`
	Seq           int    `json:"seq"`
}

// ChatMessagePayload is sent by a client to post a chat message (text-only).
type ChatMessagePayload struct {
	Content string `json:"content"`
}

// ErrorPayload describes a protocol error sent back to a client.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

func encode(e Envelope) []byte {
	b, _ := json.Marshal(e)
	return b
}
