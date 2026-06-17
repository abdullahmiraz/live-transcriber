import type { PageLoad } from './$types';

// The meeting room relies on browser-only APIs (getUserMedia, WebRTC, WebSocket),
// so render it on the client only.
export const ssr = false;

export const load: PageLoad = ({ params }) => {
	return { slug: params.slug };
};
