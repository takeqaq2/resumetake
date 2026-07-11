<script>
  import { browser } from '$app/environment';
  import { onMount, tick } from 'svelte';
  import { getTranslation } from './i18n/index.js';

  let { lang = 'en' } = $props();
  let visible = $state(false);
  let t = $derived(getTranslation(lang));
  let noticeEl;
  // R57-F9: track whether the user has already interacted with the notice.
  // When true and the notice is hidden, a small "cookie settings" button
  // is rendered so declined users can re-grant consent at any time —
  // GDPR 7(3) requires withdrawal of consent to be as easy as giving it.
  let hasInteracted = $state(false);

  // Consent expiry: 6 months. Some regulations (ePrivacy) recommend
  // re-soliciting consent periodically. Stored as JSON with a timestamp.
  const CONSENT_EXPIRY_MS = 180 * 24 * 60 * 60 * 1000;

  function readConsent() {
    try {
      const raw = localStorage.getItem('cookie_notice_accepted');
      if (!raw) return null;
      // Backward compat: old format was plain 'true' or 'declined' (no ts).
      // R37-F6: treat old 'true' as expired — re-solicit consent per GDPR
      // periodic re-consent recommendation.
      if (raw === 'true') return null;
      if (raw === 'declined') return { status: 'declined', ts: 0 };
      return JSON.parse(raw);
    } catch (e) {
      // R57-F10: log instead of silently swallowing — helps debugging
      // localStorage corruption (e.g. quota exceeded, disabled storage).
      console.warn('[CookieNotice] readConsent failed:', e);
      return null;
    }
  }

  onMount(() => {
    try {
      const consent = readConsent();
      if (!consent) {
        visible = true;
      } else if (consent.ts && Date.now() - consent.ts > CONSENT_EXPIRY_MS) {
        // Consent expired — re-solicit
        visible = true;
      } else {
        visible = false;
        hasInteracted = true;
      }
    } catch (e) {
      console.warn('[CookieNotice] onMount failed:', e);
      visible = true;
    }
    // R53b-F5: move focus to the notice so keyboard/screen-reader users can
    // interact with it without tabbing through the entire page first. We wait
    // for tick() so the DOM node exists before focusing.
    if (visible) {
      tick().then(() => { noticeEl?.focus(); });
    }
  });

  function accept() {
    try {
      localStorage.setItem('cookie_notice_accepted', JSON.stringify({ status: 'true', ts: Date.now() }));
    } catch (e) {
      // R57-F10: log instead of silently swallowing.
      console.warn('[CookieNotice] accept localStorage.setItem failed:', e);
    }
    visible = false;
    hasInteracted = true;
    window.dispatchEvent(new CustomEvent('cookie:consent', { detail: 'accepted' }));
  }

  function decline() {
    try {
      localStorage.setItem('cookie_notice_accepted', JSON.stringify({ status: 'declined', ts: Date.now() }));
    } catch (e) {
      // R57-F10: log instead of silently swallowing.
      console.warn('[CookieNotice] decline localStorage.setItem failed:', e);
    }
    visible = false;
    hasInteracted = true;
    window.dispatchEvent(new CustomEvent('cookie:consent', { detail: 'declined' }));
  }

  function reopen() {
    visible = true;
    tick().then(() => { noticeEl?.focus(); });
  }
</script>

{#if browser && visible}
  <div class="cookie-notice" role="region" aria-label={t.cookie.noticeAria} aria-live="polite" tabindex="-1" bind:this={noticeEl}>
    <p>{t.cookie.message}</p>
    <div class="cookie-actions">
      <a href="/{lang}/privacy">{t.cookie.privacyPolicy}</a>
      <button class="btn btn-secondary" onclick={decline} style="font-size:0.8125rem">{t.cookie.decline}</button>
      <button class="btn btn-primary" onclick={accept}>{t.cookie.ok}</button>
    </div>
  </div>
{:else if browser && hasInteracted}
  <!-- R57-F9: re-open entry point for declined/accepted users. GDPR 7(3)
       requires consent withdrawal to be as easy as giving it. -->
  <button class="cookie-reopen" onclick={reopen} aria-label={t.cookie.noticeAria} title={t.cookie.noticeAria}>
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M12 2a10 10 0 100 20 10 10 0 000-20zm0 14a1.5 1.5 0 110-3 1.5 1.5 0 010 3zm1-5a1 1 0 01-2 0V7a1 1 0 012 0v4z" fill="currentColor"/></svg>
  </button>
{/if}

<style>
  .cookie-notice {
    position: fixed;
    right: 1rem;
    bottom: 1rem;
    z-index: 120;
    width: min(calc(100vw - 2rem), 26rem);
    padding: 1rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg-surface);
    box-shadow: var(--shadow-lg);
  }
  .cookie-reopen {
    position: fixed;
    right: 1rem;
    bottom: 1rem;
    z-index: 119;
    width: 2.25rem;
    height: 2.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--border);
    border-radius: 50%;
    background: var(--bg-surface);
    box-shadow: var(--shadow-md);
    color: var(--text-secondary);
    cursor: pointer;
    opacity: 0.6;
    transition: opacity 0.2s;
  }
  .cookie-reopen:hover {
    opacity: 1;
  }
  p {
    margin: 0 0 0.75rem;
    color: var(--text-secondary);
    font-size: 0.875rem;
    line-height: 1.5;
  }
  .cookie-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    flex-wrap: wrap;
    gap: 0.75rem;
  }
  a {
    color: var(--primary);
    font-size: 0.875rem;
  }
</style>
