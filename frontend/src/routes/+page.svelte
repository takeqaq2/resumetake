<script>
  import { onMount } from 'svelte';
  onMount(() => { window.location.href = '/en'; });
</script>

<svelte:head>
  <meta http-equiv="refresh" content="0;url=/en">
</svelte:head>

<div style="display:flex;align-items:center;justify-content:center;min-height:100vh;color:var(--text-secondary)">
  <p>Redirecting to English version...</p>
</div>
