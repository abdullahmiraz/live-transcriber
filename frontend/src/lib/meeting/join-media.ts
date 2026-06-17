import type { MediaJoinPreferences } from '$lib/media/request-media';
import type { JoinMediaMode } from './types';

export function joinModeToPrefs(mode: JoinMediaMode): MediaJoinPreferences {
	switch (mode) {
		case 'none':
			return { audio: false, video: false };
		case 'audio':
			return { audio: true, video: false };
		case 'video':
			return { audio: false, video: true };
		case 'both':
		default:
			return { audio: true, video: true };
	}
}
