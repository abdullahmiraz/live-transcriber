// SpeechCapture wraps the browser Web Speech API to produce live transcript text with
// zero backend cost and no API keys (free, no credit card — see docs/stt-decision.md).
// The recognized text is sent to the backend as `speech.received`; the backend's STT
// provider (mock by default) and translation provider then fan results out to the room.
//
// Note: the Web Speech API is best supported in Chrome/Edge. When unavailable, captions
// simply stay off; a server-side provider (whisper/deepgram) can replace this later.

/* eslint-disable @typescript-eslint/no-explicit-any */

export interface SpeechCaptureCallbacks {
	onText: (text: string, isFinal: boolean) => void;
	onError?: (message: string) => void;
	onRunningChange?: (running: boolean) => void;
}

function getRecognitionCtor(): any {
	if (typeof window === 'undefined') return null;
	return (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition || null;
}

export class SpeechCapture {
	private recognition: any = null;
	private running = false;
	private stopped = true;

	constructor(
		private lang: string,
		private readonly cbs: SpeechCaptureCallbacks
	) {}

	static isSupported(): boolean {
		return getRecognitionCtor() !== null;
	}

	start(): void {
		const Ctor = getRecognitionCtor();
		if (!Ctor) {
			this.cbs.onError?.('Speech recognition is not supported in this browser.');
			return;
		}
		this.stopped = false;
		const rec = new Ctor();
		rec.lang = this.lang;
		rec.continuous = true;
		rec.interimResults = true;

		rec.onresult = (event: any) => {
			for (let i = event.resultIndex; i < event.results.length; i++) {
				const result = event.results[i];
				const transcript = String(result[0]?.transcript ?? '').trim();
				if (transcript) this.cbs.onText(transcript, Boolean(result.isFinal));
			}
		};

		rec.onerror = (event: any) => {
			const err = String(event?.error ?? 'unknown');
			// "no-speech"/"aborted" are normal; surface only real failures.
			if (err !== 'no-speech' && err !== 'aborted') {
				this.cbs.onError?.(`Speech recognition error: ${err}`);
			}
		};

		rec.onend = () => {
			this.running = false;
			this.cbs.onRunningChange?.(false);
			// Auto-restart for continuous capture unless explicitly stopped.
			if (!this.stopped) {
				try {
					rec.start();
					this.running = true;
					this.cbs.onRunningChange?.(true);
				} catch {
					// ignore restart races
				}
			}
		};

		this.recognition = rec;
		try {
			rec.start();
			this.running = true;
			this.cbs.onRunningChange?.(true);
		} catch (e) {
			this.cbs.onError?.(`Could not start speech recognition: ${String(e)}`);
		}
	}

	setLang(lang: string): void {
		this.lang = lang;
		if (this.recognition) this.recognition.lang = lang;
	}

	stop(): void {
		this.stopped = true;
		this.running = false;
		if (this.recognition) {
			try {
				this.recognition.stop();
			} catch {
				// ignore
			}
		}
		this.cbs.onRunningChange?.(false);
	}

	get isRunning(): boolean {
		return this.running;
	}
}
