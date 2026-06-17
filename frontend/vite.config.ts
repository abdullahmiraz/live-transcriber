import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

const polling = process.env.CHOKIDAR_USEPOLLING === 'true';
const hmrClientPort = process.env.VITE_HMR_CLIENT_PORT
	? Number(process.env.VITE_HMR_CLIENT_PORT)
	: undefined;

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		host: true,
		port: 3000,
		// File watching in Docker (especially on Windows bind mounts).
		watch: polling ? { usePolling: true, interval: 500 } : undefined,
		// HMR through nginx (HTTPS on :443 or HTTP on :80) when running in Docker dev.
		hmr: hmrClientPort
			? {
					protocol: process.env.VITE_HMR_PROTOCOL === 'ws' ? 'ws' : 'wss',
					clientPort: hmrClientPort
				}
			: undefined,
		// During local `vite dev`, proxy API/WS to the Go backend.
		proxy: {
			'/api': {
				target: process.env.VITE_BACKEND_URL || 'http://localhost:8080',
				changeOrigin: true
			},
			'/healthz': { target: process.env.VITE_BACKEND_URL || 'http://localhost:8080' },
			'/readyz': { target: process.env.VITE_BACKEND_URL || 'http://localhost:8080' },
			'/metrics': { target: process.env.VITE_BACKEND_URL || 'http://localhost:8080' },
			'/ws': {
				target: process.env.VITE_BACKEND_WS || 'ws://localhost:8080',
				ws: true,
				changeOrigin: true
			}
		}
	}
});
