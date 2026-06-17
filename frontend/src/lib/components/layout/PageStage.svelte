<script lang="ts">
	import { page } from '$app/stores';

	let { children, fullBleed = false }: { children: import('svelte').Snippet; fullBleed?: boolean } =
		$props();
</script>

{#key $page.url.pathname}
	<div class={fullBleed ? 'animate-page-enter min-h-screen' : 'animate-page-enter'}>
		{@render children()}
	</div>
{/key}
