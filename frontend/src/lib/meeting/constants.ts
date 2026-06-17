export const DISPLAY_NAME_STORAGE_KEY = 'displayName';

export const DEFAULT_GUEST_NAME = 'Guest';
export const DEFAULT_HOST_NAME = 'Host';
export const DEFAULT_MEETING_TITLE = 'Meeting';

export const CHAT_HISTORY_LIMIT = 50;
export const CHAT_POLL_INTERVAL_MS = 4000;
export const CHAT_MAX_CONTENT_LENGTH = 4000;

export const MAX_CAPTIONS = 80;
export const COPY_LINK_FEEDBACK_MS = 1500;

export const EMPTY_ROOM_TTL_MINUTES = 10;

export const DEFAULT_MEETING_TAB = 'chat' as const;

export function emptyRoomLeaveMessage(): string {
	return `If everyone leaves this meeting, the room will be deleted automatically after ${EMPTY_ROOM_TTL_MINUTES} minutes of being empty.`;
}
