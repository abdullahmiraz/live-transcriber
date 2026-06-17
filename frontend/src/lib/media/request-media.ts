export type LocalMediaResult = {
	stream: MediaStream;
	hasVideo: boolean;
	hasAudio: boolean;
	/** Human-readable note when joining with partial devices */
	note?: string;
};

/** Which devices to request when joining a meeting. */
export type MediaJoinPreferences = {
	audio: boolean;
	video: boolean;
};

function isPermissionDenied(err: unknown): boolean {
	return (
		err instanceof DOMException &&
		(err.name === 'NotAllowedError' || err.name === 'PermissionDeniedError')
	);
}

async function tryGet(
	md: MediaDevices,
	constraints: MediaStreamConstraints
): Promise<MediaStream | null> {
	try {
		return await md.getUserMedia(constraints);
	} catch (e) {
		if (isPermissionDenied(e)) throw e;
		return null;
	}
}

function mergeStreams(...streams: (MediaStream | null | undefined)[]): MediaStream | null {
	const tracks: MediaStreamTrack[] = [];
	for (const s of streams) {
		if (!s) continue;
		for (const t of s.getTracks()) tracks.push(t);
	}
	return tracks.length > 0 ? new MediaStream(tracks) : null;
}

async function tryExplicitDevices(md: MediaDevices): Promise<MediaStream | null> {
	let devices: MediaDeviceInfo[] = [];
	try {
		devices = await md.enumerateDevices();
	} catch {
		return null;
	}

	const audioId = devices.find((d) => d.kind === 'audioinput' && d.deviceId)?.deviceId;
	const videoId = devices.find((d) => d.kind === 'videoinput' && d.deviceId)?.deviceId;
	if (!audioId && !videoId) return null;

	const audio = audioId
		? await tryGet(md, { audio: { deviceId: { ideal: audioId } }, video: false })
		: null;
	const video = videoId
		? await tryGet(md, { audio: false, video: { deviceId: { ideal: videoId } } })
		: null;

	return mergeStreams(audio, video);
}

/**
 * Request camera and/or microphone with fallbacks.
 * Pass `{ audio: false, video: false }` to join without devices (chat / view only).
 */
export async function requestLocalMedia(
	prefs: MediaJoinPreferences = { audio: true, video: true }
): Promise<LocalMediaResult> {
	const { audio: wantAudio, video: wantVideo } = prefs;

	if (!wantAudio && !wantVideo) {
		return {
			stream: new MediaStream(),
			hasVideo: false,
			hasAudio: false,
			note: 'Joined without camera or microphone. You can still chat and watch others.'
		};
	}

	// On many mobile browsers (and all iOS browsers), camera/mic are HTTPS-only.
	// When served over plain HTTP on a LAN IP, `navigator.mediaDevices` may appear missing,
	// which is confusing. Surface this as a secure-context error instead.
	if (typeof window !== 'undefined' && !window.isSecureContext) {
		throw new DOMException(
			'Camera and microphone require a secure connection on phones and LAN IPs. Open the site over HTTPS (accept the certificate warning once), then try again.',
			'SecurityError'
		);
	}

	const md = navigator.mediaDevices;
	if (!md?.getUserMedia) {
		throw new DOMException('Media devices are not available in this browser.', 'NotSupportedError');
	}

	const attempts: MediaStreamConstraints[] = [];
	if (wantAudio && wantVideo) {
		attempts.push(
			{ audio: true, video: true },
			{
				audio: { echoCancellation: true, noiseSuppression: true },
				video: { width: { ideal: 1280 }, height: { ideal: 720 } }
			},
			{ audio: true, video: false },
			{ audio: false, video: true }
		);
	} else if (wantAudio) {
		attempts.push(
			{ audio: true, video: false },
			{ audio: { echoCancellation: true, noiseSuppression: true }, video: false }
		);
	} else {
		attempts.push(
			{ audio: false, video: true },
			{ audio: false, video: { width: { ideal: 1280 }, height: { ideal: 720 } } }
		);
	}

	for (const constraints of attempts) {
		const stream = await tryGet(md, constraints);
		if (stream) {
			const hasVideo = stream.getVideoTracks().length > 0;
			const hasAudio = stream.getAudioTracks().length > 0;
			let note: string | undefined;
			if (wantAudio && wantVideo) {
				if (hasAudio && !hasVideo) note = 'Joined with microphone only (no camera in use).';
				if (hasVideo && !hasAudio) note = 'Joined with camera only (no microphone in use).';
			} else if (wantAudio && !hasVideo) {
				note = 'Joined with microphone only.';
			} else if (wantVideo && !hasAudio) {
				note = 'Joined with camera only.';
			}
			return { stream, hasVideo, hasAudio, note };
		}
	}

	if (wantAudio && wantVideo) {
		const audioOnly = await tryGet(md, { audio: true, video: false });
		const videoOnly = await tryGet(md, { audio: false, video: true });
		const merged = mergeStreams(audioOnly, videoOnly);
		if (merged) {
			const hasVideo = merged.getVideoTracks().length > 0;
			const hasAudio = merged.getAudioTracks().length > 0;
			let note: string | undefined;
			if (hasAudio && !hasVideo) note = 'Joined with microphone only (no camera in use).';
			if (hasVideo && !hasAudio) note = 'Joined with camera only (no microphone in use).';
			return { stream: merged, hasVideo, hasAudio, note };
		}

		const explicit = await tryExplicitDevices(md);
		if (explicit) {
			const hasVideo = explicit.getVideoTracks().length > 0;
			const hasAudio = explicit.getAudioTracks().length > 0;
			return { stream: explicit, hasVideo, hasAudio };
		}
	}

	const deviceLabel = wantAudio && wantVideo ? 'camera or microphone' : wantAudio ? 'microphone' : 'camera';
	throw new DOMException(
		`Could not open your ${deviceLabel}. Check that devices are connected and allowed in system settings.`,
		'NotFoundError'
	);
}

/** Best-effort device counts for the lobby (labels may be blank until permission is granted). */
export async function summarizeMediaDevices(): Promise<{ cameras: number; microphones: number }> {
	const md = navigator.mediaDevices;
	if (!md?.enumerateDevices) return { cameras: 0, microphones: 0 };
	try {
		const devices = await md.enumerateDevices();
		return {
			cameras: devices.filter((d) => d.kind === 'videoinput').length,
			microphones: devices.filter((d) => d.kind === 'audioinput').length
		};
	} catch {
		return { cameras: 0, microphones: 0 };
	}
}

export function mediaErrorMessage(err: unknown): string {
	const name = err instanceof DOMException ? err.name : '';
	switch (name) {
		case 'NotAllowedError':
		case 'PermissionDeniedError':
			return 'Camera and microphone access was blocked. Click the lock icon in your browser address bar, allow camera and microphone, then try again.';
		case 'NotFoundError':
		case 'DevicesNotFoundError':
			return 'Could not open your camera or microphone. If they are connected, open Windows Settings → Privacy → Camera / Microphone and allow your browser, then reload and try again.';
		case 'NotReadableError':
		case 'TrackStartError':
			return 'Your camera or microphone is in use by another application (Zoom, Teams, etc.). Close it and try again.';
		case 'SecurityError':
			return 'Camera and microphone require HTTPS on phones and LAN IPs. Open https://<your-pc-ip>/ (same Wi‑Fi), accept the certificate warning once, then try again. On this PC, http://localhost also works.';
		case 'NotSupportedError':
			return 'This browser does not support camera or microphone access.';
		default:
			return err instanceof Error ? err.message : 'Could not access camera or microphone.';
	}
}
