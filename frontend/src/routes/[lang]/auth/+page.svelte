<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

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

  onMount(() => {
    const token = localStorage.getItem('token');
    if (token) goto(`/${lang}/editor`);
  });

  async function sendCode() {
    if (!email || countdown > 0) return;
    sendingCode = true;
    error = '';
    try {
      const res = await fetch('/api/v1/auth/send-code', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email })
      });
      const data = await res.json();
      if (data.success) {
        countdown = 60;
        const timer = setInterval(() => {
          countdown--;
          if (countdown <= 0) clearInterval(timer);
        }, 1000);
      } else {
        error = data.message || 'Failed to send code';
      }
    } catch {
      error = lang === 'zh' ? '网络错误' : 'Network error';
    } finally {
      sendingCode = false;
    }
  }

  async function handleVerifyCode() {
    if (!verifyCode || verifyCode.length !== 6) return;
    verifying = true;
    error = '';
    try {
      const res = await fetch('/api/v1/auth/verify-code', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, code: verifyCode })
      });
      const data = await res.json();
      if (data.success) {
        regStep = 2;
      } else {
        error = data.message || 'Invalid code';
      }
    } catch {
      error = lang === 'zh' ? '网络错误' : 'Network error';
    } finally {
      verifying = false;
    }
  }

  async function handleRegister() {
    loading = true;
    error = '';
    try {
      const res = await fetch('/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, name })
      });
      const data = await res.json();
      if (data.success && data.data && data.data.token) {
        localStorage.setItem('token', data.data.token);
        goto(`/${lang}/editor`);
      } else {
        error = data.message || 'Registration failed';
      }
    } catch {
      error = lang === 'zh' ? '网络错误' : 'Network error';
    } finally {
      loading = false;
    }
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if (isLogin) {
      error = '';
      loading = true;
      try {
        const res = await fetch('/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
        });
        const data = await res.json();
        if (data.success && data.token) {
          localStorage.setItem('token', data.token);
          goto(`/${lang}/editor`);
        } else {
          error = data.message || 'Login failed';
        }
      } catch {
        error = lang === 'zh' ? '网络错误，请重试' : 'Network error, please try again';
      } finally {
        loading = false;
      }
    } else {
      if (regStep === 0) {
        if (!email) { error = lang === 'zh' ? '请输入邮箱' : 'Please enter email'; return; }
        await sendCode();
        regStep = 1;
      } else if (regStep === 1) {
        await handleVerifyCode();
      } else if (regStep === 2) {
        if (!password || password.length < 6) { error = lang === 'zh' ? '密码至少6位' : 'Password must be at least 6 chars'; return; }
        await handleRegister();
      }
    }
  }

  function resetReg() {
    regStep = 0;
    verifyCode = '';
    error = '';
  }
</script>

<svelte:head>
  <title>{isLogin ? t.auth.login : t.auth.register} - ResumeTake</title>
</svelte:head>

