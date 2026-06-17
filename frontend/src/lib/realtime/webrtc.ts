import type { SignalingClient } from './signaling';
import type {
	Envelope,
	WelcomePayload,
	PeerInfo,
	SignalOfferPayload,
	SignalAnswerPayload,
	SignalICEPayload
} from './types';

// MVP uses a public STUN server only. Add a TURN server here for restrictive NATs.
const ICE_SERVERS: RTCIceServer[] = [{ urls: 'stun:stun.l.google.com:19302' }];

interface PeerState {
	pc: RTCPeerConnection;
	name: string;
	pendingCandidates: RTCIceCandidateInit[];
	remoteSet: boolean;
}

export interface MeshCallbacks {
	onRemoteStream: (peerId: string, name: string, stream: MediaStream) => void;
	onPeerLeft: (peerId: string) => void;
	onRosterChange?: (peers: PeerInfo[]) => void;
}

/**
 * MeshManager implements a full-mesh WebRTC topology: the newcomer initiates an offer to
 * every existing peer; existing peers answer. Signaling is relayed by the Go hub. This is
 * intentionally simple for the MVP (small rooms); see docs/architecture.md §6 for the SFU
 * migration path.
 */
export class MeshManager {
	private selfId = '';
	private readonly peers = new Map<string, PeerState>();
	private readonly names = new Map<string, string>();

	constructor(
		private readonly sig: SignalingClient,
		private readonly localStream: MediaStream,
		private readonly cbs: MeshCallbacks
	) {
		this.register();
	}

	get id(): string {
		return this.selfId;
	}

	private register(): void {
		this.sig.on('room.welcome', (env) => this.onWelcome(env));
		this.sig.on('participant.joined', (env) => this.onJoined(env));
		this.sig.on('participant.left', (env) => this.onLeft(env));
		this.sig.on('signal.offer', (env) => void this.onOffer(env));
		this.sig.on('signal.answer', (env) => void this.onAnswer(env));
		this.sig.on('signal.ice', (env) => void this.onIce(env));
	}

	private async onWelcome(env: Envelope): Promise<void> {
		const p = env.payload as WelcomePayload;
		this.selfId = p.selfId;
		for (const peer of p.participants) {
			this.names.set(peer.id, peer.name);
			// Newcomer (this client) initiates to each existing peer.
			const state = this.createPeer(peer.id, peer.name);
			await this.makeOffer(peer.id, state);
		}
		this.emitRoster();
	}

	private onJoined(env: Envelope): void {
		const peer = env.payload as PeerInfo;
		this.names.set(peer.id, peer.name);
		// Do not initiate; the newcomer will send us an offer. Just track presence.
		this.emitRoster();
	}

	private onLeft(env: Envelope): void {
		const peer = env.payload as PeerInfo;
		this.closePeer(peer.id);
		this.names.delete(peer.id);
		this.cbs.onPeerLeft(peer.id);
		this.emitRoster();
	}

	private async onOffer(env: Envelope): Promise<void> {
		const from = env.from;
		if (!from) return;
		const p = env.payload as SignalOfferPayload;
		let state = this.peers.get(from);
		if (!state) state = this.createPeer(from, this.names.get(from) ?? 'Guest');
		await state.pc.setRemoteDescription(new RTCSessionDescription(p.sdp));
		state.remoteSet = true;
		await this.flushCandidates(state);
		const answer = await state.pc.createAnswer();
		await state.pc.setLocalDescription(answer);
		this.sig.send<SignalAnswerPayload>({
			type: 'signal.answer',
			to: from,
			payload: { sdp: answer }
		});
	}

	private async onAnswer(env: Envelope): Promise<void> {
		const from = env.from;
		if (!from) return;
		const state = this.peers.get(from);
		if (!state) return;
		const p = env.payload as SignalAnswerPayload;
		await state.pc.setRemoteDescription(new RTCSessionDescription(p.sdp));
		state.remoteSet = true;
		await this.flushCandidates(state);
	}

	private async onIce(env: Envelope): Promise<void> {
		const from = env.from;
		if (!from) return;
		const state = this.peers.get(from);
		if (!state) return;
		const p = env.payload as SignalICEPayload;
		if (state.remoteSet) {
			await state.pc.addIceCandidate(p.candidate).catch(() => {});
		} else {
			state.pendingCandidates.push(p.candidate);
		}
	}

	private createPeer(peerId: string, name: string): PeerState {
		const pc = new RTCPeerConnection({ iceServers: ICE_SERVERS });
		const state: PeerState = { pc, name, pendingCandidates: [], remoteSet: false };

		for (const track of this.localStream.getTracks()) {
			pc.addTrack(track, this.localStream);
		}

		pc.onicecandidate = (e) => {
			if (e.candidate) {
				this.sig.send<SignalICEPayload>({
					type: 'signal.ice',
					to: peerId,
					payload: { candidate: e.candidate.toJSON() }
				});
			}
		};

		pc.ontrack = (e) => {
			const stream = e.streams[0] ?? new MediaStream([e.track]);
			this.cbs.onRemoteStream(peerId, name, stream);
		};

		pc.onconnectionstatechange = () => {
			if (pc.connectionState === 'failed' || pc.connectionState === 'closed') {
				this.closePeer(peerId);
				this.cbs.onPeerLeft(peerId);
			}
		};

		this.peers.set(peerId, state);
		return state;
	}

	private async makeOffer(peerId: string, state: PeerState): Promise<void> {
		const offer = await state.pc.createOffer();
		await state.pc.setLocalDescription(offer);
		this.sig.send<SignalOfferPayload>({
			type: 'signal.offer',
			to: peerId,
			payload: { sdp: offer }
		});
	}

	private async flushCandidates(state: PeerState): Promise<void> {
		for (const c of state.pendingCandidates) {
			await state.pc.addIceCandidate(c).catch(() => {});
		}
		state.pendingCandidates = [];
	}

	private closePeer(peerId: string): void {
		const state = this.peers.get(peerId);
		if (!state) return;
		try {
			state.pc.close();
		} catch {
			// ignore
		}
		this.peers.delete(peerId);
	}

	private emitRoster(): void {
		const roster: PeerInfo[] = [...this.names.entries()].map(([id, name]) => ({ id, name }));
		this.cbs.onRosterChange?.(roster);
	}

	close(): void {
		for (const id of [...this.peers.keys()]) this.closePeer(id);
		this.names.clear();
	}
}
