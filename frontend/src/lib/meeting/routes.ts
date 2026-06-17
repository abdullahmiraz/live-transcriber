/** Meeting room path for a slug. */
export function meetingPath(slug: string): string {
	return `/m/${slug}`;
}

/** Parse a meeting code or full URL into a slug. */
export function parseMeetingSlug(input: string): string {
	const raw = input.trim();
	if (!raw) return '';
	return raw.includes('/m/') ? raw.split('/m/')[1].split(/[?#]/)[0] : raw;
}
