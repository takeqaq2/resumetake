<script>
  import { page } from '$app/stores';
  import { goto, replaceState } from '$app/navigation';
  import { getTranslation } from '$lib/i18n/index.js';
  import { apiFetch, getToken } from '$lib/api.js';
  import { onMount } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));
  let user = $state(null);
  let userLoading = $state(true);
  let purchasing = $state('');
  let error = $state('');
  let success = $state('');
  let redirectTimer;
  let cardRefs = $state([]);
  // R55b-F3: track mount state so async callbacks don't fire goto()/setState
  // after the user has navigated away.
  let mounted = false;

  // R49-F1: display price for paid templates. Source of truth is
  // backend/handlers/product.go templatePrices map (all $2.99 USD).
  // If backend prices change, update this constant to match.
  // Payment is always charged in USD via PayPal regardless of locale.
  const templatePrice = '$2.99';

  // --- Per-template PayPal purchase state (mirrors pricing page) ---
  let paypalClientId = $state('');
  let paypalRendered = $state({}); // tplId -> bool, so each button renders once
  let paypalContainerRefs = $state([]); // index-aligned with t.templates.items
  let sdkLoadFailed = $state(false);
  let sdkLoading = $state(false);
  let isLoggedIn = $state(false);

  // Svelte 5 declarative pattern: $effect runs client-side only, replaces
  // the old onMount querySelectorAll('.reveal') approach. bind:this collects
  // element refs without manual DOM lookup.
  $effect(() => {
    const cards = cardRefs.filter(Boolean);
    if (cards.length === 0 || cards.length < (t.templates?.items?.length || 0)) return;
    cards.forEach(el => el.classList.add('js-ready'));
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(e => { if (e.isIntersecting) { e.target.classList.add('visible'); observer.unobserve(e.target); } });
    }, { threshold: 0.1 });
    cards.forEach(el => observer.observe(el));
    return () => observer.disconnect();
  });

  onMount(() => {
    mounted = true;
    let cancelled = false;
    const token = getToken();
    isLoggedIn = !!token;
    if (token) {
      apiFetch("/api/v1/auth/me", { skipAuth: true }).then(res => res.ok ? res.json() : null).then(data => {
        if (cancelled) return;
        if (data?.data) user = data.data;
        userLoading = false;
        // User data changed the free/owned state of cards → ensure PayPal
        // buttons are rendered for the now-paid templates.
        setTimeout(() => maybeLoadPaypal(), 0);
      }).catch(() => { if (!cancelled) userLoading = false; setTimeout(() => maybeLoadPaypal(), 0); });
    } else {
      userLoading = false;
    }

    // Fetch the PayPal client id so we can render the buy buttons.
    apiFetch('/api/config', { skipAuth: true }).then(r => r.json()).then(data => {
      if (cancelled) return;
      if (data?.paypal_client_id) {
        paypalClientId = data.paypal_client_id;
        setTimeout(() => maybeLoadPaypal(), 0);
      }
    }).catch(() => {});

    // Handle the browser-redirect return (?payment=success / ?payment=cancelled).
    // The primary capture path is the PayPal SDK onApprove (client-side); this
    // is a safety net for users who close the popup and are redirected back.
    const params = new URLSearchParams(window.location.search);
    if (params.get('payment') === 'success') {
      success = t.templates.purchaseSuccess;
      replaceState(window.location.pathname, {});
    } else if (params.get('payment') === 'cancelled') {
      error = t.templates.paymentRequired;
      replaceState(window.location.pathname, {});
    }

    return () => { cancelled = true; mounted = false; clearTimeout(redirectTimer); };
  });

  function isTemplateFree(tplId) {
    if (user?.plan === "pro" || user?.plan === "enterprise") return true;
    // "professional" is the free template; everything else requires purchase
    // or a Pro/Enterprise plan. Owned (purchased) templates are handled
    // separately so they show an "Owned" badge instead of "Free".
    return tplId === 'professional';
  }

  // Load the PayPal JS SDK once. Mirrors the pricing page's loader, including
  // the 15s timeout so a silently-dropped script (ad blocker) surfaces a retry.
  function loadPaypalScript() {
    return new Promise((resolve, reject) => {
      const existing = document.getElementById('paypal-sdk');
      if (existing) {
        if (typeof window.paypal !== 'undefined') { resolve(); return; }
        existing.remove();
      }
      const script = document.createElement('script');
      script.id = 'paypal-sdk';
      script.src = `https://www.paypal.com/sdk/js?client-id=${encodeURIComponent(paypalClientId)}&currency=USD&intent=capture`;
      script.async = true;
      const loadTimeout = setTimeout(() => {
        script.onload = script.onerror = null;
        script.remove();
        reject(new Error('paypal-sdk load timeout'));
      }, 15000);
      script.onload = () => { clearTimeout(loadTimeout); resolve(); };
      script.onerror = () => { clearTimeout(loadTimeout); reject(new Error('paypal-sdk load failed')); };
      document.head.appendChild(script);
    });
  }

  function maybeLoadPaypal() {
    if (sdkLoading || sdkLoadFailed) return;
    if (!paypalClientId || !isLoggedIn) return;
    if (typeof window.paypal !== 'undefined') { renderTemplateButtons(); return; }
    sdkLoading = true;
    loadPaypalScript()
      .then(() => { sdkLoading = false; if (mounted) renderTemplateButtons(); })
      .catch(() => { sdkLoading = false; sdkLoadFailed = true; });
  }

  function renderTemplateButtons() {
    if (typeof window.paypal === 'undefined' || !mounted) return;
    (t.templates?.items || []).forEach((tpl, i) => {
      if (isTemplateFree(tpl.id, i)) return;
      if (paypalRendered[tpl.id]) return;
      const container = paypalContainerRefs[i];
      if (!container) return;
      container.replaceChildren();
      window.paypal.Buttons({
        style: { layout: 'vertical', color: 'blue', shape: 'rect', label: 'pay', height: 40 },
        createOrder: (data, actions) => createTemplateOrder(tpl.id, data, actions),
        onApprove: (data, actions) => onTemplateApprove(tpl.id, data, actions),
        onError: () => { if (mounted) error = t.templates.purchaseFailed; },
        onCancel: () => { if (mounted && !error) error = t.templates.paymentRequired; }
      }).render(container)
        .then(() => { paypalRendered = { ...paypalRendered, [tpl.id]: true }; })
        .catch(() => { /* render failed — fall back to the plain Buy button */ });
    });
  }

  async function createTemplateOrder(tplId, data, actions) {
    error = '';
    const token = getToken();
    if (!token) {
      goto(`/${lang}/auth?redirect=${encodeURIComponent(`/${lang}/templates`)}`);
      throw new Error('NOT_AUTHENTICATED');
    }
    purchasing = tplId;
    try {
      const res = await apiFetch('/api/v1/purchase-template', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ template_id: tplId, lang })
      });
      let result;
      try { result = await res.json(); } catch { error = t.templates.networkError; purchasing = ''; throw new Error('ORDER_CREATE_FAILED'); }
      if (!res.ok || !result.success || !result.data?.order_id) {
        error = result.message || t.templates.purchaseFailed;
        purchasing = '';
        throw new Error(result.message || 'ORDER_CREATE_FAILED');
      }
      return result.data.order_id;
    } catch (e) {
      if (e.message === 'NOT_AUTHENTICATED' || e.message === 'ORDER_CREATE_FAILED') throw e;
      error = t.templates.networkError;
      purchasing = '';
      throw new Error('NETWORK_ERROR');
    }
  }

  async function onTemplateApprove(tplId, data, actions) {
    const token = getToken();
    if (!token) {
      error = t.templates.purchaseFailed;
      purchasing = '';
      return;
    }
    error = '';
    try {
      const res = await apiFetch('/api/v1/capture-template-order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ order_id: data.orderID, template_id: tplId })
      });
      if (!mounted) return;
      let result;
      try { result = await res.json(); } catch { error = t.templates.purchaseFailed; purchasing = ''; return; }
      if (res.ok && result.success) {
        // Update local user so the "purchased" badge shows immediately. The
        // next /auth/me will also return it (persisted server-side).
        if (user) {
          const purchased = user.purchased_templates ? [...user.purchased_templates] : [];
          if (!purchased.includes(tplId)) purchased.push(tplId);
          user = { ...user, purchased_templates: purchased };
        }
        success = t.templates.purchaseSuccess;
        purchasing = '';
      } else {
        error = result.message || t.templates.purchaseFailed;
        purchasing = '';
      }
    } catch (e) {
      if (mounted) { error = t.templates.networkError; purchasing = ''; }
    }
  }

  // Fallback for not-logged-in users or when the SDK failed to load: redirect
  // to login, or surface an error. The primary path is the PayPal button.
  async function purchaseTemplateFallback(tplId) {
    const token = getToken();
    if (!token) {
      goto(`/${lang}/auth?redirect=${encodeURIComponent(`/${lang}/templates`)}`);
      return;
    }
    if (sdkLoadFailed) {
      error = t.templates.purchaseFailed;
      return;
    }
    maybeLoadPaypal();
  }
