<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
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
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Card from '$lib/components/ui/card';
	import * as Tabs from '$lib/components/ui/tabs';
	import * as Select from '$lib/components/ui/select';
	import * as Avatar from '$lib/components/ui/avatar';
	import { ScrollArea } from '$lib/components/ui/scroll-area';
	import Chat from '$lib/components/Chat.svelte';
	import AppHeader from '$lib/components/layout/AppHeader.svelte';
	import BrandLogo from '$lib/components/layout/BrandLogo.svelte';
	import ThemeToggle from '$lib/components/layout/ThemeToggle.svelte';
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
		Users,
		VideoIcon,
		Loader2,
		AlertCircle
	} from '@lucide/svelte';

	type Phase = 'loading' | 'lobby' | 'joining' | 'in-call' | 'error';

	// Slug from the URL — do NOT use load() data (adapter-node SSR sends data: [null,null]).
	const slug = $derived(String(page.params.slug ?? ''));

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
	let phase = $state<Phase>('loading');
	let status = $state<'connecting' | 'open' | 'closed'>('connecting');
	let loadError = $state('');
	let mediaError = $state('');
	let ready = $state(false);
	let joining = $state(false);

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

	// media + realtime (localStream must be $state for lobby preview)
	let localStream = $state<MediaStream | null>(null);
	let sig: SignalingClient | null = null;
	let mesh: MeshManager | null = null;
	let speech: SpeechCapture | null = null;
	let localVideoEl: HTMLVideoElement | undefined = $state();
	let previewVideoEl: HTMLVideoElement | undefined = $state();
	let selfId = $state('');
	let seq = 0;

	function mediaErrorMessage(err: unknown): string {
		const name = err instanceof DOMException ? err.name : '';
		switch (name) {
			case 'NotAllowedError':
			case 'PermissionDeniedError':
				return 'Camera and microphone access was blocked. Click the lock icon in your browser address bar, allow camera and microphone, then try again.';
			case 'NotFoundError':
			case 'DevicesNotFoundError':
				return 'No camera or microphone was found on this device.';
			case 'NotReadableError':
			case 'TrackStartError':
				return 'Your camera or microphone is in use by another application.';
			case 'SecurityError':
				return 'Camera and microphone require a secure connection (HTTPS or localhost).';
			default:
				return err instanceof Error ? err.message : 'Could not access camera or microphone.';
		}
	}

	async function requestMedia(): Promise<MediaStream> {
		if (!navigator.mediaDevices?.getUserMedia) {
			throw new DOMException('Media devices are not available in this browser.', 'NotSupportedError');
		}
		return navigator.mediaDevices.getUserMedia({ video: true, audio: true });
	}

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

	// Keep preview / in-call video elements in sync with the local stream.
	$effect(() => {
		if (localStream && previewVideoEl) previewVideoEl.srcObject = localStream;
		if (localStream && localVideoEl) localVideoEl.srcObject = localStream;
	});

	onMount(async () => {
		speechSupported = SpeechCapture.isSupported();
		displayName = sessionStorage.getItem('displayName') || 'Guest';

		if (!slug) {
			loadError = 'Invalid meeting link.';
			phase = 'error';
			return;
		}

		try {
			const m = await getMeeting(slug);
			meetingTitle = m.title || 'Meeting';
			if (m.status === 'ended') {
				loadError = 'This meeting has already ended.';
				phase = 'error';
				return;
			}
		} catch (e) {
			loadError = (e as Error).message || 'Meeting not found.';
			phase = 'error';
			return;
		}

		phase = 'lobby';
	});

	function joinMeeting() {
		if (joining || phase === 'in-call' || !slug) return;

		// Invoke getUserMedia in the same turn as the click — required for the permission prompt.
		const mediaPromise = requestMedia();

		joining = true;
		mediaError = '';
		phase = 'joining';

		sessionStorage.setItem('displayName', displayName.trim() || 'Guest');
		displayName = displayName.trim() || 'Guest';

		mediaPromise
			.then(async (stream) => {
				localStream = stream;
				try {
					const history = await listMessages(slug, { limit: 50 });
					for (const m of history) addChatMessage(m);
				} catch {
					// non-fatal
				}
				startRealtime();
				ready = true;
				phase = 'in-call';
				joining = false;
			})
			.catch((e) => {
				console.error('getUserMedia failed:', e);
				mediaError = mediaErrorMessage(e);
				phase = 'lobby';
				joining = false;
				toast.error(mediaError);
			});
	}

	function startRealtime() {
		if (!localStream) return;

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
	}

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

