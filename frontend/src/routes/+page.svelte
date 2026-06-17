<script lang="ts">
	import { goto } from '$app/navigation';
	import { createMeeting } from '$lib/api';
	import { meetingPath, parseMeetingSlug } from '$lib/meeting/routes';
	import { setDisplayName, DEFAULT_GUEST_NAME, DEFAULT_HOST_NAME } from '$lib/meeting/session';
	import AppHeader from '$lib/components/layout/AppHeader.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Separator } from '$lib/components/ui/separator';
	import * as Card from '$lib/components/ui/card';
	import { toast } from 'svelte-sonner';
	import { Video, MessagesSquare, Languages, Captions, ArrowRight, Sparkles } from '@lucide/svelte';

	let name = $state('');
	let title = $state('');
	let joinCode = $state('');
	let creating = $state(false);

	const features = [
		{ icon: Video, label: 'Browser video & audio', detail: 'No installs — join from any modern browser' },
		{ icon: MessagesSquare, label: 'Realtime chat', detail: 'Redis-backed messaging in every room' },
		{ icon: Captions, label: 'Live captions', detail: 'Speech-to-text as people speak' },
		{ icon: Languages, label: 'Translated subtitles', detail: 'Follow along in your language' }
	];

	async function handleCreate() {
		creating = true;
		try {
			const m = await createMeeting({ title, host_name: name });
			setDisplayName(name, DEFAULT_HOST_NAME);
			window.location.assign(meetingPath(m.slug));
		} catch (e) {
			toast.error((e as Error).message || 'Could not create meeting');
		} finally {
			creating = false;
		}
	}

	function handleJoin() {
		const slug = parseMeetingSlug(joinCode);
		if (!slug) {
			toast.error('Enter a meeting code or link');
			return;
		}
		setDisplayName(name, DEFAULT_GUEST_NAME);
		window.location.assign(meetingPath(slug));
	}
</script>

<svelte:head>
	<title>Live Meet — Real-time AI meetings</title>
</svelte:head>

<div class="mx-auto flex min-h-screen max-w-6xl flex-col px-5 pb-8">
	<AppHeader>
		<span
			class="text-muted-foreground hidden items-center gap-1.5 text-xs font-medium sm:inline-flex"
		>
			<Sparkles class="text-primary size-3.5" />
			MVP · WebRTC · Redis realtime
		</span>
	</AppHeader>

	<main class="grid flex-1 items-center gap-8 py-6 lg:grid-cols-[1.05fr_0.95fr] lg:gap-16 lg:py-12">
		<section class="animate-fade-in-up order-2 max-w-xl lg:order-1">
			<p
				class="text-primary mb-4 inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/8 px-3 py-1 text-xs font-semibold tracking-wide uppercase"
			>
				Real-time AI meetings
			</p>
			<h1 class="text-4xl leading-[1.1] font-bold tracking-tight md:text-5xl lg:text-[3.25rem]">
				Meet, chat, and read along
				<span class="text-gradient-brand mt-1 block">in real time.</span>
			</h1>
			<p class="text-muted-foreground mt-5 max-w-prose text-base leading-relaxed md:text-lg">
				Spin up a room, share the link, and get video, live chat, captions, and automatic
				translation — powered by WebRTC and a pluggable AI pipeline.
			</p>

			<ul class="animate-stagger mt-8 grid gap-3 sm:grid-cols-2">
				{#each features as f (f.label)}
					<li
						class="surface-card flex gap-3 p-3.5 transition-colors duration-200 hover:border-primary/25"
					>
						<span
							class="bg-primary/10 text-primary flex size-9 shrink-0 items-center justify-center rounded-lg"
						>
							<f.icon class="size-4" />
						</span>
						<div class="min-w-0">
							<p class="text-sm font-semibold">{f.label}</p>
							<p class="text-muted-foreground mt-0.5 text-xs leading-relaxed">{f.detail}</p>
						</div>
					</li>
				{/each}
			</ul>
		</section>

		<Card.Root class="surface-card animate-scale-in order-1 w-full border-0 shadow-none lg:order-2">
			<Card.Header class="pb-2">
				<Card.Title class="text-xl">Get started</Card.Title>
				<Card.Description>Create a new meeting or join with a code.</Card.Description>
			</Card.Header>
			<Card.Content class="grid gap-6 pt-2">
				<div class="grid gap-2">
					<Label for="name">Your name</Label>
					<Input id="name" placeholder="e.g. Alex" bind:value={name} autocomplete="name" />
				</div>

				<div class="grid gap-2">
					<Label for="title">Meeting title <span class="text-muted-foreground font-normal">(optional)</span></Label>
					<Input id="title" placeholder="Team sync" bind:value={title} />
					<Button class="mt-2 gap-2" onclick={handleCreate} disabled={creating} size="lg">
						{#if creating}
							Creating…
						{:else}
							Create meeting
							<ArrowRight class="size-4" />
						{/if}
					</Button>
				</div>

				<div class="flex items-center gap-3">
					<Separator class="flex-1" />
					<span class="text-muted-foreground text-xs font-medium">or join</span>
					<Separator class="flex-1" />
				</div>

				<div class="grid gap-2">
					<Label for="code">Meeting code or link</Label>
					<Input
						id="code"
						placeholder="abc-defg-hij"
						bind:value={joinCode}
						class="font-mono text-sm"
					/>
					<Button variant="secondary" class="mt-2" onclick={handleJoin} size="lg">
						Join meeting
					</Button>
				</div>
			</Card.Content>
		</Card.Root>
	</main>

	<footer class="text-muted-foreground space-y-1 border-t pt-6 text-center text-xs">
		<p>On your phone (same Wi‑Fi): open <strong>https://&lt;your-pc-ip&gt;/</strong> — accept the certificate warning once.</p>
		<p>Swappable STT &amp; translation · Postgres source of truth · Redis fan-out</p>
	</footer>
</div>
