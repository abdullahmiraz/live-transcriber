<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { getMeeting, listMessages, type ChatMessage } from '$lib/api';
	import { SignalingClient } from '$lib/realtime/signaling';
	import { MeshManager } from '$lib/realtime/webrtc';
	import { SpeechCapture } from '$lib/realtime/speech';
	import type {
		WelcomePayload,
		TranscriptPayload,
		TranslationPayload,
		PeerInfo
	} from '$lib/realtime/types';

	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import * as Tabs from '$lib/components/ui/tabs';
	import * as Select from '$lib/components/ui/select';
	import * as Avatar from '$lib/components/ui/avatar';
	import { ScrollArea } from '$lib/components/ui/scroll-area';
	import Chat from '$lib/components/Chat.svelte';
	import { toast } from 'svelte-sonner';
	import {
		Mic,
		MicOff,
		Video,
		VideoOff,
		Captions,
		CaptionsOff,
		PhoneOff,
		Copy,
		Check,
		Users
	} from '@lucide/svelte';

	let { data } = $props();
	const slug: string = $derived(data.slug);

	const LANGS = [
		{ code: 'en-US', label: 'English' },
		{ code: 'ru-RU', label: 'Russian' },
		{ code: 'es-ES', label: 'Spanish' },
		{ code: 'fr-FR', label: 'French' },
		{ code: 'de-DE', label: 'German' },
		{ code: 'zh-CN', label: 'Chinese' },
		{ code: 'ja-JP', label: 'Japanese' }
	];

	// reactive UI state
	let displayName = $state('Guest');
	let meetingTitle = $state('Meeting');
	let status = $state<'connecting' | 'open' | 'closed'>('connecting');
	let loadError = $state('');
	let ready = $state(false);

	let micOn = $state(true);
	let camOn = $state(true);
	let captionsOn = $state(false);
	let speechSupported = $state(false);
	let sourceLang = $state('en-US');
	let liveInterim = $state('');
	let copied = $state(false);
	let activeTab = $state('chat');

	type Tile = { id: string; name: string; stream: MediaStream };
	let remoteTiles = $state<Tile[]>([]);

	type Caption = {
		key: string;
		participantId: string;
		name: string;
		original: string;
		translated: string;
	};
	let captions = $state<Caption[]>([]);
	let roster = $state<PeerInfo[]>([]);
	let chatMessages = $state<ChatMessage[]>([]);
	const seenChatIds = new Set<string>();

	const langLabel = $derived(LANGS.find((l) => l.code === sourceLang)?.label ?? 'English');

	// non-reactive refs
	let localStream: MediaStream | null = null;
	let sig: SignalingClient | null = null;
	let mesh: MeshManager | null = null;
	let speech: SpeechCapture | null = null;
	let localVideoEl: HTMLVideoElement | undefined = $state();
	let selfId = $state('');
	let seq = 0;

	function utf8ToBase64(s: string): string {
		return btoa(String.fromCharCode(...new TextEncoder().encode(s)));
	}

	function nameFor(id: string): string {
		if (id === selfId) return `${displayName} (you)`;
		return roster.find((p) => p.id === id)?.name ?? 'Guest';
	}

	function addChatMessage(m: ChatMessage) {
		if (seenChatIds.has(m.id)) return;
		seenChatIds.add(m.id);
		chatMessages = [...chatMessages, m];
	}

	function sendChat(text: string) {
		sig?.send({ type: 'chat.message', payload: { content: text } });
	}

	function upsertTranscript(p: TranscriptPayload) {
		const key = `${p.participantId}:${p.seq}`;
		const name = nameFor(p.participantId);
		const idx = captions.findIndex((c) => c.key === key);
		if (idx >= 0) {
			captions[idx] = { ...captions[idx], original: p.text, name };
			captions = [...captions];
		} else {
			captions = [
				...captions,
				{ key, participantId: p.participantId, name, original: p.text, translated: '' }
			].slice(-80);
		}
	}

	function applyTranslation(p: TranslationPayload) {
		const key = `${p.participantId}:${p.seq}`;
		const idx = captions.findIndex((c) => c.key === key);
		if (idx >= 0) {
			captions[idx] = { ...captions[idx], translated: p.text };
			captions = [...captions];
		}
	}

	// Svelte action: bind a MediaStream to a <video> element.
	function attach(node: HTMLVideoElement, stream: MediaStream) {
		node.srcObject = stream;
		return {
			update(s: MediaStream) {
				node.srcObject = s;
			},
			destroy() {
				node.srcObject = null;
			}
		};
	}

	$effect(() => {
		speech?.setLang(sourceLang);
	});

	onMount(async () => {
		speechSupported = SpeechCapture.isSupported();
		displayName = sessionStorage.getItem('displayName') || 'Guest';

		try {
			const m = await getMeeting(slug);
			meetingTitle = m.title || 'Meeting';
			if (m.status === 'ended') {
				loadError = 'This meeting has already ended.';
				return;
			}
		} catch (e) {
			loadError = (e as Error).message || 'Meeting not found.';
			return;
		}

		// Load chat history (chronological) before subscribing to live events.
		try {
			const history = await listMessages(slug, { limit: 50 });
			for (const m of history) addChatMessage(m);
		} catch {
			// non-fatal; live chat still works
		}

		try {
			localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
		} catch {
			loadError = 'Camera and microphone access is required to join this meeting.';
			return;
		}
		ready = true;
		if (localVideoEl) localVideoEl.srcObject = localStream;

		sig = new SignalingClient(slug, displayName);
		sig.onStatusChange = (s) => (status = s);
		sig.on('room.welcome', (env) => {
			selfId = (env.payload as WelcomePayload).selfId;
		});
		sig.on('transcript.updated', (env) => upsertTranscript(env.payload as TranscriptPayload));
		sig.on('translation.updated', (env) => applyTranslation(env.payload as TranslationPayload));
		sig.on('chat.new', (env) => addChatMessage(env.payload as ChatMessage));
		sig.on('error', (env) => {
			const p = env.payload as { code?: string; message?: string };
			if (p?.code) toast.error(p.message || p.code);
		});

		mesh = new MeshManager(sig, localStream, {
			onRemoteStream: (id, name, stream) => {
				const idx = remoteTiles.findIndex((t) => t.id === id);
				if (idx >= 0) {
					remoteTiles[idx] = { id, name, stream };
					remoteTiles = [...remoteTiles];
				} else {
					remoteTiles = [...remoteTiles, { id, name, stream }];
				}
			},
			onPeerLeft: (id) => {
				remoteTiles = remoteTiles.filter((t) => t.id !== id);
			},
			onRosterChange: (peers) => (roster = peers)
		});

		speech = new SpeechCapture(sourceLang, {
			onText: (text, isFinal) => {
				liveInterim = isFinal ? '' : text;
				if (isFinal && sig) {
					sig.send({
						type: 'speech.received',
						payload: { audio: utf8ToBase64(text), seq: seq++, lang: sourceLang.split('-')[0] }
					});
				}
			},
			onError: (m) => console.warn(m),
			onRunningChange: (r) => {
				if (!r && captionsOn) captionsOn = false;
			}
		});

		sig.connect();
	});

	onDestroy(() => {
		speech?.stop();
		mesh?.close();
		sig?.close();
		localStream?.getTracks().forEach((t) => t.stop());
	});

	function toggleMic() {
		micOn = !micOn;
		localStream?.getAudioTracks().forEach((t) => (t.enabled = micOn));
	}
	function toggleCam() {
		camOn = !camOn;
		localStream?.getVideoTracks().forEach((t) => (t.enabled = camOn));
	}
	function toggleCaptions() {
		if (!speech) return;
		if (captionsOn) {
			speech.stop();
			captionsOn = false;
			liveInterim = '';
		} else {
			speech.setLang(sourceLang);
			speech.start();
			captionsOn = true;
			activeTab = 'captions';
		}
	}
	async function copyLink() {
		try {
			await navigator.clipboard.writeText(location.href);
			copied = true;
			toast.success('Meeting link copied');
			setTimeout(() => (copied = false), 1500);
		} catch {
			toast.error('Could not copy link');
		}
	}
	function leave() {
		goto('/');
	}
