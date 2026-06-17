<script lang="ts">
	import '../app.css';
	import { ModeWatcher } from 'mode-watcher';
	import { Toaster } from '$lib/components/ui/sonner';
	import PageStage from '$lib/components/layout/PageStage.svelte';

	let { children } = $props();
</script>

<ModeWatcher defaultMode="system" />
<Toaster richColors position="top-center" />

<div class="app-ambient min-h-screen">
	<PageStage>
		{@render children()}
	</PageStage>
</div>
