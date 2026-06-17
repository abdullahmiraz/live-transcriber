<script lang="ts">
	import BrandLogo from './BrandLogo.svelte';
	import ThemeToggle from './ThemeToggle.svelte';
	import { cn } from '$lib/utils';

	let {
		class: className = '',
		children
	}: { class?: string; children?: import('svelte').Snippet } = $props();
</script>

<header class={cn('flex items-center justify-between gap-4 py-5', className)}>
	<BrandLogo />
	<div class="flex items-center gap-2">
		{@render children?.()}
		<ThemeToggle />
	</div>
</header>
