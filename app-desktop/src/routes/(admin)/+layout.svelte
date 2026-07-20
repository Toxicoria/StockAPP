<script>
  import { onMount } from 'svelte'; import { goto } from '$app/navigation'; import { esAdmin } from '$lib/sesion.svelte.js';
  let { children } = $props();
  onMount(() => { if (!esAdmin()) goto('/'); });
</script>
{#if esAdmin()}{@render children()}{/if}
