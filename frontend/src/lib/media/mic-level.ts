/** Samples microphone input level (0–1) from a MediaStream for UI visualizers. */
export class MicLevelMonitor {
	private ctx: AudioContext | null = null;
	private analyser: AnalyserNode | null = null;
	private source: MediaStreamAudioSourceNode | null = null;
	private raf = 0;
	private buffer: Uint8Array | null = null;
	private onLevel: (level: number) => void;

	constructor(onLevel: (level: number) => void) {
		this.onLevel = onLevel;
	}

	attach(stream: MediaStream | null, active: boolean): void {
		this.detach();
		if (!active || !stream?.getAudioTracks().some((t) => t.enabled && t.readyState === 'live')) {
			this.onLevel(0);
			return;
		}

		try {
			this.ctx = new AudioContext();
			this.analyser = this.ctx.createAnalyser();
			this.analyser.fftSize = 256;
			this.analyser.smoothingTimeConstant = 0.65;
			this.source = this.ctx.createMediaStreamSource(stream);
			this.source.connect(this.analyser);
			this.buffer = new Uint8Array(this.analyser.frequencyBinCount);

			const tick = () => {
				if (!this.analyser || !this.buffer) return;
				this.analyser.getByteFrequencyData(this.buffer);
				// Voice energy sits in lower bins; weight them more than treble hiss.
				let weighted = 0;
				let totalWeight = 0;
				for (let i = 0; i < this.buffer.length; i++) {
					const w = i < 16 ? 2.2 : i < 32 ? 1.2 : 0.5;
					weighted += this.buffer[i] * w;
					totalWeight += w;
				}
				const raw = weighted / (totalWeight * 255);
				// Noise gate + gentle curve for visible but not jumpy motion.
				const gated = raw < 0.04 ? 0 : (raw - 0.04) / 0.96;
				this.onLevel(Math.min(1, Math.pow(gated, 0.75) * 1.15));
				this.raf = requestAnimationFrame(tick);
			};
			tick();
		} catch {
			this.onLevel(0);
		}
	}

	detach(): void {
		if (this.raf) cancelAnimationFrame(this.raf);
		this.raf = 0;
		this.source?.disconnect();
		this.source = null;
		this.analyser = null;
		this.buffer = null;
		if (this.ctx) {
			void this.ctx.close();
			this.ctx = null;
		}
		this.onLevel(0);
	}
}
