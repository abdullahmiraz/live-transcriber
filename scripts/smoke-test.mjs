#!/usr/bin/env node
/**
 * End-to-end smoke test against http://localhost (nginx).
 * Usage: node scripts/smoke-test.mjs [BASE_URL]
 */
import { readFileSync, writeFileSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

const BASE = process.argv[2] ?? 'http://localhost';
const TMP = join(tmpdir(), 'live-meet-smoke-body.json');
const METRICS_TMP = join(tmpdir(), 'live-meet-smoke-metrics.txt');

let pass = 0;
let fail = 0;

function ok(msg) {
	console.log(`  OK  ${msg}`);
	pass++;
}
function bad(msg) {
	console.log(`  FAIL ${msg}`);
	fail++;
}

function curl(args) {
	const r = spawnSync('curl', ['-s', ...args], { encoding: 'utf8' });
	return { body: r.stdout ?? '', stderr: r.stderr ?? '', status: r.status };
}

function curlStatus(url, expect) {
	const r = spawnSync(
		'curl',
		['-s', '-o', TMP, '-w', '%{http_code}', url],
		{ encoding: 'utf8' }
	);
	const code = r.stdout?.trim() ?? '';
	if (code === expect) ok(`${url} (${code})`);
	else bad(`${url} (expected ${expect}, got ${code})`);
	return code;
}

function json() {
	try {
		return JSON.parse(readFileSync(TMP, 'utf8'));
	} catch {
		return null;
	}
}

async function testWebSocket(slug, name, onMessage, timeoutMs = 8000) {
	const wsUrl = `${BASE.replace(/^http/, 'ws')}/ws?meeting=${encodeURIComponent(slug)}&name=${encodeURIComponent(name)}`;
	return new Promise((resolve, reject) => {
		const ws = new WebSocket(wsUrl);
		const timer = setTimeout(() => {
			ws.close();
			reject(new Error('timeout'));
		}, timeoutMs);
		ws.addEventListener('message', (ev) => {
			try {
				const msg = JSON.parse(ev.data.toString());
				onMessage(msg, ws, () => {
					clearTimeout(timer);
					ws.close();
					resolve(true);
				});
			} catch {
				/* ignore */
			}
		});
		ws.addEventListener('error', () => {
			clearTimeout(timer);
			reject(new Error('WebSocket error'));
		});
	});
}

console.log('=== Live Meet smoke test ===');
console.log(`Base: ${BASE}\n`);

console.log('--- Health ---');
curlStatus(`${BASE}/healthz`, '200');
if (readFileSync(TMP, 'utf8').includes('"status":"ok"')) ok('healthz body');
else bad('healthz body');

curlStatus(`${BASE}/readyz`, '200');
if (readFileSync(TMP, 'utf8').includes('"status":"ready"')) ok('readyz body');
else bad('readyz body');

{
	const r = spawnSync('curl', ['-s', '-o', METRICS_TMP, '-w', '%{http_code}', `${BASE}/metrics`], {
		encoding: 'utf8'
	});
	const code = r.stdout?.trim();
	const metrics = readFileSync(METRICS_TMP, 'utf8');
	if (code === '200' && metrics.includes('http_requests_total')) ok('GET /metrics');
	else bad(`GET /metrics (${code})`);
}

console.log('\n--- Frontend pages ---');
curlStatus(`${BASE}/`, '200');
const home = readFileSync(TMP, 'utf8');
if (home.includes('Create meeting')) ok('home contains Create meeting');
else bad('home missing Create meeting');
const isDev = home.includes('__sveltekit_dev');
if (isDev) ok('home is Vite dev mode (hot reload)');
else if (home.includes('href="/_app/')) ok('home uses root-relative assets');
else bad('home asset paths');

console.log('\n--- REST API: meetings ---');
spawnSync(
	'curl',
	[
		'-s',
		'-o',
		TMP,
		'-X',
		'POST',
		`${BASE}/api/meetings`,
		'-H',
		'Content-Type: application/json',
		'-d',
		'{"title":"Smoke Test","host_name":"Tester"}'
	],
	{ encoding: 'utf8' }
);
const created = json();
const slug = created?.slug;
if (slug) ok(`POST /api/meetings created slug=${slug}`);
else {
	bad('POST /api/meetings');
	process.exit(1);
}

curlStatus(`${BASE}/api/meetings/${slug}`, '200');
if (json()?.status === 'active') ok('meeting status active');
else bad('meeting status');

console.log('\n--- REST API: chat history ---');
curlStatus(`${BASE}/api/meetings/${slug}/messages?limit=10`, '200');
if (Array.isArray(json()?.messages)) ok('messages array in response');
else bad('messages response shape');

console.log('\n--- Meeting room page ---');
curlStatus(`${BASE}/m/${slug}`, '200');
const room = readFileSync(TMP, 'utf8');
const roomDev = room.includes('__sveltekit_dev');
if (roomDev) ok('room page is Vite dev mode');
else if (room.includes('href="/_app/')) ok('room page root-relative assets');
else bad('room asset paths');

if (roomDev) {
	const src = curl(['-o', '-', `${BASE}/src/routes/m/%5Bslug%5D/+page.svelte`]).body;
	if (src.includes('Join with camera')) ok('room source has lobby join UI');
	else bad('room source missing lobby join UI');
	if (src.includes('requestLocalMedia') || src.includes('getUserMedia')) ok('room source has media request');
	else bad('room source missing media request');
} else {
const nodeMatch = room.match(/\/_app\/immutable\/nodes\/3\.[^"]+\.js/);
if (nodeMatch) {
	const bundle = curl(['-o', '-', `${BASE}${nodeMatch[0]}`]).body;
	if (bundle.includes('Join with camera')) ok('room bundle has lobby join UI');
	else bad('room bundle missing lobby join UI');
	if (bundle.includes('getUserMedia')) ok('room bundle has getUserMedia');
	else bad('room bundle missing getUserMedia');
} else {
	// Node chunk is loaded dynamically — resolve via app entry
	const appEntry = room.match(/\/_app\/immutable\/entry\/app\.[^"]+\.js/);
	if (appEntry) {
		const appJs = curl(['-o', '-', `${BASE}${appEntry[0]}`]).body;
		const dyn = appJs.match(/nodes\/3\.[^"]+\.js/);
		if (dyn) {
			const bundle = curl(['-o', '-', `${BASE}/_app/immutable/${dyn[0]}`]).body;
			if (bundle.includes('Join with camera')) ok('room bundle has lobby join UI');
			else bad('room bundle missing lobby join UI');
			if (bundle.includes('getUserMedia')) ok('room bundle has getUserMedia');
			else bad('room bundle missing getUserMedia');
		} else bad('could not resolve meeting page JS bundle from app entry');
	} else bad('could not find app entry JS');
}
}

console.log('\n--- WebSocket ---');
try {
	await testWebSocket(slug, 'SmokeTester', (msg, _ws, done) => {
		if (msg.type === 'room.welcome') {
			ok(`WebSocket room.welcome selfId=${msg.payload?.selfId ?? '?'}`);
			done();
		}
	});
} catch (e) {
	bad(`WebSocket welcome: ${e.message}`);
}

console.log('\n--- WebSocket chat round-trip ---');
try {
	await testWebSocket(slug, 'ChatTester', (msg, ws, done) => {
		if (msg.type === 'room.welcome') {
			ws.send(JSON.stringify({ type: 'chat.message', payload: { content: 'smoke hello' } }));
		}
		if (msg.type === 'chat.new' && msg.payload?.content === 'smoke hello') {
			ok(`chat.new received id=${msg.payload.id}`);
			done();
		}
	});
} catch (e) {
	bad(`chat round-trip: ${e.message}`);
}

await new Promise((r) => setTimeout(r, 500));
curlStatus(`${BASE}/api/meetings/${slug}/messages?limit=5`, '200');
if ((json()?.messages ?? []).some((m) => m.content === 'smoke hello')) {
	ok('chat message persisted to REST history');
} else {
	bad('chat message not in REST history');
}

console.log('\n--- 404 cases ---');
curlStatus(`${BASE}/api/meetings/does-not-exist-xyz`, '404');

console.log(`\n=== Results: ${pass} passed, ${fail} failed ===`);
process.exit(fail === 0 ? 0 : 1);
