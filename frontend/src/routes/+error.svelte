<script>
  import { page } from '$app/stores';
  import { getTranslation, LANGUAGES } from '$lib/i18n/index.js';
  import { onMount } from 'svelte';

  // R58-F3: root error page bypasses [lang] layout's afterNavigate focus
  // management. Without this, screen reader/keyboard users don't know the
  // page changed after a root-level error (e.g. /nonexistent).
  onMount(() => {
    const main = document.getElementById('main-content');
    if (main) main.focus({ preventScroll: true });
  });

  let status = $derived($page.status);
  // $page.params.lang is empty when the 404 occurs at the root error boundary
  // (e.g. /ja/nonexistent-page renders the root +error.svelte, not [lang]'s).
  // Fall back to extracting the lang segment from the URL pathname.
  let lang = $derived.by(() => {
    if ($page.params.lang && LANGUAGES[$page.params.lang]) return $page.params.lang;
    const m = $page.url.pathname.match(/^\/([a-z]{2})(?:\/|$)/);
    if (m && LANGUAGES[m[1]]) return m[1];
    return 'en';
  });
  let t = $derived(getTranslation(lang));

  let title = $derived(
    status === 404 ? (t.error?.notFoundTitle || 'Page Not Found')
    : status === 500 ? (t.error?.serverTitle || 'Server Error')
    : status === 429 ? (t.error?.rateLimitTitle || 'Too Many Requests')
    : status === 503 ? (t.error?.maintenanceTitle || 'Under Maintenance')
    : (t.error?.genericTitle || 'Something Went Wrong')
  );
  let message = $derived(
    status === 404 ? (t.error?.notFoundMsg || 'The page you are looking for does not exist or has been moved.')
    : status === 500 ? (t.error?.serverMsg || 'An internal server error occurred. Please try again later.')
    : status === 429 ? (t.error?.rateLimitMsg || 'You are making requests too fast. Please wait a moment and try again.')
    : status === 503 ? (t.error?.maintenanceMsg || 'The service is temporarily unavailable. Please try again later.')
    : (t.error?.genericMsg || 'An unexpected error occurred.')
  );
</script>

<svelte:head>
  <title>{title} — ResumeTake</title>
  <meta name="robots" content="noindex, nofollow">
</svelte:head>

<div id="main-content" tabindex="-1" style="min-height:60vh;display:flex;align-items:center;justify-content:center;padding:2rem 1rem;outline:none">
  <div style="text-align:center;max-width:28rem">
    <div style="font-size:clamp(3rem,8vw,5rem);font-weight:800;color:var(--primary);line-height:1">{status}</div>
    <h1 style="font-size:clamp(1.25rem,3vw,1.75rem);font-weight:700;margin:1rem 0 0.5rem;color:var(--text)">{title}</h1>
    <p style="color:var(--text-secondary);margin-bottom:2rem">{message}</p>
    <a href="/{lang}" class="btn btn-primary">{t.error?.backHome || 'Back to Home'}</a>
  </div>
</div>
