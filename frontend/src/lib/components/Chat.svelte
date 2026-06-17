<script lang="ts">
	import { ScrollArea } from '$lib/components/ui/scroll-area';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import * as Avatar from '$lib/components/ui/avatar';
	import { Send } from '@lucide/svelte';
	import { cn } from '$lib/utils';
	import { CHAT_MAX_CONTENT_LENGTH } from '$lib/meeting/constants';
	import type { ChatMessage } from '$lib/api';

	let {
		messages,
		selfId,
		onSend,
		connected = true
	}: {
		messages: ChatMessage[];
		selfId: string;
		onSend: (text: string) => void;
		connected?: boolean;
	} = $props();

	let draft = $state('');
	let viewport = $state<HTMLElement | null>(null);
	let inputEl = $state<HTMLInputElement | null>(null);

	function sendNow() {
		const t = draft.trim();
		if (!t) return;
		onSend(t);
		draft = '';
	}

	function submit(e?: SubmitEvent) {
		e?.preventDefault();
		sendNow();
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			sendNow();
		}
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
			{#if !connected}
				<p class="text-destructive rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2 text-xs leading-relaxed">
					Reconnecting to chat… messages may be delayed until the connection is back.
				</p>
			{/if}

			{#if messages.length === 0}
				<div class="text-muted-foreground flex flex-col items-center justify-center gap-2 py-10 text-center">
					<p class="text-sm font-medium">No messages yet</p>
					<p class="text-xs">Say hello to everyone in the room</p>
				</div>
			{/if}

			{#each messages as m (m.id)}
				{@const mine = m.senderId === selfId}
				<div class={cn('flex gap-2.5', mine && 'flex-row-reverse')}>
					<Avatar.Root class="size-8 shrink-0">
						<Avatar.Fallback class={cn('text-xs', mine && 'bg-primary/15 text-primary')}>
							{initials(m.senderName)}
						</Avatar.Fallback>
					</Avatar.Root>

					<div class={cn('flex max-w-[85%] min-w-0 flex-col gap-1', mine && 'items-end')}>
						<div class={cn('flex items-baseline gap-2', mine && 'flex-row-reverse')}>
							<span class={cn('text-sm font-semibold', mine && 'text-primary')}>
								{mine ? 'You' : m.senderName}
							</span>
							<span class="text-muted-foreground text-[0.65rem]">{fmtTime(m.createdAt)}</span>
						</div>
						<p
							class={cn(
								'rounded-2xl px-3 py-2 text-sm leading-relaxed break-words whitespace-pre-wrap',
								mine
									? 'bg-primary text-primary-foreground rounded-br-md'
									: 'bg-muted text-foreground rounded-bl-md'
							)}
						>
							{m.content}
						</p>
					</div>
				</div>
			{/each}
		</div>
	</ScrollArea>

	<form
		class="bg-card shrink-0 border-t p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]"
		onsubmit={submit}
	>
		<div class="flex items-center gap-2">
			<Input
				bind:ref={inputEl}
				placeholder="Write a message…"
				bind:value={draft}
				autocomplete="off"
				autocapitalize="sentences"
				enterkeyhint="send"
				inputmode="text"
				maxlength={CHAT_MAX_CONTENT_LENGTH}
				class="min-h-11 flex-1 rounded-full text-base"
				onkeydown={onKeydown}
			/>
			<Button
				type="submit"
				size="icon"
				disabled={!draft.trim()}
				aria-label="Send message"
				class="touch-target size-11 shrink-0 rounded-full"
			>
				<Send class="size-4" />
			</Button>
		</div>
	</form>
</div>
