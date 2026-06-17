import type { Envelope, EventType } from './types';

type Handler = (env: Envelope) => void;

function buildWsUrl(slug: string, name: string): string {
	const proto = location.protocol === 'https:' ? 'wss' : 'ws';
	const params = new URLSearchParams({ meeting: slug, name });
	return `${proto}://${location.host}/ws?${params.toString()}`;
}

/**
 * SignalingClient is a thin, typed wrapper over a reconnecting WebSocket carrying both
 * WebRTC signaling and realtime app events (see docs/api-design.md).
 */
export class SignalingClient {
	private ws: WebSocket | null = null;
	private readonly url: string;
	private readonly handlers = new Map<EventType, Set<Handler>>();
	private outbox: string[] = [];
	private shouldReconnect = true;
	private reconnectDelay = 1000;

	onStatusChange?: (status: 'connecting' | 'open' | 'closed') => void;
	/** Fires after each successful WebSocket open (initial + reconnect). */
	onOpen?: () => void;

	constructor(slug: string, name: string) {
		this.url = buildWsUrl(slug, name);
	}

	connect(): void {
		this.onStatusChange?.('connecting');
		const ws = new WebSocket(this.url);
		this.ws = ws;

		ws.onopen = () => {
			this.reconnectDelay = 1000;
			this.onStatusChange?.('open');
			for (const msg of this.outbox) ws.send(msg);
			this.outbox = [];
			this.onOpen?.();
		};

		ws.onmessage = (ev) => {
			let env: Envelope;
			try {
				env = JSON.parse(ev.data) as Envelope;
			} catch {
				return;
			}
			const set = this.handlers.get(env.type);
			if (set) for (const h of set) h(env);
		};

		ws.onclose = () => {
			this.onStatusChange?.('closed');
			if (this.shouldReconnect) {
				setTimeout(() => this.connect(), this.reconnectDelay);
				this.reconnectDelay = Math.min(this.reconnectDelay * 2, 10000);
			}
		};

		ws.onerror = () => ws.close();
	}

	on(type: EventType, handler: Handler): () => void {
		let set = this.handlers.get(type);
		if (!set) {
			set = new Set();
			this.handlers.set(type, set);
		}
		set.add(handler);
		return () => set?.delete(handler);
	}

	send<T>(env: Envelope<T>): boolean {
		const data = JSON.stringify(env);
		if (this.ws && this.ws.readyState === WebSocket.OPEN) {
			this.ws.send(data);
			return true;
		}
		this.outbox.push(data);
		return false;
	}

	isOpen(): boolean {
		return this.ws?.readyState === WebSocket.OPEN;
	}

	isConnecting(): boolean {
		return this.ws?.readyState === WebSocket.CONNECTING;
	}

	/** Reconnect when the socket dropped (e.g. phone backgrounded). */
	reconnectIfNeeded(): void {
		if (!this.shouldReconnect) return;
		const state = this.ws?.readyState;
		if (state === WebSocket.OPEN || state === WebSocket.CONNECTING) return;
		this.connect();
	}

	close(): void {
		this.shouldReconnect = false;
		this.ws?.close();
		this.ws = null;
	}
}
