<script>
  import '../lib/styles/themes.css';
  let { children } = $props();
</script>

<svelte:head>
  <meta name="theme-color" content="#2563eb">
</svelte:head>

{@render children()}
