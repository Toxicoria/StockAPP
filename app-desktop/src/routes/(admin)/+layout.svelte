<script>
  import { onMount } from 'svelte'; import { goto } from '$app/navigation'; import { esDueno } from '$lib/sesion.svelte.js';
  let { children } = $props();
  onMount(() => { if (!esDueno()) goto('/'); });
</script>
{#if esDueno()}{@render children()}{/if}