{#if phase === 'loading'}
	<div class="grid min-h-screen place-items-center p-4">
		<div class="flex flex-col items-center gap-3 animate-fade-in">
			<Loader2 class="text-primary size-8 animate-spin" />
			<p class="text-muted-foreground text-sm font-medium">Loading meeting…</p>
		</div>
	</div>
{:else if phase === 'error'}
	<div class="mx-auto flex min-h-screen max-w-lg flex-col px-5">
		<AppHeader />
		<div class="grid flex-1 place-items-center pb-12">
			<div class="surface-card w-full max-w-sm p-8 text-center animate-scale-in">
				<span class="bg-destructive/10 text-destructive mx-auto mb-4 flex size-12 items-center justify-center rounded-full">
					<AlertCircle class="size-6" />
				</span>
				<h2 class="text-lg font-semibold tracking-tight">Can’t join the meeting</h2>
				<p class="text-muted-foreground mt-2 text-sm leading-relaxed">{loadError}</p>
				<Button class="mt-6" onclick={leave}>Back to home</Button>
			</div>
		</div>
	</div>
{:else if phase === 'lobby' || phase === 'joining'}
	<div class="mx-auto flex min-h-screen max-w-lg flex-col px-5">
		<AppHeader />
		<div class="grid flex-1 place-items-center pb-12">
			<Card.Root class="surface-card w-full animate-scale-in border-0 shadow-none">
				<Card.Header>
					<Card.Title class="text-xl">{meetingTitle}</Card.Title>
					<Card.Description class="leading-relaxed">
						Check your name, then join. Your browser will ask for camera and microphone access.
					</Card.Description>
				</Card.Header>
				<Card.Content class="grid gap-4">
					<div class="bg-video-surface relative aspect-video overflow-hidden rounded-xl ring-1 ring-border/60">
						{#if localStream}
							<!-- svelte-ignore a11y_media_has_caption -->
							<video
								bind:this={previewVideoEl}
								autoplay
								playsinline
								muted
								class="h-full w-full -scale-x-100 object-cover animate-fade-in"
							></video>
						{:else}
							<div class="text-muted-foreground flex h-full flex-col items-center justify-center gap-3 p-6 text-center">
								<span class="bg-muted/20 flex size-14 items-center justify-center rounded-full">
									<VideoIcon class="size-7 opacity-70" />
								</span>
								<p class="text-sm leading-relaxed">Camera preview appears after you allow access</p>
							</div>
						{/if}
					</div>

					<div class="grid gap-2">
						<Label for="lobby-name">Your name</Label>
						<Input id="lobby-name" bind:value={displayName} autocomplete="name" />
					</div>

					{#if mediaError}
						<p class="text-destructive rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2 text-sm leading-relaxed">
							{mediaError}
						</p>
					{/if}
				</Card.Content>
				<Card.Footer class="flex gap-2 pt-2">
					<Button variant="secondary" onclick={leave}>Cancel</Button>
					<Button class="flex-1 gap-2" onclick={joinMeeting} disabled={joining} size="lg">
						{#if joining}
							<Loader2 class="size-4 animate-spin" />
							Joining…
						{:else}
							<Video class="size-4" />
							Join with camera &amp; microphone
						{/if}
					</Button>
				</Card.Footer>
			</Card.Root>
		</div>
	</div>
{:else if phase === 'in-call'}
	<div class="bg-video-surface flex h-screen flex-col text-white">
		<!-- Top bar -->
		<header class="flex items-center justify-between gap-3 border-b border-white/10 bg-black/20 px-4 py-3 backdrop-blur-sm">
			<div class="flex min-w-0 items-center gap-3">
				<BrandLogo compact inverted />
				<div class="hidden h-5 w-px bg-white/15 sm:block"></div>
				<div class="min-w-0">
					<strong class="block truncate text-sm font-semibold">{meetingTitle}</strong>
					<Badge variant="secondary" class="font-mono mt-0.5 text-[0.65rem]">{slug}</Badge>
				</div>
			</div>
			<div class="flex items-center gap-2">
				<Badge
					variant={status === 'open' ? 'default' : status === 'closed' ? 'destructive' : 'outline'}
					class="capitalize"
				>
					{status}
				</Badge>
				<Button variant="outline" size="sm" onclick={copyLink} class="border-white/15 bg-white/5 text-white hover:bg-white/10 hover:text-white">
					{#if copied}<Check class="size-4" />{:else}<Copy class="size-4" />{/if}
					Copy link
				</Button>
				<span class="hidden sm:inline-flex"><ThemeToggle /></span>
			</div>
		</header>

		<!-- Body -->
		<div class="grid min-h-0 flex-1 grid-cols-1 gap-3 p-3 lg:grid-cols-[1fr_380px]">
			<!-- Video grid -->
			<div
				class="grid auto-rows-fr content-start gap-3 overflow-auto p-1"
				style="grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));"
			>
				<div class="relative aspect-video overflow-hidden rounded-2xl bg-black/40 ring-1 ring-white/10 shadow-lg">
					<!-- svelte-ignore a11y_media_has_caption -->
					<video bind:this={localVideoEl} autoplay playsinline muted class="h-full w-full -scale-x-100 object-cover"
					></video>
					<span class="absolute bottom-3 left-3 rounded-full bg-black/60 px-2.5 py-1 text-xs font-medium backdrop-blur-sm">
						{displayName} (you){!camOn ? ' · cam off' : ''}{!micOn ? ' · muted' : ''}
					</span>
				</div>

				{#each remoteTiles as tile (tile.id)}
					<div class="relative aspect-video overflow-hidden rounded-2xl bg-black/40 ring-1 ring-white/10 shadow-lg animate-fade-in">
						<!-- svelte-ignore a11y_media_has_caption -->
						<video use:attach={tile.stream} autoplay playsinline class="h-full w-full object-cover"></video>
						<span
							class="absolute bottom-3 left-3 rounded-full bg-black/60 px-2.5 py-1 text-xs font-medium backdrop-blur-sm"
						>
							{tile.name}
						</span>
					</div>
				{/each}
			</div>

			<!-- Sidebar: Chat / Captions / People -->
			<aside class="bg-card text-card-foreground flex min-h-0 flex-col overflow-hidden rounded-2xl border border-white/10 shadow-xl">
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
									<p class="text-muted-foreground text-sm leading-relaxed">
										Turn on captions and start speaking — live transcription and translation appear
										here.
									</p>
								{/if}
								{#each captions as c (c.key)}
									<div class="animate-fade-in rounded-xl border bg-muted/40 p-3">
										<div class="text-primary text-xs font-semibold">{c.name}</div>
										<div class="mt-1 text-sm leading-relaxed">{c.original}</div>
										{#if c.translated}
											<div class="text-success mt-1.5 text-sm italic leading-relaxed">{c.translated}</div>
										{/if}
									</div>
								{/each}
							</div>
						</ScrollArea>
						{#if liveInterim}
							<div class="text-muted-foreground animate-pulse-soft border-t p-3 text-sm">
								🎙 {liveInterim}
							</div>
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
		<footer class="flex justify-center border-t border-white/10 bg-black/25 px-4 py-4 backdrop-blur-sm">
			<div class="meeting-control-bar flex flex-wrap items-center justify-center gap-2 px-3 py-2">
				<Button variant={micOn ? 'secondary' : 'destructive'} onclick={toggleMic} class="rounded-full">
					{#if micOn}<Mic class="size-4" /> Mute{:else}<MicOff class="size-4" /> Unmute{/if}
				</Button>
				<Button variant={camOn ? 'secondary' : 'destructive'} onclick={toggleCam} class="rounded-full">
					{#if camOn}<Video class="size-4" /> Stop video{:else}<VideoOff class="size-4" /> Start video{/if}
				</Button>
				<Button
					variant={captionsOn ? 'default' : 'secondary'}
					onclick={toggleCaptions}
					disabled={!speechSupported || !ready}
					class="rounded-full"
				>
					{#if captionsOn}<Captions class="size-4" /> Captions on{:else}<CaptionsOff class="size-4" /> Captions{/if}
				</Button>

				<Select.Root type="single" bind:value={sourceLang}>
					<Select.Trigger class="w-[140px] rounded-full">{langLabel}</Select.Trigger>
					<Select.Content>
						{#each LANGS as l (l.code)}
							<Select.Item value={l.code} label={l.label}>{l.label}</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>

				<Button variant="destructive" onclick={leave} class="rounded-full">
					<PhoneOff class="size-4" /> Leave
				</Button>
			</div>
		</footer>
	</div>
{/if}
