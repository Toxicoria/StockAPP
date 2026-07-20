<script>
  import { goto } from '$app/navigation';
</script>

<button class="btn btn-ghost volver" onclick={() => goto('/')}>
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="m12 19-7-7 7-7"></path><path d="M19 12H5"></path></svg>
  Volver al inicio
</button>

<style>
  .volver {
    margin-bottom: 16px;
    font-size: 14px;
    min-height: 44px;
  }
</style>
