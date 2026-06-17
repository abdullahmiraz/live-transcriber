import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		host: true,
		port: 3000,
		// During local `vite dev`, proxy API/WS to the Go backend so the browser can use
		// relative paths just like it does behind nginx in production.
		proxy: {
			'/api': 'http://localhost:8080',
			'/healthz': 'http://localhost:8080',
			'/ws': {
				target: 'ws://localhost:8080',
				ws: true
			}
		}
	}
});
