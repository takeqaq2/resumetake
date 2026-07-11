<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';
  import { goto } from '$app/navigation';
  import { onMount, tick } from 'svelte';
  import { apiFetch } from '$lib/api.js';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  // R52b-F5: redirect back to where the user was after auth, instead of
  // always dumping them on /editor. Validates the redirect param is a
  // same-origin path (starts with /) to prevent open-redirect attacks.
  function redirectAfterAuth() {
    const redirect = $page.url.searchParams.get('redirect');
    if (redirect && redirect.startsWith('/') && !redirect.startsWith('//')) {
      goto(redirect);
    } else {
      goto(`/${lang}/editor`);
    }
  }

  let isLogin = $state(true);
  let email = $state('');
  let password = $state('');
  let name = $state('');
  let error = $state('');
  let loading = $state(false);

  let regStep = $state(0);
  let verifyCode = $state('');
  let sendingCode = $state(false);
  let countdown = $state(0);
  let verifying = $state(false);
  let countdownTimer = null;
  // R49-F2: track when the countdown started so we can compute remaining
  // time from real elapsed time. Without this, setInterval is throttled
  // when the tab is in the background, causing the countdown to drift —
  // the server may already allow a new request but the frontend still
  // shows countdown > 0, blocking the resend button.
  let countdownSentAt = 0;
  let checked = $state(false);
  // R55b-F1/F2: track mount state so async functions (sendCode, handleVerify,
  // handleRegister, handleSubmit) can skip writing to $state after unmount.
  // Without this, a countdownTimer set after navigation leaks for 60s,
  // and loading/error state writes to an unmounted component.
  let mounted = false;

  function startCountdown(seconds) {
    if (!mounted) return;
    countdown = seconds;
    countdownSentAt = Date.now();
    if (countdownTimer) clearInterval(countdownTimer);
    countdownTimer = setInterval(() => {
      // Compute from real elapsed time so background tab throttling
      // doesn't cause drift.
      const elapsed = Math.floor((Date.now() - countdownSentAt) / 1000);
      countdown = Math.max(0, seconds - elapsed);
      if (countdown <= 0) { clearInterval(countdownTimer); countdownTimer = null; }
    }, 1000);
  }

  onMount(() => {
    mounted = true;
    let token = null;
    try { token = localStorage.getItem('token'); } catch (e) {}
    if (token) {
      // R41b-M3: use redirectAfterAuth instead of hardcoded /editor —
      // already-logged-in users visiting /auth?redirect=/pricing should
      // go to /pricing, not /editor. Consistent with login/register paths.
      redirectAfterAuth();
      // R37b-F2: even on the early-return path, register the cleanup
      // function. goto() is async — the component stays mounted during
      // the navigation transition, and if sendCode runs in that window
      // (or if navigation is intercepted), a countdownTimer could be
      // set. Returning undefined here means Svelte never clears it.
      return () => { mounted = false; if (countdownTimer) clearInterval(countdownTimer); };
    }
    checked = true;
    return () => { mounted = false; if (countdownTimer) clearInterval(countdownTimer); };
  });

  async function sendCode() {
    if (!email || countdown > 0 || sendingCode) return false;
    sendingCode = true;
    error = '';
    try {
      const res = await apiFetch('/api/v1/auth/send-code', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
        skipAuth: true
      });
      let data;
      try { data = await res.json(); } catch (e) { error = t.auth.networkError || 'Network error'; return; }
      if (res.status === 429) {
        // Rate limited — start the countdown to prevent repeated requests
        // (the resend button stays disabled while countdown > 0).
        error = data.message || t.auth.rateLimited || t.auth.sendCodeFailed;
        startCountdown(60);
        return false;
      }
      if (data.success) {
        startCountdown(60);
        return true;
      } else {
        error = data.message || t.auth.sendCodeFailed;
        return false;
      }
    } catch {
      error = t.auth.networkError;
      return false;
    } finally {
      if (mounted) sendingCode = false;
    }
  }

  async function handleVerifyCode() {
    if (!verifyCode || verifyCode.length !== 6) {
      error = t.auth.verifyCodePrompt;
      return;
    }
    verifying = true;
    error = '';
    try {
      const res = await apiFetch('/api/v1/auth/verify-code', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, code: verifyCode }),
        skipAuth: true
      });
      let data;
      try { data = await res.json(); } catch (e) { error = t.auth.networkError || 'Network error'; return; }
      if (res.status === 429) {
        error = data.message || t.auth.rateLimited || t.auth.invalidCode;
      } else if (data.success) {
        // R55-F2: guard against user navigating away during async verify.
        // "换邮箱" and login/register toggle buttons are not disabled during
        // verifying, so the user may have returned to step 0 by now.
        if (regStep !== 1) return;
        regStep = 2;
        await tick();
        document.getElementById('name')?.focus();
      } else {
        error = data.message || t.auth.invalidCode;
      }
    } catch {
      error = t.auth.networkError;
    } finally {
      if (mounted) verifying = false;
    }
  }

  async function handleRegister() {
    loading = true;
    error = '';
    try {
      const res = await apiFetch('/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, name }),
        skipAuth: true
      });
      let data;
      try { data = await res.json(); } catch (e) { error = t.auth.networkError || 'Network error'; return; }
      if (res.status === 429) {
        error = data.message || t.auth.rateLimited || t.auth.registrationFailed;
      } else if (data.success && data.data && data.data.token) {
        try { localStorage.setItem('token', data.data.token); } catch (e) {}
        redirectAfterAuth();
      } else {
        error = data.message || t.auth.registrationFailed;
      }
    } catch {
      error = t.auth.networkError;
    } finally {
      if (mounted) loading = false;
    }
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if (isLogin) {
      error = '';
      if (!email) { error = t.auth.emailRequired; return; }
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email)) { error = t.auth.emailRequired; return; }
      if (!password) { error = t.auth.passwordRequired || 'Password is required'; return; }
      loading = true;
      try {
        const res = await apiFetch('/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password }),
          skipAuth: true
        });
        let data;
      try { data = await res.json(); } catch (e) { error = t.auth.networkError || 'Network error'; return; }
        if (res.status === 429) {
          error = data.message || t.auth.rateLimited || t.auth.loginFailed;
        } else if (data.success && data.data && data.data.token) {
          try { localStorage.setItem('token', data.data.token); } catch (e) {}
          redirectAfterAuth();
        } else {
          error = data.message || t.auth.loginFailed;
        }
      } catch {
        error = t.auth.networkErrorRetry;
      } finally {
        if (mounted) loading = false;
      }
    } else {
      if (regStep === 0) {
        if (!email) { error = t.auth.emailRequired; return; }
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) { error = t.auth.emailRequired; return; }
        loading = true;
        const sent = await sendCode();
        loading = false;
        if (sent) {
          regStep = 1;
          await tick();
          document.getElementById('verify-code')?.focus();
        }
      } else if (regStep === 1) {
        await handleVerifyCode();
      } else if (regStep === 2) {
        if (!name || !name.trim()) { error = t.auth.nameRequired; return; }
        if (!password || password.length < 6) { error = t.auth.passwordMinLength; return; }
        // R43-F3: Require at least one letter and one digit — "123456" or
        // "abcdef" alone are too weak for account security.
        if (!/[a-zA-Z]/.test(password) || !/\d/.test(password)) {
          error = t.auth.passwordWeak;
          return;
        }
        await handleRegister();
      }
    }
  }

  function resetReg() {
    regStep = 0;
    verifyCode = '';
    error = '';
    // Reset the resend countdown so a user who navigated away and came back
    // isn't silently blocked by a stale timer (sendCode returns false while
    // countdown > 0, leaving regStep at 0 with no visible feedback).
    countdown = 0;
    if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
  }
