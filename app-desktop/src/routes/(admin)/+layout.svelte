<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { esDueno, sesion } from '$lib/sesion.svelte.js';
  let { children } = $props();
  onMount(() => {
    if (!esDueno()) {
      goto('/');
    } else if (!sesion.perfilCompleto) {
      goto('/setup');
    }
  });
</script>
{#if esDueno() && sesion.perfilCompleto}{@render children()}{/if}
