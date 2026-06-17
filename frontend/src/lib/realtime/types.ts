// Mirror of the backend WebSocket contract (see docs/api-design.md).

export type EventType =
	| 'room.welcome'
	| 'participant.joined'
	| 'participant.left'
	| 'signal.offer'
	| 'signal.answer'
	| 'signal.ice'
	| 'speech.received'
	| 'transcript.updated'
	| 'translation.updated'
	| 'chat.message'
	| 'chat.new'
	| 'error';

export interface Envelope<T = unknown> {
	type: EventType;
	from?: string;
	to?: string;
	payload?: T;
}

export interface PeerInfo {
	id: string;
	name: string;
}

export interface WelcomePayload {
	selfId: string;
	participants: PeerInfo[];
}

export interface SignalOfferPayload {
	sdp: RTCSessionDescriptionInit;
}
export interface SignalAnswerPayload {
	sdp: RTCSessionDescriptionInit;
}
export interface SignalICEPayload {
	candidate: RTCIceCandidateInit;
}

export interface SpeechReceivedPayload {
	audio: string; // base64
	seq: number;
	lang: string;
	targetLang?: string;
}

export interface TranscriptPayload {
	participantId: string;
	text: string;
	lang: string;
	isFinal: boolean;
	seq: number;
}

export interface TranslationPayload {
	participantId: string;
	text: string;
	sourceLang: string;
	targetLang: string;
	seq: number;
}

export interface ChatMessageOutPayload {
	content: string;
}

export type { ChatMessage } from '$lib/api/types';

export interface ErrorPayload {
	code: string;
	message: string;
}

export type WsConnectionStatus = 'connecting' | 'open' | 'closed';