</script>

<svelte:head>
  <title>{isLogin ? t.auth.login : t.auth.register} - ResumeTake</title>
  <meta name="robots" content="noindex, nofollow">
  <link rel="canonical" href="https://resume.takee.top/{lang}/auth">
</svelte:head>

<div class="auth-page">
  <div class="auth-bg">
    <div class="orb orb-blue animate-float" aria-hidden="true" style="width:300px;height:300px;top:-10%;left:-5%"></div>
    <div class="orb orb-purple animate-float" aria-hidden="true" style="width:250px;height:250px;bottom:10%;right:-8%;animation-delay:3s"></div>
  </div>

  {#if checked}
  <div class="auth-container animate-fade-in-up">
    <div class="auth-header">
      <a href="/{lang}" class="auth-logo">
        <div class="logo-icon"><span aria-hidden="true" style="color:white;font-weight:700;position:relative;z-index:1">R</span></div>
        <span style="font-weight:700;font-size:1.25rem">ResumeTake</span>
      </a>
      <h1>{isLogin ? t.auth.login : t.auth.register}</h1>
      <p>{isLogin ? (t.auth.welcomeBack) : (t.auth.createAccountPrompt)}</p>
    </div>

    {#if error}
      <div class="auth-error" role="alert"><span aria-hidden="true">⚠️</span> {error}</div>
    {/if}

    <form onsubmit={handleSubmit} novalidate>
      {#if isLogin}
        <div class="form-group">
          <label for="email" class="label">{t.auth.email}</label>
          <input id="email" class="input" type="email" bind:value={email} required autocomplete="email" placeholder={t.auth.emailPlaceholder}>
        </div>
        <div class="form-group">
          <label for="password" class="label">{t.auth.password}</label>
          <input id="password" class="input" type="password" bind:value={password} required minlength="6" autocomplete="current-password" placeholder={t.auth.passwordPlaceholder}>
        </div>
      {:else}
        {#if regStep === 0}
          <div class="form-group">
            <label for="email" class="label">{t.auth.email}</label>
            <input id="email" class="input" type="email" bind:value={email} required autocomplete="email" placeholder={t.auth.emailPlaceholder}>
          </div>
        {:else if regStep === 1}
          <div class="form-group">
            <label class="label" for="verify-code">{t.auth.codeSentTo} {email}</label>
            <div style="display:flex;gap:0.5rem">
              <input id="verify-code" class="input" type="text" inputmode="numeric" autocomplete="one-time-code" value={verifyCode} maxlength="6" oninput={(e) => { verifyCode = e.currentTarget.value.replace(/\D/g, '').slice(0, 6); }} placeholder="000000" style="flex:1;letter-spacing:4px;text-align:center;font-size:1.25rem">
              <button type="button" class="btn btn-secondary" onclick={sendCode} disabled={countdown > 0 || sendingCode} style="white-space:nowrap;font-size:0.8125rem">
                {countdown > 0 ? `${countdown}s` : (t.auth.resend)}
              </button>
            </div>
          </div>
          <button type="button" class="btn-link" onclick={resetReg} style="font-size:0.8125rem;color:var(--text-secondary);margin-bottom:0.5rem">
            <span class="auth-back-arrow" aria-hidden="true">←</span> {t.auth.changeEmail}
          </button>
        {:else if regStep === 2}
          <div class="form-group">
            <label for="name" class="label">{t.auth.name}</label>
            <input id="name" class="input" type="text" bind:value={name} required autocomplete="name" placeholder={t.auth.namePlaceholder}>
          </div>
          <div class="form-group">
            <label for="password" class="label">{t.auth.password}</label>
            <input id="password" class="input" type="password" bind:value={password} required minlength="6" autocomplete="new-password" placeholder={t.auth.passwordPlaceholder}>
          </div>
        {/if}
      {/if}

      <button class="btn btn-primary auth-submit" type="submit" disabled={loading || verifying} aria-label={loading || verifying ? t.auth.pleaseWait : undefined}>
        {#if loading || verifying}
          <span class="auth-spinner" aria-hidden="true"></span>
          <span class="sr-only">{t.auth.pleaseWait}</span>
        {:else}
          {#if isLogin}
            {t.auth.loginBtn}
          {:else if regStep === 0}
            {t.auth.sendCode}
          {:else if regStep === 1}
            {t.auth.verify}
          {:else}
            {t.auth.registerBtn}
          {/if}
        {/if}
      </button>
    </form>

    <div class="auth-toggle">
      {#if isLogin}
        <span>{t.auth.noAccount}</span>
      {:else}
        <span>{t.auth.hasAccount}</span>
      {/if}
      <button onclick={() => { isLogin = !isLogin; resetReg(); }}>
        {isLogin ? t.auth.registerBtn : t.auth.loginBtn}
      </button>
    </div>
  </div>
  {:else}
    <div class="loading-screen" role="status" aria-live="polite" style="display:flex;justify-content:center;align-items:center;min-height:50vh">
      <div class="spinner" style="width:2rem;height:2rem;border:2px solid var(--border);border-top-color:var(--primary);border-radius:50%;animation:spin 0.8s linear infinite" aria-hidden="true"></div>
      <span class="sr-only">{t.auth.pleaseWait || 'Loading...'}</span>
    </div>
  {/if}
</div>

<style>
  /* Mirror the back arrow in RTL so it points "forward" visually. */
  :global([dir="rtl"]) .auth-back-arrow { display: inline-block; transform: scaleX(-1); }
  .auth-page {
    min-height: 100vh; display: flex; align-items: center; justify-content: center;
    padding: 2rem 1.5rem; position: relative; overflow: hidden;
    background: var(--gradient-hero); background-size: 200% 200%;
    animation: gradientShift 10s ease-in-out infinite;
  }
  .auth-bg {
    position: absolute; inset: 0; pointer-events: none;
  }
  .auth-container {
    width: 100%; max-width: 26rem; background: var(--bg-glass);
    border: 1px solid var(--border); border-radius: var(--radius-lg);
    padding: 2.5rem; backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px); box-shadow: var(--shadow-lg);
    position: relative; z-index: 1;
  }
  .auth-header { text-align: center; margin-bottom: 2rem; }
  .auth-logo {
    display: inline-flex; align-items: center; gap: 0.5rem;
    text-decoration: none; margin-bottom: 1.5rem;
  }
  .auth-header h1 {
    font-size: 1.5rem; font-weight: 700; color: var(--text);
    margin-bottom: 0.5rem;
  }
  .auth-header p {
    font-size: 0.875rem; color: var(--text-secondary);
  }
  .form-group { margin-bottom: 1.25rem; }
  .auth-submit {
    width: 100%; padding: 0.875rem; font-size: 1rem; font-weight: 600;
    margin-top: 0.5rem;
  }
  .auth-spinner {
    width: 18px; height: 18px; border: 2px solid rgba(255,255,255,0.3);
    border-top-color: white; border-radius: 50%;
    animation: spin 0.6s linear infinite; display: inline-block;
  }
  .auth-error {
    padding: 0.75rem 1rem; border-radius: var(--radius);
    background: rgba(239,68,68,0.08); border: 1px solid rgba(239,68,68,0.2);
    color: var(--error-text); font-size: 0.875rem; margin-bottom: 1.25rem;
  }
  .auth-toggle {
    text-align: center; margin-top: 1.5rem;
    font-size: 0.875rem; color: var(--text-secondary);
  }
  .auth-toggle button {
    background: none; border: none; color: var(--primary);
    font-weight: 600; cursor: pointer; font-size: 0.875rem;
    padding: 0; margin-inline-start: 0.375rem;
  }
  .auth-toggle button:hover { text-decoration: underline; }
  @media (prefers-reduced-motion: reduce) {
    .auth-page { animation: none !important; }
  }
</style>
