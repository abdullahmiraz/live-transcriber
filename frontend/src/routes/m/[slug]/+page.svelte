<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getMeeting, listMessages, type ChatMessage } from '$lib/api';
	import { requestLocalMedia, summarizeMediaDevices, mediaErrorMessage } from '$lib/media/request-media';
	import { SignalingClient } from '$lib/realtime/signaling';
	import { MeshManager } from '$lib/realtime/webrtc';
	import { SpeechCapture } from '$lib/realtime/speech';
	import type {
		WelcomePayload,
		TranscriptPayload,
		TranslationPayload,
		PeerInfo,
		WsConnectionStatus
	} from '$lib/realtime/types';
	import {
		LANGUAGES,
		DEFAULT_SOURCE_LANG,
		DEFAULT_TARGET_LANG,
		langBase,
		languageLabel,
		firstLanguageOtherThan
	} from '$lib/config/languages';
	import type { JoinMediaMode, MeetingPhase, MeetingTab, VideoTile, Caption } from '$lib/meeting/types';
	import {
		CHAT_HISTORY_LIMIT,
		CHAT_POLL_INTERVAL_MS,
		MAX_CAPTIONS,
		COPY_LINK_FEEDBACK_MS,
		DEFAULT_MEETING_TITLE,
		DEFAULT_MEETING_TAB,
		EMPTY_ROOM_TTL_MINUTES,
		emptyRoomLeaveMessage
	} from '$lib/meeting/constants';
	import { joinModeToPrefs } from '$lib/meeting/join-media';
	import { getDisplayName, setDisplayName, DEFAULT_GUEST_NAME } from '$lib/meeting/session';
	import { videoGridClassForCount } from '$lib/meeting/video-grid';
	import { utf8ToBase64 } from '$lib/utils/encoding';
	import { cn } from '$lib/utils';

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
	import MicControlButton from '$lib/components/meeting/MicControlButton.svelte';
	import MeetingControlButton from '$lib/components/meeting/MeetingControlButton.svelte';
	import { MicLevelMonitor } from '$lib/media/mic-level';
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
		AlertCircle,
		MessagesSquare
	} from '@lucide/svelte';

	// Slug from the URL — do NOT use load() data (adapter-node SSR sends data: [null,null]).
	const slug = $derived(String(page.params.slug ?? ''));

	// reactive UI state
	let displayName = $state(DEFAULT_GUEST_NAME);
	let meetingTitle = $state(DEFAULT_MEETING_TITLE);
	let phase = $state<MeetingPhase>('loading');
	let status = $state<WsConnectionStatus>('connecting');
	let loadError = $state('');
	let mediaError = $state('');
	let ready = $state(false);
	let joining = $state(false);

	let micOn = $state(true);
	let camOn = $state(true);
	let captionsOn = $state(false);
	let speechSupported = $state(false);
	let sourceLang = $state(DEFAULT_SOURCE_LANG);
	let targetLang = $state(DEFAULT_TARGET_LANG);
	let liveInterim = $state('');
	let copied = $state(false);
	let activeTab = $state<MeetingTab>(DEFAULT_MEETING_TAB);

	let remoteTiles = $state<VideoTile[]>([]);

	let captions = $state<Caption[]>([]);
	let roster = $state<PeerInfo[]>([]);
	let chatMessages = $state<ChatMessage[]>([]);
	const seenChatIds = new Set<string>();

	const langLabel = $derived(languageLabel(sourceLang));
	const targetLabel = $derived(languageLabel(targetLang));

	// media + realtime (localStream must be $state for lobby preview)
	let localStream = $state<MediaStream | null>(null);
	const hasLocalVideo = $derived(Boolean(localStream?.getVideoTracks().length));
	const hasLocalAudio = $derived(Boolean(localStream?.getAudioTracks().length));
	let sig: SignalingClient | null = null;
	let mesh: MeshManager | null = null;
	let speech: SpeechCapture | null = null;
	let localVideoEl: HTMLVideoElement | undefined = $state();
	let previewVideoEl: HTMLVideoElement | undefined = $state();
	let selfId = $state('');
	let seq = 0;
	let deviceHint = $state('');
	let chatPollTimer: ReturnType<typeof setInterval> | null = null;
	let visibilityHandler: (() => void) | null = null;
	let micLevel = $state(0);

	const chatConnected = $derived(status === 'open');
	const videoTileCount = $derived(1 + remoteTiles.length);
	const videoGridClass = $derived.by(() => videoGridClassForCount(videoTileCount));

	function sidebarTabClass(tab: MeetingTab): string {
		return cn(
			'h-9 w-full rounded-md text-sm font-medium transition-all',
			activeTab === tab
				? 'bg-primary text-primary-foreground shadow-sm font-semibold'
				: 'text-muted-foreground hover:bg-muted/70 hover:text-foreground'
		);
	}

	async function requestMedia(mode: JoinMediaMode) {
		const result = await requestLocalMedia(joinModeToPrefs(mode));
		camOn = result.hasVideo;
		micOn = result.hasAudio;
		if (result.note) toast.info(result.note);
		return result.stream;
	}

	function nameFor(id: string): string {
		if (id === selfId) return `${displayName} (you)`;
		return roster.find((p) => p.id === id)?.name ?? DEFAULT_GUEST_NAME;
	}

	function addChatMessage(m: ChatMessage) {
		if (seenChatIds.has(m.id)) return;
		seenChatIds.add(m.id);
		chatMessages = [...chatMessages, m];
	}

	function sendChat(text: string) {
		if (!sig) return;
		const sent = sig.send({ type: 'chat.message', payload: { content: text } });
		if (!sent && !sig.isOpen() && !sig.isConnecting()) {
			toast.error('Chat is disconnected — reconnecting…');
			sig.reconnectIfNeeded();
		}
	}

	async function syncChatHistory() {
		if (!slug) return;
		try {
			const history = await listMessages(slug, { limit: CHAT_HISTORY_LIMIT });
			for (const m of history) addChatMessage(m);
		} catch {
			// non-fatal
		}
	}

	function startChatSync() {
		stopChatSync();
		// Fallback for mobile browsers that suspend WebSockets (background tab / flaky LAN).
		chatPollTimer = setInterval(() => {
			if (phase === 'in-call') void syncChatHistory();
		}, CHAT_POLL_INTERVAL_MS);
		visibilityHandler = () => {
			if (document.visibilityState !== 'visible' || phase !== 'in-call') return;
			sig?.reconnectIfNeeded();
			void syncChatHistory();
		};
		document.addEventListener('visibilitychange', visibilityHandler);
	}

	function stopChatSync() {
		if (chatPollTimer) {
			clearInterval(chatPollTimer);
			chatPollTimer = null;
		}
		if (visibilityHandler) {
			document.removeEventListener('visibilitychange', visibilityHandler);
			visibilityHandler = null;
		}
	}

	function upsertTranscript(p: TranscriptPayload, from?: string) {
		const participantId = p.participantId || from || '';
		const key = `${participantId}:${p.seq}`;
		const name = nameFor(participantId);
		const idx = captions.findIndex((c) => c.key === key);
		if (idx >= 0) {
			captions[idx] = { ...captions[idx], original: p.text, name };
			captions = [...captions];
		} else {
			captions = [
				...captions,
				{ key, participantId, name, original: p.text, translated: '' }
			].slice(-MAX_CAPTIONS);
		}
	}

	function applyTranslation(p: TranslationPayload, from?: string) {
		const participantId = p.participantId || from || '';
		const key = `${participantId}:${p.seq}`;
		const idx = captions.findIndex((c) => c.key === key);
		if (idx >= 0) {
			captions[idx] = { ...captions[idx], translated: p.text };
			captions = [...captions];
		} else {
			captions = [
				...captions,
				{ key, participantId, name: nameFor(participantId), original: '', translated: p.text }
			].slice(-MAX_CAPTIONS);
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

	$effect(() => {
		if (langBase(sourceLang) === langBase(targetLang)) {
			const alt = firstLanguageOtherThan(sourceLang);
			if (alt) targetLang = alt.code;
		}
	});

	// Keep preview / in-call video elements in sync with the local stream.
	$effect(() => {
		if (localStream && previewVideoEl) previewVideoEl.srcObject = localStream;
		if (localStream && localVideoEl) localVideoEl.srcObject = localStream;
	});

	$effect(() => {
		if (phase !== 'in-call' || !localStream) {
			micLevel = 0;
			return;
		}
		const monitor = new MicLevelMonitor((l) => {
			micLevel = l;
		});
		monitor.attach(localStream, micOn && hasLocalAudio);
		return () => monitor.detach();
	});

	onMount(async () => {
		speechSupported = SpeechCapture.isSupported();
		displayName = getDisplayName();

		if (!slug) {
			loadError = 'Invalid meeting link.';
			phase = 'error';
			return;
		}

		try {
			const m = await getMeeting(slug);
			meetingTitle = m.title || DEFAULT_MEETING_TITLE;
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

		const { cameras, microphones } = await summarizeMediaDevices();
		if (cameras > 0 || microphones > 0) {
			const parts: string[] = [];
			if (microphones > 0) parts.push(`${microphones} microphone${microphones > 1 ? 's' : ''}`);
			if (cameras > 0) parts.push(`${cameras} camera${cameras > 1 ? 's' : ''}`);
			deviceHint = `Detected: ${parts.join(', ')}. Choose how you want to join below.`;
		} else {
			deviceHint =
				'You can join with camera, microphone, both, or neither (chat only). Device names appear after you allow access.';
		}
	});

	function joinMeeting(mode: JoinMediaMode = 'both') {
		if (joining || phase === 'in-call' || !slug) return;

		// getUserMedia must run in the same turn as the click when devices are requested.
		const mediaPromise = requestMedia(mode);

		joining = true;
		mediaError = '';
		phase = 'joining';

		setDisplayName(displayName);
		displayName = getDisplayName();

		mediaPromise
			.then(async (stream) => {
				localStream = stream;
				try {
					await syncChatHistory();
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
		const stream = localStream ?? new MediaStream();

		sig = new SignalingClient(slug, displayName);
		sig.onStatusChange = (s) => (status = s);
		sig.onOpen = () => {
			void syncChatHistory();
		};
		sig.on('room.welcome', (env) => {
			selfId = (env.payload as WelcomePayload).selfId;
		});
		sig.on('transcript.updated', (env) => upsertTranscript(env.payload as TranscriptPayload, env.from));
		sig.on('translation.updated', (env) => applyTranslation(env.payload as TranslationPayload, env.from));
		sig.on('chat.new', (env) => addChatMessage(env.payload as ChatMessage));
		sig.on('error', (env) => {
			const p = env.payload as { code?: string; message?: string };
			if (p?.code) toast.error(p.message || p.code);
		});

		mesh = new MeshManager(sig, stream, {
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
						payload: {
							audio: utf8ToBase64(text),
							seq: seq++,
							lang: langBase(sourceLang),
							targetLang: langBase(targetLang)
						}
					});
				}
			},
			onError: (m) => console.warn(m),
			onRunningChange: (r) => {
				if (!r && captionsOn) captionsOn = false;
			}
		});

		sig.connect();
		startChatSync();
	}

	onDestroy(() => {
		stopChatSync();
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
			setTimeout(() => (copied = false), COPY_LINK_FEEDBACK_MS);
		} catch {
			toast.error('Could not copy link');
		}
	}
	function leave() {
		if (!confirm(`Leave meeting?\n\n${emptyRoomLeaveMessage()}`)) return;
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
						Check your name, then choose how to join. You can use camera, microphone, both, or
						neither (chat and view only).
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

					{#if deviceHint}
						<p class="text-muted-foreground text-xs leading-relaxed">{deviceHint}</p>
					{/if}

					{#if mediaError}
						<p class="text-destructive rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2 text-sm leading-relaxed">
							{mediaError}
						</p>
					{/if}
				</Card.Content>
				<Card.Footer class="flex flex-col gap-3 pt-2">
					<Button
						class="touch-target w-full gap-2"
						onclick={() => joinMeeting('both')}
						disabled={joining}
						size="lg"
					>
						{#if joining}
							<Loader2 class="size-4 animate-spin" />
							Joining…
						{:else}
							<Video class="size-4" />
							Join with camera &amp; microphone
						{/if}
					</Button>

					<div class="grid grid-cols-1 gap-2 sm:grid-cols-3">
						<Button
							variant="secondary"
							class="touch-target gap-2"
							onclick={() => joinMeeting('audio')}
							disabled={joining}
						>
							<Mic class="size-4" />
							Mic only
						</Button>
						<Button
							variant="secondary"
							class="touch-target gap-2"
							onclick={() => joinMeeting('video')}
							disabled={joining}
						>
							<Video class="size-4" />
							Camera only
						</Button>
						<Button
							variant="outline"
							class="touch-target gap-2"
							onclick={() => joinMeeting('none')}
							disabled={joining}
						>
							<MessagesSquare class="size-4" />
							Chat only
						</Button>
					</div>

					<Button variant="ghost" class="text-muted-foreground w-full" onclick={leave} disabled={joining}>
						Cancel
					</Button>
				</Card.Footer>
			</Card.Root>
		</div>
	</div>
{:else if phase === 'in-call'}
	<div class="bg-background text-foreground flex h-screen flex-col">
		<!-- Top bar -->
		<header
			class="bg-card/90 flex items-center justify-between gap-3 border-b border-border px-4 py-3 backdrop-blur-sm"
		>
			<div class="flex min-w-0 items-center gap-3">
				<BrandLogo compact />
				<div class="bg-border hidden h-5 w-px sm:block"></div>
				<div class="min-w-0">
					<strong class="block truncate text-sm font-semibold">{meetingTitle}</strong>
					<span class="meeting-code-badge">{slug}</span>
				</div>
			</div>
			<div class="flex items-center gap-2">
				<Badge
					variant={status === 'open' ? 'default' : status === 'closed' ? 'destructive' : 'outline'}
					class="capitalize"
				>
					{status}
				</Badge>
				<span class="text-muted-foreground hidden text-xs md:inline">
					Empty rooms auto-delete after <strong class="text-foreground">{EMPTY_ROOM_TTL_MINUTES} minutes</strong>.
				</span>
				<Button variant="outline" size="sm" onclick={copyLink} class="touch-target gap-1.5">
					{#if copied}<Check class="size-4" />{:else}<Copy class="size-4" />{/if}
					<span class="hidden sm:inline"> Copy link</span>
				</Button>
				<ThemeToggle />
			</div>
		</header>

		<!-- Body -->
		<div class="flex min-h-0 flex-1 flex-col gap-2 overflow-hidden p-2 sm:gap-3 sm:p-3 lg:grid lg:grid-cols-[1fr_380px] lg:grid-rows-1 lg:overflow-hidden">
			<!-- Video stage: fills column height, tiles scale — no scroll -->
			<div class="video-stage bg-video-surface max-lg:max-h-[42vh] max-lg:shrink-0 rounded-2xl lg:h-full lg:min-h-0 lg:max-h-none">
				<div class="video-grid {videoGridClass}">
					<div class="video-tile ring-1 ring-white/10">
						{#if hasLocalVideo && camOn}
							<!-- svelte-ignore a11y_media_has_caption -->
							<video
								bind:this={localVideoEl}
								autoplay
								playsinline
								muted
								class="-scale-x-100"
							></video>
						{:else}
							<div
								class="video-tile-placeholder flex flex-col items-center justify-center gap-2 p-6 text-center text-white/80"
							>
								<span
									class="flex size-16 items-center justify-center rounded-full bg-white/10 text-lg font-semibold text-white"
								>
									{(displayName[0] ?? '?').toUpperCase()}
								</span>
								<p class="text-xs text-white/70">{!hasLocalVideo ? 'No camera' : 'Camera off'}</p>
							</div>
						{/if}
						<span
							class="absolute bottom-3 left-3 z-10 max-w-[calc(100%-1.5rem)] truncate rounded-full bg-black/60 px-2.5 py-1 text-xs font-medium text-white backdrop-blur-sm"
						>
							{displayName} (you){!camOn || !hasLocalVideo ? ' · cam off' : ''}{!micOn || !hasLocalAudio
								? ' · muted'
								: ''}
						</span>
					</div>

					{#each remoteTiles as tile (tile.id)}
						<div class="video-tile animate-fade-in ring-1 ring-white/10">
							<!-- svelte-ignore a11y_media_has_caption -->
							<video use:attach={tile.stream} autoplay playsinline></video>
							<span
								class="absolute bottom-3 left-3 z-10 max-w-[calc(100%-1.5rem)] truncate rounded-full bg-black/60 px-2.5 py-1 text-xs font-medium text-white backdrop-blur-sm"
							>
								{tile.name}
							</span>
						</div>
					{/each}
				</div>
			</div>

			<!-- Sidebar: Chat / Captions / People -->
			<aside
				class="meeting-sidebar bg-card text-card-foreground rounded-2xl border border-border shadow-xl lg:max-h-none"
			>
				<Tabs.Root bind:value={activeTab} class="meeting-sidebar-tabs">
					<Tabs.List
						class="mx-2 mt-2 mb-1 grid h-auto min-h-10 w-[calc(100%-1rem)] grid-cols-3 gap-1 bg-muted p-1 shadow-inner"
					>
						<Tabs.Trigger value="chat" class={sidebarTabClass('chat')}>Chat</Tabs.Trigger>
						<Tabs.Trigger value="captions" class={sidebarTabClass('captions')}>Captions</Tabs.Trigger>
						<Tabs.Trigger value="people" class={sidebarTabClass('people')}>People</Tabs.Trigger>
					</Tabs.List>

					<div class="meeting-sidebar-tabs__panels">
						<Tabs.Content value="chat" class="meeting-sidebar-tabs__panel">
							<Chat messages={chatMessages} {selfId} onSend={sendChat} connected={chatConnected} />
						</Tabs.Content>

						<Tabs.Content value="captions" class="meeting-sidebar-tabs__panel">
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
								<div class="text-muted-foreground animate-pulse-soft shrink-0 border-t p-3 text-sm">
									🎙 {liveInterim}
								</div>
							{/if}
						</Tabs.Content>

						<Tabs.Content value="people" class="meeting-sidebar-tabs__panel">
							<ScrollArea class="min-h-0 flex-1">
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
					</div>
				</Tabs.Root>
			</aside>
		</div>

		<!-- Controls -->
		<footer class="meeting-controls-footer">
			<div class="meeting-control-bar">
				<div class="control-group">
					<MicControlButton
						micOn={micOn}
						level={micLevel}
						disabled={!hasLocalAudio}
						onclick={toggleMic}
					/>
					<MeetingControlButton
						variant={camOn ? 'neutral' : 'danger'}
						disabled={!hasLocalVideo}
						label={camOn ? 'Stop camera' : 'Start camera'}
						onclick={toggleCam}
					>
						{#if camOn}<Video class="size-[1.125rem]" />{:else}<VideoOff class="size-[1.125rem]" />{/if}
					</MeetingControlButton>
				</div>

				<div class="control-divider" aria-hidden="true"></div>

				<div class="control-group">
					<MeetingControlButton
						variant={captionsOn ? 'active' : 'neutral'}
						disabled={!speechSupported || !ready || !hasLocalAudio}
						label={captionsOn ? 'Turn off captions' : 'Turn on captions'}
						onclick={toggleCaptions}
					>
						{#if captionsOn}<Captions class="size-[1.125rem]" />{:else}<CaptionsOff class="size-[1.125rem]" />{/if}
					</MeetingControlButton>

					<Select.Root type="single" bind:value={sourceLang}>
						<Select.Trigger class="control-lang" title="Language you speak">{langLabel}</Select.Trigger>
						<Select.Content>
							{#each LANGUAGES as l (l.code)}
								<Select.Item value={l.code} label={l.label}>{l.label}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>

					<span class="text-muted-foreground hidden text-xs sm:inline" aria-hidden="true">→</span>

					<Select.Root type="single" bind:value={targetLang}>
						<Select.Trigger class="control-lang" title="Translate captions to">{targetLabel}</Select.Trigger>
						<Select.Content>
							{#each LANGUAGES as l (l.code)}
								<Select.Item value={l.code} label={l.label}>{l.label}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>

				<div class="control-divider" aria-hidden="true"></div>

				<MeetingControlButton variant="leave" label="Leave meeting" onclick={leave}>
					<PhoneOff class="size-[1.125rem]" />
				</MeetingControlButton>
			</div>
		</footer>
	</div>
{/if}