<div class="auth-page">
  <div class="auth-bg">
    <div class="orb orb-blue animate-float" style="width:300px;height:300px;top:-10%;left:-5%"></div>
    <div class="orb orb-purple animate-float" style="width:250px;height:250px;bottom:10%;right:-8%;animation-delay:3s"></div>
  </div>

  <div class="auth-container animate-fade-in-up">
    <div class="auth-header">
      <a href="/{lang}" class="auth-logo">
        <div class="logo-icon"><span style="color:white;font-weight:700;position:relative;z-index:1">R</span></div>
        <span style="font-weight:700;font-size:1.25rem">ResumeTake</span>
      </a>
      <h1>{isLogin ? t.auth.login : t.auth.register}</h1>
      <p>{isLogin ? (lang === 'zh' ? '欢迎回来，请登录您的账户' : 'Welcome back, please sign in') : (lang === 'zh' ? '创建账户，开始优化简历' : 'Create an account to start optimizing')}</p>
    </div>

    {#if error}
      <div class="auth-error">⚠️ {error}</div>
    {/if}

    <form onsubmit={handleSubmit}>
      {#if isLogin}
        <div class="form-group">
          <label for="email" class="label">{t.auth.email}</label>
          <input id="email" class="input" type="email" bind:value={email} required placeholder={lang === 'zh' ? '请输入邮箱' : 'Enter your email'}>
        </div>
        <div class="form-group">
          <label for="password" class="label">{t.auth.password}</label>
          <input id="password" class="input" type="password" bind:value={password} required minlength="6" placeholder={lang === 'zh' ? '请输入密码（至少6位）' : 'Enter password (min 6 chars)'}>
        </div>
      {:else}
        {#if regStep === 0}
          <div class="form-group">
            <label for="email" class="label">{t.auth.email}</label>
            <input id="email" class="input" type="email" bind:value={email} required placeholder={lang === 'zh' ? '请输入邮箱' : 'Enter your email'}>
          </div>
        {:else if regStep === 1}
          <div class="form-group">
            <label class="label">{lang === 'zh' ? '验证码已发送至' : 'Code sent to'} {email}</label>
            <div style="display:flex;gap:0.5rem">
              <input class="input" type="text" bind:value={verifyCode} maxlength="6" placeholder="000000" style="flex:1;letter-spacing:4px;text-align:center;font-size:1.25rem">
              <button type="button" class="btn btn-secondary" onclick={sendCode} disabled={countdown > 0 || sendingCode} style="white-space:nowrap;font-size:0.8125rem">
                {countdown > 0 ? `${countdown}s` : (lang === 'zh' ? '重新发送' : 'Resend')}
              </button>
            </div>
          </div>
          <button type="button" class="btn-link" onclick={resetReg} style="font-size:0.8125rem;color:var(--text-secondary);margin-bottom:0.5rem">
            ← {lang === 'zh' ? '更换邮箱' : 'Change email'}
          </button>
        {:else if regStep === 2}
          <div class="form-group">
            <label for="name" class="label">{t.auth.name}</label>
            <input id="name" class="input" type="text" bind:value={name} required placeholder={lang === 'zh' ? '请输入姓名' : 'Enter your name'}>
          </div>
          <div class="form-group">
            <label for="password" class="label">{t.auth.password}</label>
            <input id="password" class="input" type="password" bind:value={password} required minlength="6" placeholder={lang === 'zh' ? '请输入密码（至少6位）' : 'Enter password (min 6 chars)'}>
          </div>
        {/if}
      {/if}

      <button class="btn btn-primary auth-submit" type="submit" disabled={loading || verifying}>
        {#if loading || verifying}
          <span class="auth-spinner"></span>
        {:else}
          {#if isLogin}
            {t.auth.loginBtn}
          {:else if regStep === 0}
            {lang === 'zh' ? '发送验证码' : 'Send Code'}
          {:else if regStep === 1}
            {lang === 'zh' ? '验证' : 'Verify'}
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
      <button onclick={() => { isLogin = !isLogin; error = ''; regStep = 0; verifyCode = ''; }}>
        {isLogin ? t.auth.registerBtn : t.auth.loginBtn}
      </button>
    </div>
  </div>
</div>

<style>
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
  @keyframes spin { to { transform: rotate(360deg); } }
  .auth-error {
    padding: 0.75rem 1rem; border-radius: var(--radius);
    background: rgba(239,68,68,0.08); border: 1px solid rgba(239,68,68,0.2);
    color: #ef4444; font-size: 0.875rem; margin-bottom: 1.25rem;
  }
  .auth-toggle {
    text-align: center; margin-top: 1.5rem;
    font-size: 0.875rem; color: var(--text-secondary);
  }
  .auth-toggle button {
    background: none; border: none; color: var(--primary);
    font-weight: 600; cursor: pointer; font-size: 0.875rem;
    padding: 0; margin-left: 0.375rem;
  }
  .auth-toggle button:hover { text-decoration: underline; }
</style>
