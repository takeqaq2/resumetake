<script>
  import { page } from '$app/stores';
  import { getTranslation, LANGUAGES } from '$lib/i18n/index.js';

  let lang = $derived(
    $page.params.lang && LANGUAGES[$page.params.lang] ? $page.params.lang : 'en'
  );
  let t = $derived(getTranslation(lang));
  let status = $derived($page.status);

  let title = $derived(
    status === 404 ? (t.error?.notFoundTitle || 'Page Not Found')
    : status === 429 ? (t.error?.rateLimitTitle || 'Too Many Requests')
    : status === 500 ? (t.error?.serverTitle || 'Server Error')
    : status === 503 ? (t.error?.maintenanceTitle || 'Under Maintenance')
    : (t.error?.genericTitle || 'Something Went Wrong')
  );
  let message = $derived(
    status === 404 ? (t.error?.notFoundMsg || 'The page you are looking for does not exist or has been moved.')
    : status === 429 ? (t.error?.rateLimitMsg || 'You are making too many requests. Please wait a moment and try again.')
    : status === 500 ? (t.error?.serverMsg || 'An internal server error occurred. Please try again later.')
    : status === 503 ? (t.error?.maintenanceMsg || 'The service is temporarily unavailable. Please try again later.')
    : (t.error?.genericMsg || 'An unexpected error occurred.')
  );
</script>

<svelte:head>
  <title>{title} — ResumeTake</title>
  <meta name="robots" content="noindex, nofollow">
</svelte:head>

<div tabindex="-1" style="min-height:60vh;display:flex;align-items:center;justify-content:center;padding:2rem 1rem;outline:none">
  <div style="text-align:center;max-width:28rem">
    <div style="font-size:clamp(3rem,8vw,5rem);font-weight:800;color:var(--primary);line-height:1">{status}</div>
    <h1 style="font-size:clamp(1.25rem,3vw,1.75rem);font-weight:700;margin:1rem 0 0.5rem;color:var(--text)">{title}</h1>
    <p style="color:var(--text-secondary);margin-bottom:2rem">{message}</p>
    <a href="/{lang}" class="btn btn-primary">{t.error?.backHome || 'Back to Home'}</a>
  </div>
</div>
