export type LocalMediaResult = {
	stream: MediaStream;
	hasVideo: boolean;
	hasAudio: boolean;
	/** Human-readable note when joining with partial devices */
	note?: string;
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
 * A single `{ video: true, audio: true }` call fails with NotFoundError when either
 * device is missing or misreported — common on Windows with privacy/driver quirks.
 */
export async function requestLocalMedia(): Promise<LocalMediaResult> {
	const md = navigator.mediaDevices;
	if (!md?.getUserMedia) {
		throw new DOMException('Media devices are not available in this browser.', 'NotSupportedError');
	}

	const attempts: MediaStreamConstraints[] = [
		{ audio: true, video: true },
		{
			audio: { echoCancellation: true, noiseSuppression: true },
			video: { width: { ideal: 1280 }, height: { ideal: 720 } }
		},
		{ audio: true, video: false },
		{ audio: false, video: true }
	];

	for (const constraints of attempts) {
		const stream = await tryGet(md, constraints);
		if (stream) {
			const hasVideo = stream.getVideoTracks().length > 0;
			const hasAudio = stream.getAudioTracks().length > 0;
			let note: string | undefined;
			if (hasAudio && !hasVideo) note = 'Joined with microphone only (no camera in use).';
			if (hasVideo && !hasAudio) note = 'Joined with camera only (no microphone in use).';
			return { stream, hasVideo, hasAudio, note };
		}
	}

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

	throw new DOMException(
		'Could not open a camera or microphone. Check that devices are connected and allowed in system settings.',
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
			return 'Camera and microphone require a secure connection. Use http://localhost (not an IP address) or HTTPS.';
		case 'NotSupportedError':
			return 'This browser does not support camera or microphone access.';
		default:
			return err instanceof Error ? err.message : 'Could not access camera or microphone.';
	}
}
