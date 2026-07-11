<script>
  import { browser } from '$app/environment';
  import { env } from '$env/dynamic/public';
  import { onMount } from 'svelte';

  const client = env.PUBLIC_GOOGLE_ADSENSE_CLIENT || '';
  const adsConfigured = env.PUBLIC_ADS_ENABLED === 'true' && /^ca-pub-\d+$/.test(client);

  // GDPR/ePrivacy: only load AdSense script after user has explicitly
  // accepted advertising cookies. Without this gate, tracking cookies are
  // set before consent, violating EU regulations.
  let consentGranted = $state(false);

  // R51-F5: consent must be checked for expiry — CookieNotice uses a 6-month
  // expiry, but checkConsent previously accepted stale consent (or the old
  // 'true' format with no timestamp). Under GDPR/ePrivacy, consent must be
  // revocable and time-bound; loading ad scripts on expired consent violates
  // the regulation.
  const CONSENT_EXPIRY_MS = 180 * 24 * 60 * 60 * 1000; // 6 months

  function checkConsent() {
    try {
      const raw = localStorage.getItem('cookie_notice_accepted');
      // R49-F1: legacy 'true' format has no timestamp — treat as invalid
      // (same as CookieNotice does) to prevent loading AdSense before the
      // user re-confirms consent. GDPR/ePrivacy requires consent before
      // loading tracking scripts.
      if (raw === 'true') return false;
      if (raw === 'declined') return false;
      const consent = JSON.parse(raw);
      if (consent?.status !== 'true') return false;
      // Check expiry — if consent is older than 6 months, treat as not given.
      if (consent.ts && Date.now() - consent.ts > CONSENT_EXPIRY_MS) {
        return false;
      }
      return true;
    } catch {
      return false;
    }
  }

  onMount(() => {
    if (!browser || !adsConfigured) return;
    consentGranted = checkConsent();
    // R52b-F4: always register the consent listener regardless of initial
    // state. Previously only registered when !consentGranted — if consent
    // was already granted on mount, a later "decline" event was never
    // received and the AdSense script stayed in <head> tracking after
    // revocation (GDPR gap).
    const onConsent = (e) => {
      if (e.detail === 'accepted') consentGranted = true;
      else if (e.detail === 'declined') consentGranted = false;
    };
    window.addEventListener('cookie:consent', onConsent);
    return () => window.removeEventListener('cookie:consent', onConsent);
  });

  // R56b-F5: when consent state CHANGES after initial load, reload the page.
  // Removing a <script> tag from <head> (via the {#if} below) does NOT stop
  // the already-evaluated AdSense library — it keeps making ad requests and
  // setting tracking cookies (AD1). Conversely, re-adding a previously-
  // removed script tag does NOT re-execute it (browsers cache the src), so
  // re-accepting consent after declining leaves ads broken (AD2). A full
  // reload is the only reliable client-side way to sync the AdSense script
  // with the new consent state. consentInitialized skips the first effect
  // run (initial hydration) so we don't reload on page load.
  let consentInitialized = false;
  $effect(() => {
    if (!browser || !adsConfigured) return;
    // R57-F1: 必须在 effect 体内读取 consentGranted ($state) 才能建立
    // 响应式追踪。此前 effect 体只读写非响应变量 (consentInitialized
    // 是普通 let，browser/adsConfigured 是 const)，Svelte 5 不会把
    // consentGranted 视为依赖，effect 只在挂载时运行一次，导致
    // consentGranted 变化后 reload 永不触发 — 同意/拒绝切换后广告
    // 脚本与实际授权状态长期不一致 (GDPR 合规漏洞)。
    const current = consentGranted;
    if (!consentInitialized) {
      consentInitialized = true;
      return;
    }
    // consentGranted changed after init — reload to sync the script tag.
    // current 只用于建立追踪，实际 reload 由 consentGranted 变化触发。
    void current;
    window.location.reload();
  });
</script>

<svelte:head>
  {#if adsConfigured && consentGranted}
    <link rel="preconnect" href="https://pagead2.googlesyndication.com">
    <script async src="https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client={client}" crossorigin="anonymous"></script>
  {/if}
</svelte:head>
