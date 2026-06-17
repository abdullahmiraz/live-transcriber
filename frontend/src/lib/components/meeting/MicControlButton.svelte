<script lang="ts">
	import { Mic, MicOff } from '@lucide/svelte';
	import { cn } from '$lib/utils';

	let {
		micOn = true,
		level = 0,
		disabled = false,
		onclick
	}: {
		micOn?: boolean;
		level?: number;
		disabled?: boolean;
		onclick?: () => void;
	} = $props();

	let tick = $state(0);
	let raf = 0;

	$effect(() => {
		void level;
		const loop = () => {
			tick = performance.now();
			raf = requestAnimationFrame(loop);
		};
		raf = requestAnimationFrame(loop);
		return () => cancelAnimationFrame(raf);
	});

	const speaking = $derived(micOn && !disabled && level > 0.06);

	function barHeight(index: number): string {
		if (!speaking) return '20%';
		const phase = tick / 140 + index * 0.75;
		const wobble = 0.78 + Math.sin(phase) * 0.22;
		const tier = 0.55 + index * 0.12;
		const h = Math.min(100, level * wobble * tier * 135);
		return `${Math.max(18, h)}%`;
	}
</script>

<button
	type="button"
	{disabled}
	{onclick}
	title={disabled ? 'Joined without a microphone' : micOn ? 'Mute microphone' : 'Unmute microphone'}
	aria-label={micOn ? 'Mute microphone' : 'Unmute microphone'}
	aria-pressed={micOn}
	class={cn(
		'control-btn control-btn--icon control-btn--mic',
		micOn ? 'control-btn--on' : 'control-btn--off',
		speaking && 'control-btn--speaking'
	)}
>
	{#if micOn}
		<Mic class="size-[1.125rem]" />
	{:else}
		<MicOff class="size-[1.125rem]" />
	{/if}

	<!-- Overlay wave — does not affect button width -->
	<span
		class={cn('mic-wave-overlay', (!micOn || !speaking) && 'mic-wave-overlay--idle')}
		aria-hidden="true"
	>
		{#each [0, 1, 2] as i (i)}
			<span class="mic-wave-bar" style="height: {barHeight(i)}"></span>
		{/each}
	</span>
</button>
