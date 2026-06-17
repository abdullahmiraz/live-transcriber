import {
	DEFAULT_GUEST_NAME,
	DEFAULT_HOST_NAME,
	DISPLAY_NAME_STORAGE_KEY
} from './constants';

export function getDisplayName(fallback = DEFAULT_GUEST_NAME): string {
	if (typeof sessionStorage === 'undefined') return fallback;
	return sessionStorage.getItem(DISPLAY_NAME_STORAGE_KEY) || fallback;
}

export function setDisplayName(name: string, fallback = DEFAULT_GUEST_NAME): void {
	if (typeof sessionStorage === 'undefined') return;
	sessionStorage.setItem(DISPLAY_NAME_STORAGE_KEY, name.trim() || fallback);
}

export { DEFAULT_GUEST_NAME, DEFAULT_HOST_NAME };
