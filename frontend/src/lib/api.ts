// Thin REST client for the meetings API. Uses relative paths so it works behind nginx
// and via the Vite dev proxy.

import { API_BASE } from '$lib/config/app';
import type { ChatMessage, Meeting } from '$lib/api/types';

export type { ChatMessage, Meeting, MeetingStatus } from '$lib/api/types';

interface ApiErrorBody {
	error?: { code: string; message: string };
}

async function handle<T>(res: Response): Promise<T> {
	if (!res.ok) {
		let message = `request failed (${res.status})`;
		try {
			const body = (await res.json()) as ApiErrorBody;
			if (body.error?.message) message = body.error.message;
		} catch {
			// ignore parse errors
		}
		throw new Error(message);
	}
	return (await res.json()) as T;
}

export async function createMeeting(input: {
	title?: string;
	host_name?: string;
}): Promise<Meeting> {
	const res = await fetch(`${API_BASE}/meetings`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ title: input.title ?? '', host_name: input.host_name ?? '' })
	});
	return handle<Meeting>(res);
}

export async function getMeeting(slug: string): Promise<Meeting> {
	const res = await fetch(`${API_BASE}/meetings/${encodeURIComponent(slug)}`);
	return handle<Meeting>(res);
}

export async function endMeeting(slug: string): Promise<Meeting> {
	const res = await fetch(`${API_BASE}/meetings/${encodeURIComponent(slug)}/end`, {
		method: 'POST'
	});
	return handle<Meeting>(res);
}

export async function listMessages(
	slug: string,
	opts?: { limit?: number; before?: string }
): Promise<ChatMessage[]> {
	const params = new URLSearchParams();
	if (opts?.limit) params.set('limit', String(opts.limit));
	if (opts?.before) params.set('before', opts.before);
	const qs = params.toString();
	const res = await fetch(
		`${API_BASE}/meetings/${encodeURIComponent(slug)}/messages${qs ? `?${qs}` : ''}`
	);
	const data = await handle<{ messages: ChatMessage[] }>(res);
	return data.messages;
}
