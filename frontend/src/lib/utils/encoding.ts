/** Encode UTF-8 text as base64 for speech.received payloads. */
export function utf8ToBase64(text: string): string {
	return btoa(String.fromCharCode(...new TextEncoder().encode(text)));
}
