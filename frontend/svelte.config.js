import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter(),
		// Root-relative /_app/... assets so nested routes (/m/slug) work after client nav.
		paths: {
			relative: false
		}
	}
};

export default config;
