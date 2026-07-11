<script>
  import { browser } from '$app/environment';
  import { env } from '$env/dynamic/public';
  import { onMount, tick } from 'svelte';

  let {
    slot,
    format = 'auto',
    layout = '',
    responsive = true,
    label = 'Advertisement'
  } = $props();

  const client = env.PUBLIC_GOOGLE_ADSENSE_CLIENT || '';
  const testMode = env.PUBLIC_ADS_TEST_MODE === 'true';
  const adsEnabled = env.PUBLIC_ADS_ENABLED === 'true' && /^ca-pub-\d+$/.test(client) && /^\d+$/.test(slot || '');

  // GDPR: respect cookie consent — don't push ads if user declined.
  // Default false — onMount checks localStorage; never render ad markup
  // before consent is confirmed (prevents tracking pixel flash).
  let consentGranted = $state(false);

  // R45-F1: consent must be checked for expiry — CookieNotice uses a 6-month
  // expiry. AdSenseScript checks this (R51-F5), but AdSlot did not, causing
  // inconsistency: expired consent still rendered ad markup here. Align with
  // AdSenseScript so both gates behave identically.
  const CONSENT_EXPIRY_MS = 180 * 24 * 60 * 60 * 1000; // 6 months

  function checkConsent() {
    try {
      const raw = localStorage.getItem('cookie_notice_accepted');
      // R49-F1: legacy 'true' format has no timestamp — treat as invalid
      // (same as CookieNotice does) to prevent loading ads before the user
      // re-confirms consent. Previously this returned true, loading AdSense
      // while CookieNotice simultaneously re-shows the consent prompt.
      if (raw === 'true') return false;
      if (raw === 'declined') return false;
      const consent = JSON.parse(raw);
      if (consent?.status !== 'true') return false;
      if (consent.ts && Date.now() - consent.ts > CONSENT_EXPIRY_MS) return false;
      return true;
    } catch {
      return false;
    }
  }

  onMount(() => {
    if (!adsEnabled || !browser) return;
    consentGranted = checkConsent();
    // Always listen for consent changes — if consent was expired (returned
    // false) and user re-accepts via the re-shown CookieNotice, we need to
    // push ads without requiring a page refresh. Previously this listener was
    // only registered when consent was false AND not via the early-return
    // path, missing the expired-consent case.
    const onConsent = (e) => {
      if (e.detail === 'accepted') {
        if (!consentGranted) {
          consentGranted = true;
        }
      } else if (e.detail === 'declined') {
        consentGranted = false;
      }
    };
    window.addEventListener('cookie:consent', onConsent);
    return () => window.removeEventListener('cookie:consent', onConsent);
  });

  // Fix 2: push adsbygoogle in a $effect that depends on consentGranted and
  // waits for tick() so the <ins> element is in the DOM on SPA navigation.
  // Previously the push ran in onMount before the ins was rendered.
  // R50-F1: removed the `typeof window.adsbygoogle !== 'undefined'` guard —
  // it blocked the standard AdSense queue pattern. When consent is granted
  // before the AdSense script finishes loading, the guard prevented push()
  // entirely, and no reactive dependency would re-trigger $effect once the
  // script loaded. The queue pattern `(window.adsbygoogle = window.adsbygoogle
  // || []).push({})` creates the array if missing; the script processes the
  // queue automatically once loaded.
  $effect(() => {
    if (consentGranted) {
      // R49-F4: add cancelled flag so rapid consent toggles (accept→decline→accept)
      // don't leave orphaned tick().then() callbacks that push after decline.
      let cancelled = false;
      tick().then(() => {
        if (!cancelled) {
          (window.adsbygoogle = window.adsbygoogle || []).push({});
        }
      });
      return () => { cancelled = true; };
    }
  });
</script>

{#if (adsEnabled && consentGranted) || testMode}
  <aside class="ad-shell" aria-label={label}>
    <span>{label}</span>
    {#if adsEnabled && !testMode}
      <ins
        class="adsbygoogle"
        style="display:block"
        data-ad-client={client}
        data-ad-slot={slot}
        data-ad-format={format}
        data-ad-layout={layout || undefined}
        data-full-width-responsive={responsive ? 'true' : 'false'}
      ></ins>
    {:else}
      <div class="ad-placeholder">AdSense preview</div>
    {/if}
  </aside>
{/if}

<style>
  .ad-shell {
    width: 100%;
    margin: 2rem auto;
    padding: 0.75rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg-surface);
    color: var(--text-secondary);
    text-align: center;
  }
  .ad-shell > span {
    display: block;
    margin-bottom: 0.5rem;
    font-size: 0.6875rem;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }
  .ad-placeholder {
    display: grid;
    min-height: 90px;
    place-items: center;
    border: 1px dashed var(--border);
    border-radius: calc(var(--radius) - 0.25rem);
    font-size: 0.875rem;
  }
  :global(.adsbygoogle) {
    min-height: 90px;
  }
  @media (max-width: 768px) {
    .ad-shell {
      margin: 1.5rem auto;
    }
  }
</style>