</script>

<svelte:head><title>{meetingTitle} · Live Meet</title></svelte:head>

{#if loadError}
	<div class="grid min-h-screen place-items-center p-4">
		<div class="bg-card w-full max-w-sm rounded-xl border p-8 text-center">
			<h2 class="text-lg font-semibold">Can’t join the meeting</h2>
			<p class="text-muted-foreground mt-2 text-sm">{loadError}</p>
			<Button class="mt-5" onclick={leave}>Back to home</Button>
		</div>
	</div>
{:else}
	<div class="flex h-screen flex-col">
		<!-- Top bar -->
		<header class="flex items-center justify-between gap-3 border-b px-4 py-3">
			<div class="flex min-w-0 items-center gap-2">
				<strong class="truncate">{meetingTitle}</strong>
				<Badge variant="secondary" class="font-mono">{slug}</Badge>
			</div>
			<div class="flex items-center gap-2">
				<Badge
					variant={status === 'open' ? 'default' : status === 'closed' ? 'destructive' : 'outline'}
					class="capitalize"
				>
					{status}
				</Badge>
				<Button variant="outline" size="sm" onclick={copyLink}>
					{#if copied}<Check class="size-4" />{:else}<Copy class="size-4" />{/if}
					Copy link
				</Button>
			</div>
		</header>

		<!-- Body -->
		<div class="grid min-h-0 flex-1 grid-cols-1 gap-3 p-3 lg:grid-cols-[1fr_360px]">
			<!-- Video grid -->
			<div
				class="grid auto-rows-fr content-start gap-3 overflow-auto"
				style="grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));"
			>
				<div class="bg-muted relative aspect-video overflow-hidden rounded-xl border">
					<!-- svelte-ignore a11y_media_has_caption -->
					<video bind:this={localVideoEl} autoplay playsinline muted class="h-full w-full -scale-x-100 object-cover"
					></video>
					<span class="absolute bottom-2 left-2 rounded-md bg-black/55 px-2 py-0.5 text-xs text-white">
						{displayName} (you){!camOn ? ' · cam off' : ''}{!micOn ? ' · muted' : ''}
					</span>
				</div>

				{#each remoteTiles as tile (tile.id)}
					<div class="bg-muted relative aspect-video overflow-hidden rounded-xl border">
						<!-- svelte-ignore a11y_media_has_caption -->
						<video use:attach={tile.stream} autoplay playsinline class="h-full w-full object-cover"></video>
						<span
							class="absolute bottom-2 left-2 rounded-md bg-black/55 px-2 py-0.5 text-xs text-white"
						>
							{tile.name}
						</span>
					</div>
				{/each}
			</div>

			<!-- Sidebar: Chat / Captions / People -->
			<aside class="bg-card flex min-h-0 flex-col overflow-hidden rounded-xl border">
				<Tabs.Root bind:value={activeTab} class="flex h-full min-h-0 flex-col gap-0">
					<Tabs.List class="m-2 grid grid-cols-3">
						<Tabs.Trigger value="chat">Chat</Tabs.Trigger>
						<Tabs.Trigger value="captions">Captions</Tabs.Trigger>
						<Tabs.Trigger value="people">People</Tabs.Trigger>
					</Tabs.List>

					<Tabs.Content value="chat" class="mt-0 min-h-0 flex-1 overflow-hidden">
						<Chat messages={chatMessages} {selfId} onSend={sendChat} />
					</Tabs.Content>

					<Tabs.Content value="captions" class="mt-0 flex min-h-0 flex-1 flex-col overflow-hidden">
						<ScrollArea class="min-h-0 flex-1">
							<div class="flex flex-col gap-3 p-4">
								{#if captions.length === 0}
									<p class="text-muted-foreground text-sm">
										Turn on captions and start speaking — live transcription and translation appear
										here.
									</p>
								{/if}
								{#each captions as c (c.key)}
									<div>
										<div class="text-primary text-xs font-semibold">{c.name}</div>
										<div class="text-sm">{c.original}</div>
										{#if c.translated}
											<div class="text-sm text-emerald-400 italic">{c.translated}</div>
										{/if}
									</div>
								{/each}
							</div>
						</ScrollArea>
						{#if liveInterim}
							<div class="text-muted-foreground border-t p-3 text-sm">🎙 {liveInterim}</div>
						{/if}
					</Tabs.Content>

					<Tabs.Content value="people" class="mt-0 min-h-0 flex-1 overflow-hidden">
						<ScrollArea class="h-full">
							<div class="flex flex-col gap-2 p-4">
								<div class="flex items-center gap-2.5">
									<Avatar.Root class="size-8"><Avatar.Fallback class="text-xs">
										{(displayName[0] ?? '?').toUpperCase()}</Avatar.Fallback></Avatar.Root>
									<span class="text-sm">{displayName} <span class="text-muted-foreground">(you)</span></span>
								</div>
								{#each roster as p (p.id)}
									<div class="flex items-center gap-2.5">
										<Avatar.Root class="size-8"><Avatar.Fallback class="text-xs">
											{(p.name[0] ?? '?').toUpperCase()}</Avatar.Fallback></Avatar.Root>
										<span class="text-sm">{p.name}</span>
									</div>
								{/each}
								<p class="text-muted-foreground mt-1 flex items-center gap-1 text-xs">
									<Users class="size-3" /> {roster.length + 1} in call
								</p>
							</div>
						</ScrollArea>
					</Tabs.Content>
				</Tabs.Root>
			</aside>
		</div>

		<!-- Controls -->
		<footer class="flex flex-wrap items-center justify-center gap-2 border-t p-3">
			<Button variant={micOn ? 'secondary' : 'destructive'} onclick={toggleMic}>
				{#if micOn}<Mic class="size-4" /> Mute{:else}<MicOff class="size-4" /> Unmute{/if}
			</Button>
			<Button variant={camOn ? 'secondary' : 'destructive'} onclick={toggleCam}>
				{#if camOn}<Video class="size-4" /> Stop video{:else}<VideoOff class="size-4" /> Start video{/if}
			</Button>
			<Button
				variant={captionsOn ? 'default' : 'secondary'}
				onclick={toggleCaptions}
				disabled={!speechSupported || !ready}
			>
				{#if captionsOn}<Captions class="size-4" /> Captions on{:else}<CaptionsOff class="size-4" /> Captions{/if}
			</Button>

			<Select.Root type="single" bind:value={sourceLang}>
				<Select.Trigger class="w-[140px]">{langLabel}</Select.Trigger>
				<Select.Content>
					{#each LANGS as l (l.code)}
						<Select.Item value={l.code} label={l.label}>{l.label}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>

			<Button variant="destructive" onclick={leave}><PhoneOff class="size-4" /> Leave</Button>
		</footer>
	</div>
{/if}