</script>

<svelte:head>
  <title>{t.meta.templatesTitle}</title>
  <meta name="description" content={t.meta.templatesDesc}>
  <meta name="keywords" content={t.meta.templatesKeywords}>
  <link rel="canonical" href="https://resume.takee.top/{lang}/templates">
  <meta property="og:title" content={t.meta.templatesTitle}>
  <meta property="og:description" content={t.meta.templatesDesc}>
  <meta property="og:url" content="https://resume.takee.top/{lang}/templates">
  <meta property="og:type" content="website">
</svelte:head>

<div class="editor-header">
  <div class="orb orb-blue animate-float" aria-hidden="true" style="width:200px;height:200px;top:-20%;left:10%"></div>
  <div class="orb orb-pink animate-float" aria-hidden="true" style="width:160px;height:160px;bottom:-10%;right:15%;animation-delay:2s"></div>
  <div class="container" style="position:relative;text-align:center">
    <h1 class="anim-hero anim-hero-1" style="font-size:clamp(1.75rem,3.5vw,2.25rem);font-weight:700;margin-bottom:0.75rem;color:var(--text)">{t.templates.title}</h1>
    <p class="anim-hero anim-hero-2" style="color:var(--text-secondary);max-width:32rem;margin:0 auto">{t.templates.subtitle}</p>
  </div>
