<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		variant = 'neutral',
		disabled = false,
		label,
		onclick,
		children
	}: {
		variant?: 'neutral' | 'active' | 'danger' | 'leave';
		disabled?: boolean;
		label: string;
		onclick?: () => void;
		children: Snippet;
	} = $props();
</script>

<button
	type="button"
	{disabled}
	{onclick}
	aria-label={label}
	title={label}
	class={cn(
		'control-btn control-btn--icon',
		variant === 'active' && 'control-btn--active',
		variant === 'danger' && 'control-btn--danger',
		variant === 'leave' && 'control-btn-leave',
		variant === 'neutral' && 'control-btn--neutral'
	)}
>
	{@render children()}
</button>
