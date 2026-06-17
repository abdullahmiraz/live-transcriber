<script lang="ts">
	import { goto } from '$app/navigation';
	import { createMeeting } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Separator } from '$lib/components/ui/separator';
	import * as Card from '$lib/components/ui/card';
	import { toast } from 'svelte-sonner';
	import { Video, MessagesSquare, Languages, Captions } from '@lucide/svelte';

	let name = $state('');
	let title = $state('');
	let joinCode = $state('');
	let creating = $state(false);

	async function handleCreate() {
		creating = true;
		try {
			const m = await createMeeting({ title, host_name: name });
			sessionStorage.setItem('displayName', name.trim() || 'Host');
			await goto(`/m/${m.slug}`);
		} catch (e) {
			toast.error((e as Error).message || 'Could not create meeting');
		} finally {
			creating = false;
		}
	}

	function handleJoin() {
		const raw = joinCode.trim();
		if (!raw) {
			toast.error('Enter a meeting code or link');
			return;
		}
		const slug = raw.includes('/m/') ? raw.split('/m/')[1].split(/[?#]/)[0] : raw;
		sessionStorage.setItem('displayName', name.trim() || 'Guest');
		goto(`/m/${slug}`);
	}
</script>

<div class="mx-auto flex min-h-screen max-w-6xl flex-col px-5">
	<header class="flex items-center justify-between py-5">
		<div class="flex items-center gap-2 font-semibold">
			<span class="bg-primary flex size-7 items-center justify-center rounded-md">
				<Video class="size-4 text-white" />
			</span>
			Live Meet
		</div>
		<span class="text-muted-foreground text-sm">AI meeting platform</span>
	</header>

	<main class="grid flex-1 items-center gap-10 py-8 md:grid-cols-2">
		<section>
			<h1 class="text-4xl leading-tight font-bold tracking-tight md:text-5xl">
				Meet, chat, and read along
				<span
					class="from-primary block bg-gradient-to-r to-emerald-400 bg-clip-text text-transparent"
				>
					in real time.
				</span>
			</h1>
			<p class="text-muted-foreground mt-4 max-w-prose text-lg">
				Spin up a room, share the link, and get video, realtime text chat, live captions, and
				automatic translation — powered by WebRTC and a pluggable AI pipeline.
			</p>
			<ul class="mt-6 grid gap-3 text-sm">
				<li class="flex items-center gap-2"><Video class="text-primary size-4" /> Browser video & audio — no installs</li>
				<li class="flex items-center gap-2"><MessagesSquare class="text-primary size-4" /> Realtime text chat (Redis-backed)</li>
				<li class="flex items-center gap-2"><Captions class="text-primary size-4" /> Live speech-to-text captions</li>
				<li class="flex items-center gap-2"><Languages class="text-primary size-4" /> Real-time translated subtitles</li>
			</ul>
		</section>

		<Card.Root class="w-full">
			<Card.Header>
				<Card.Title>Get started</Card.Title>
				<Card.Description>Create a new meeting or join an existing one.</Card.Description>
			</Card.Header>
			<Card.Content class="grid gap-5">
				<div class="grid gap-2">
					<Label for="name">Your name</Label>
					<Input id="name" placeholder="e.g. Alex" bind:value={name} />
				</div>

				<div class="grid gap-2">
					<Label for="title">Meeting title (optional)</Label>
					<Input id="title" placeholder="Team sync" bind:value={title} />
					<Button class="mt-1" onclick={handleCreate} disabled={creating}>
						{creating ? 'Creating…' : 'Create meeting'}
					</Button>
				</div>

				<div class="flex items-center gap-3">
					<Separator class="flex-1" />
					<span class="text-muted-foreground text-xs">or</span>
					<Separator class="flex-1" />
				</div>

				<div class="grid gap-2">
					<Label for="code">Meeting code or link</Label>
					<Input id="code" placeholder="abc-defg-hij" bind:value={joinCode} />
					<Button variant="secondary" class="mt-1" onclick={handleJoin}>Join meeting</Button>
				</div>
			</Card.Content>
		</Card.Root>
	</main>

	<footer class="text-muted-foreground py-6 text-center text-sm">
		MVP · WebRTC mesh · Redis realtime · swappable STT &amp; translation
	</footer>
</div>
