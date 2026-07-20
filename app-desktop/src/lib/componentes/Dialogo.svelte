<script>
  let { abierto = false, ancho = '420px', children } = $props();
</script>

{#if abierto}
  <div class="dialog-backdrop">
    <div class="dialog" style="width:min({ancho},100%)">
      {@render children()}
    </div>
  </div>
{/if}