</div>

<div class="container" style="padding:2.5rem 1.5rem;margin-top:-1rem">
  {#if error}
    <div class="error-banner" role="alert">{error}</div>
  {/if}
  {#if success}
    <div class="success-banner" role="status">{success}</div>
  {/if}
  {#if user?.plan === "pro" || user?.plan === "enterprise"}
    <div class="pro-notice">{t.templates.proNotice}</div>
  {/if}
  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:1.5rem">
    {#each t.templates.items as tpl, i (tpl.id)}
      {@const gradients = {
        professional: 'linear-gradient(135deg,#3b82f6,#2563eb)',
        modern: 'linear-gradient(135deg,#8b5cf6,#7c3aed)',
        creative: 'linear-gradient(135deg,#ec4899,#db2777)',
        academic: 'linear-gradient(135deg,#047857,#047857)',
        executive: 'linear-gradient(135deg,#374151,#111827)',
        minimal: 'linear-gradient(135deg,#f59e0b,#d97706)'
      }}
      {@const free = isTemplateFree(tpl.id)}
      {@const owned = user?.purchased_templates?.includes(tpl.id)}
      <div class="feature-card reveal template-card" bind:this={cardRefs[i]} style="padding:0;overflow:hidden;transition-delay:{i * 0.08}s">
        <a href="/{lang}/editor" style="text-decoration:none;color:inherit;display:block">
          <div style="height:10rem;background:{gradients[tpl.id]||gradients.modern};display:flex;align-items:center;justify-content:center;position:relative;overflow:hidden">
            <div style="position:absolute;inset:0;background:radial-gradient(circle at 30% 40%, rgba(255,255,255,0.15) 0%, transparent 60%)"></div>
            <span aria-hidden="true" style="color:rgba(255,255,255,0.25);font-size:3.5rem;font-weight:700;position:relative;z-index:1">Aa</span>
          </div>
          <div style="padding:1.25rem;text-align:left">
            <h3 style="font-weight:600;margin-bottom:0.375rem;color:var(--text)">{tpl.name}</h3>
            <p style="font-size:0.875rem;color:var(--text-secondary);line-height:1.5">{tpl.desc}</p>
          </div>
        </a>
        <div style="padding:0 1.25rem 1.25rem;display:flex;align-items:center;justify-content:space-between">
          {#if userLoading}
            <span class="template-badge" style="background:var(--bg-glass);border:1px solid var(--border);color:var(--text-secondary)">…</span>
            <button class="template-use-btn" disabled style="opacity:0.5">…</button>
          {:else if free}
            <span class="template-badge free">{t.templates.freeBadge}</span>
            <a href="/{lang}/editor" class="template-use-btn">{t.templates.use}</a>
          {:else if owned}
            <span class="template-badge owned">{t.templates.ownedBadge}</span>
            <a href="/{lang}/editor" class="template-use-btn">{t.templates.use}</a>
          {:else if isLoggedIn && paypalClientId && !sdkLoadFailed}
            <span class="template-badge price">{templatePrice}</span>
            <div bind:this={paypalContainerRefs[i]} class="paypal-template-btn" aria-label={t.templates.buy}></div>
          {:else}
            <span class="template-badge price">{templatePrice}</span>
            <button class="template-buy-btn" onclick={() => purchaseTemplateFallback(tpl.id)} disabled={!!purchasing || sdkLoading}>
              {purchasing === tpl.id ? t.templates.processing : t.templates.buy}
            </button>
          {/if}
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
  .anim-hero {
    animation: heroFadeIn 0.6s cubic-bezier(0.16, 1, 0.3, 1) both;
  }
  .anim-hero-1 { animation-delay: 0.1s; }
  .anim-hero-2 { animation-delay: 0.2s; }
  @keyframes heroFadeIn {
    from { opacity: 0; transform: translateY(16px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .template-card { display: flex; flex-direction: column; }
  .template-badge {
    font-size: 0.8125rem; font-weight: 600;
    padding: 0.25rem 0.625rem; border-radius: 9999px;
  }
  .template-badge.free {
    background: rgba(16,185,129,0.1); color: var(--success-text);
  }
  .template-badge.owned {
    background: rgba(245,158,11,0.12); color: var(--accent, #d97706);
  }
  .template-badge.price {
    background: rgba(99,102,241,0.1); color: var(--primary);
  }
  .template-use-btn {
    padding: 0.375rem 0.875rem; border-radius: 9999px;
    background: var(--primary); color: white;
    font-size: 0.8125rem; font-weight: 500;
    text-decoration: none; transition: all 0.2s;
  }
  .template-use-btn:hover { opacity: 0.9; }
  .template-buy-btn {
    padding: 0.375rem 0.875rem; border-radius: 9999px;
    background: transparent; color: var(--primary);
    border: 1px solid var(--primary);
    font-size: 0.8125rem; font-weight: 500;
    cursor: pointer; transition: all 0.2s;
  }
  .template-buy-btn:hover {
    background: var(--primary); color: white;
  }
  .template-buy-btn:disabled {
    opacity: 0.5; cursor: not-allowed;
  }
  /* PayPal button container: keep it compact and aligned with the badge. */
  .paypal-template-btn {
    min-width: 140px; max-width: 180px;
  }
  .error-banner {
    max-width: 24rem; margin: 0 auto 1.5rem; padding: 0.75rem 1rem;
    border-radius: var(--radius); background: rgba(239,68,68,0.08);
    border: 1px solid rgba(239,68,68,0.2); color: #dc2626;
    font-size: 0.875rem; text-align: center;
  }
  .success-banner {
    max-width: 24rem; margin: 0 auto 1.5rem; padding: 0.75rem 1rem;
    border-radius: var(--radius); background: rgba(16,185,129,0.08);
    border: 1px solid rgba(16,185,129,0.2); color: var(--success-text);
    font-size: 0.875rem; text-align: center;
  }
  .pro-notice {
    max-width: 24rem; margin: 0 auto 1.5rem; padding: 0.75rem 1rem;
    border-radius: var(--radius); background: rgba(16,185,129,0.08);
    border: 1px solid rgba(16,185,129,0.2); color: var(--success-text);
    font-size: 0.875rem; text-align: center; font-weight: 500;
  }
</style>
