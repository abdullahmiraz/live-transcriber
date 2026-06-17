<script lang="ts">
	import { Video } from '@lucide/svelte';
	import { cn } from '$lib/utils';

	let {
		class: className = '',
		compact = false,
		inverted = false
	}: { class?: string; compact?: boolean; inverted?: boolean } = $props();
</script>

<a href="/" class={cn('group inline-flex items-center gap-2.5', inverted && 'text-white', className)}>
	<span
		class={cn(
			'flex size-9 items-center justify-center rounded-xl shadow-sm transition-transform duration-200 group-hover:scale-105',
			inverted ? 'bg-white/15 text-white' : 'bg-primary text-primary-foreground'
		)}
	>
		<Video class="size-4.5" />
	</span>
	{#if !compact}
		<span class="flex flex-col leading-none">
			<span class="text-base font-semibold tracking-tight">Live Meet</span>
			<span class={cn('text-[0.7rem] font-medium', inverted ? 'text-white/70' : 'text-muted-foreground')}>
				AI meetings
			</span>
		</span>
	{/if}
</a>
