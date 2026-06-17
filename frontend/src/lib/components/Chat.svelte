<script lang="ts">
	import { ScrollArea } from '$lib/components/ui/scroll-area';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import * as Avatar from '$lib/components/ui/avatar';
	import { Send } from '@lucide/svelte';
	import type { ChatMessage } from '$lib/api';

	let {
		messages,
		selfId,
		onSend
	}: { messages: ChatMessage[]; selfId: string; onSend: (text: string) => void } = $props();

	let draft = $state('');
	let viewport = $state<HTMLElement | null>(null);

	function submit(e?: SubmitEvent) {
		e?.preventDefault();
		const t = draft.trim();
		if (!t) return;
		onSend(t);
		draft = '';
	}

	function initials(name: string): string {
		const parts = name.trim().split(/\s+/);
		return ((parts[0]?.[0] ?? '') + (parts[1]?.[0] ?? '')).toUpperCase() || '?';
	}

	function fmtTime(iso: string): string {
		try {
			return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
		} catch {
			return '';
		}
	}

	// Auto-scroll to the newest message.
	$effect(() => {
		void messages.length;
		requestAnimationFrame(() => {
			if (viewport) viewport.scrollTop = viewport.scrollHeight;
		});
	});
</script>

<div class="flex h-full min-h-0 flex-col">
	<ScrollArea class="min-h-0 flex-1" bind:viewportRef={viewport}>
		<div class="flex flex-col gap-3 p-4">
			{#if messages.length === 0}
				<p class="text-muted-foreground text-sm">No messages yet. Say hello 👋</p>
			{/if}
			{#each messages as m (m.id)}
				<div class="flex gap-2.5">
					<Avatar.Root class="size-8 shrink-0">
						<Avatar.Fallback class="text-xs">{initials(m.senderName)}</Avatar.Fallback>
					</Avatar.Root>
					<div class="min-w-0 flex-1">
						<div class="flex items-baseline gap-2">
							<span
								class="text-sm font-medium {m.senderId === selfId ? 'text-primary' : ''}"
							>
								{m.senderId === selfId ? 'You' : m.senderName}
							</span>
							<span class="text-muted-foreground text-xs">{fmtTime(m.createdAt)}</span>
						</div>
						<p class="text-sm break-words whitespace-pre-wrap">{m.content}</p>
					</div>
				</div>
			{/each}
		</div>
	</ScrollArea>

	<form class="flex items-center gap-2 border-t p-3" onsubmit={submit}>
		<Input placeholder="Message…" bind:value={draft} autocomplete="off" maxlength={4000} />
		<Button type="submit" size="icon" disabled={!draft.trim()} aria-label="Send message">
			<Send class="size-4" />
		</Button>
	</form>
</div>
