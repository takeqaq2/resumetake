<script>
  import { page } from '$app/stores';
  import { goto, replaceState } from '$app/navigation';
  import { getTranslation } from '$lib/i18n/index.js';
  import { apiFetch, getToken } from '$lib/api.js';
  import { onMount } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  let billingCycle = $state('monthly');
  let error = $state('');
  let paymentSuccess = $state('');
  let success = $state('');
  let paypalRendered = $state(false);
  let redirectTimer;
  let paypalClientId = $state('');
  let paypalContainer = $state();
  // R54b-F3: track auth state so the PayPal button isn't rendered for
  // unauthenticated users — clicking it would just redirect to /auth anyway,
  // which is confusing UX. Show a "log in to upgrade" CTA instead.
  let isLoggedIn = $state(false);
  // R54b-F4: track SDK load failure so we can show a retry button instead
  // of leaving the user with a static error message and no recourse.
  let sdkLoadFailed = $state(false);
  // R51-F3: persist the planId chosen at createOrder time so onApprove
  // uses the same value. Without this, switching billingCycle while the
  // PayPal popup is open causes onApprove to send a different plan than
  // the one the order was created with.
  let pendingPlanId = '';
  // R57b-F1: track mount state so onApprove's async capture call doesn't
  // write $state or set redirectTimer after the user navigates away.
  // Without this, a capture request completing after unmount sets
  // redirectTimer which fires goto() 1.5s later, overriding the user's
  // new navigation (same pattern as auth/templates mounted guard).
  let mounted = false;

  // R39b-F-L4: annual discount percentage shown in the billing toggle badge.
  // R40b-L7: compute from actual prices — $9.99*12=$119.88, annual $79.99,
  // saving $39.89 = 33% off, not 40%. Using annualSave (a dollar amount) as
  // the percentage was inaccurate.
  const ANNUAL_DISCOUNT_PERCENT = Math.round((1 - 79.99 / (9.99 * 12)) * 100);

  let plans = $derived([
    {
      id: 'free',
      name: t.pricing.free,
      price: '0',
      annualPrice: '0',
      period: '',
      features: t.pricing.freeFeatures,
      cta: t.pricing.current,
      highlighted: false,
      disabled: true
    },
    {
      id: 'pro',
      name: t.pricing.pro,
      price: '9.99',
      annualPrice: '79.99',
      annualSave: 40,
      period: t.pricing.billingMonthly,
      annualPeriod: t.pricing.billingAnnual,
      features: t.pricing.proFeatures,
      cta: t.pricing.upgrade,
      highlighted: true,
      disabled: false
    }
  ]);

  async function createOrder(data, actions) {
    // R54-F3: clear stale error from a previous failed attempt so the
    // user doesn't see an old message alongside the new PayPal popup.
    error = '';
    const token = getToken();
    if (!token) {
      goto(`/${lang}/auth`);
      throw new Error('NOT_AUTHENTICATED');
    }
    const planId = billingCycle === 'monthly' ? 'pro' : 'pro_annual';
    pendingPlanId = planId; // R51-F3: persist for onApprove
    try {
      const res = await apiFetch('/api/v1/create-paypal-order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ plan: planId, lang })
      });
      let result;
      try { result = await res.json(); } catch { error = t.pricing.paymentUnavailable; throw new Error('ORDER_CREATE_FAILED'); }
      if (!res.ok || !result.order_id) {
        error = result.message || t.pricing.paymentUnavailable;
        throw new Error(result.message || 'ORDER_CREATE_FAILED');
      }
      return result.order_id;
    } catch (e) {
      if (e.message === 'NOT_AUTHENTICATED' || e.message === 'ORDER_CREATE_FAILED') throw e;
      error = t.pricing.networkError;
      throw new Error('NETWORK_ERROR');
    }
  }

  async function onApprove(data, actions) {
    const token = getToken();
    if (!token) {
      // R37-F1: don't silently return — PayPal has already charged the
      // user. Show an error so they know to re-login and retry capture.
      error = t.pricing.paymentUnavailable || 'Payment could not be completed. Please log in and try again.';
      return;
    }
    error = '';
    try {
      // R51-F3: use the planId from createOrder, not the current billingCycle
      // (which may have changed while the PayPal popup was open).
      const planId = pendingPlanId || (billingCycle === 'monthly' ? 'pro' : 'pro_annual');
      pendingPlanId = '';
      const res = await apiFetch('/api/v1/capture-paypal-order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ order_id: data.orderID, plan: planId })
      });
      // R57b-F1: if the user navigated away during the capture request,
      // don't write $state or set redirectTimer — the goto() would
      // override the user's new navigation 1.5s later.
      if (!mounted) return;
      let result;
      try { result = await res.json(); } catch { error = t.pricing.paymentUnavailable; return; }
      if (res.ok && result.success) {
        paymentSuccess = t.pricing.paymentSuccess;
        if (redirectTimer) clearTimeout(redirectTimer);
        redirectTimer = setTimeout(() => { if (mounted) goto(`/${lang}/editor`); }, 1500);
      } else {
        error = result.message || t.pricing.paymentUnavailable;
      }
    } catch (e) {
      if (mounted) error = t.pricing.networkError;
    }
  }

  function onError(err) {
    // R58-F-L2: guard against async callback firing after unmount.
    if (!mounted) return;
    // Don't overwrite a specific error already set by createOrder (e.g.
    // rate-limit message from the backend) — only set the generic fallback.
    if (!error) error = t.pricing.paymentUnavailable;
  }

  function loadPaypalScript() {
    return new Promise((resolve, reject) => {
      const existing = document.getElementById('paypal-sdk');
      if (existing) {
        // Script tag exists — but it may still be downloading. If
        // window.paypal is already defined, resolve immediately.
        // Otherwise wait for the existing script's onload to avoid
        // a race where renderPaypalButtons sees paypal === undefined
        // and silently never renders.
        if (typeof window.paypal !== 'undefined') {
          resolve();
          return;
        }
        // R40b-M4: if the existing script already fired its error event
        // (e.g. blocked by ad blocker), attaching new load/error listeners
        // will never fire — the Promise hangs forever. Remove the stale
        // tag and fall through to create a fresh one.
        existing.remove();
      }
      const script = document.createElement('script');
      script.id = 'paypal-sdk';
      script.src = `https://www.paypal.com/sdk/js?client-id=${encodeURIComponent(paypalClientId)}&currency=USD&intent=capture`;
      script.async = true;
      // R58-F-M1: add a 15s timeout — if neither onload nor onerror fires
      // (e.g. ad blocker silently drops the request without triggering
      // onerror), the Promise would hang forever, leaving the user stuck
      // on "Loading payment..." with no retry button (sdkLoadFailed stays
      // false). The timeout ensures .catch fires → sdkLoadFailed = true →
      // retry button appears.
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

  function renderPaypalButtons() {
    if (paypalRendered || !paypalClientId) return;
    if (typeof window.paypal === 'undefined') return;
    if (!paypalContainer) return;
    paypalContainer.replaceChildren();
    window.paypal.Buttons({
      style: {
        layout: 'vertical',
        color: 'blue',
        shape: 'rect',
        label: 'pay',
        height: 50
      },
      createOrder,
      onApprove,
      onError,
      onCancel: () => {
        // R58-F-L2: guard against async callback firing after unmount.
        if (!mounted) return;
        // User closed the PayPal popup before completing payment.
        // R48-F4: don't overwrite a specific error already set by createOrder
        // (e.g. rate-limit message) — same guard pattern as onError.
        if (!error) error = t.pricing.paymentCancelled;
      }
    }).render(paypalContainer)
      .then(() => { paypalRendered = true; })
      .catch(() => {
        // Render can fail if the container was removed mid-render or the SDK
        // has an internal error. Show a user-facing message and allow retry.
        // R51-F3: set sdkLoadFailed so the retry button appears — previously
        // only SDK *load* failure set this flag, leaving render-failure
        // users stuck with an error message and no retry option.
        error = t.pricing.paymentUnavailable;
        paypalRendered = false;
        sdkLoadFailed = true;
      });
  }

  let sdkObserver = null;
  let sdkLoadTimeout = null;
  let paypalInitStarted = $state(false);
  onMount(() => {
    mounted = true;
    let cancelled = false;
    // R54b-F3: check auth state on mount so the PayPal button is only
    // rendered for logged-in users.
    isLoggedIn = !!getToken();
    // R57b-F3: sync isLoggedIn across tabs — if the user logs in/out in
    // another tab, the pricing page should update its CTA without requiring
    // a refresh. Mirrors the pattern in +layout.svelte.
    const onAuthChange = () => { try { isLoggedIn = !!getToken(); } catch (e) {} };
    window.addEventListener('storage', onAuthChange);
    window.addEventListener('auth:unauthorized', onAuthChange);
    try {
      const saved = localStorage.getItem('billingCycle');
      if (saved === 'annual') billingCycle = 'annual';
    } catch (e) {}
    const params = new URLSearchParams(window.location.search);
    if (params.get('payment') === 'success') {
      success = t.pricing.paymentSuccess;
      // Use SvelteKit's replaceState (not window.history.replaceState) so
      // $page.url stays in sync — otherwise the stale ?payment=success query
      // persists in $page store and could re-trigger this logic on back/nav.
      replaceState(window.location.pathname, {});
    } else if (params.get('payment') === 'cancelled') {
      error = t.pricing.paymentCancelled;
      replaceState(window.location.pathname, {});
    }

    apiFetch('/api/config', { skipAuth: true })
      .then(r => r.json())
      .then(data => {
        if (cancelled) return;
        if (data.paypal_client_id) {
          paypalClientId = data.paypal_client_id;
        }
      })
      // R51-F4: surface config load failure to the user instead of silently
      // leaving the PayPal button disabled with no explanation.
      .catch(() => {
        if (!cancelled) error = t.pricing.paymentUnavailable;
      });

    return () => {
      cancelled = true;
      mounted = false;
      window.removeEventListener('storage', onAuthChange);
      window.removeEventListener('auth:unauthorized', onAuthChange);
      if (sdkObserver) sdkObserver.disconnect();
      clearTimeout(sdkLoadTimeout);
      if (redirectTimer) clearTimeout(redirectTimer);
    };
  });

  // Initialize PayPal SDK lazy-loading via $effect — reacts to both
  // paypalClientId (from async API call) and paypalContainer (from bind:this
  // after {#if paypalClientId} renders). This replaces the prior tick().then()
  // approach which had a race condition: bind:this may not be assigned yet
  // when tick() resolves, causing paypalContainer to be undefined and the
  // SDK to never load.
  $effect(() => {
    if (!paypalClientId || !paypalContainer || paypalInitStarted) return;
    paypalInitStarted = true;

    if ('IntersectionObserver' in window) {
      sdkObserver = new IntersectionObserver((entries) => {
        if (entries.some(e => e.isIntersecting)) {
          sdkObserver.disconnect();
          sdkObserver = null;
          loadPaypalScript().then(() => {
            sdkLoadTimeout = setTimeout(renderPaypalButtons, 300);
          }).catch(() => {
            // R47-F2: do NOT reset paypalInitStarted — doing so re-triggers
            // this $effect (paypalInitStarted is a dependency), causing an
            // infinite retry loop that spams failed script loads if the SDK
            // is blocked (e.g. by an ad blocker). The error message is
            // already shown to the user.
            // R54b-F4: set sdkLoadFailed so the retry button appears.
            error = t.pricing.paymentLoadFailed;
            sdkLoadFailed = true;
          });
        }
      }, { rootMargin: '100px' });
      sdkObserver.observe(paypalContainer);
    } else {
      loadPaypalScript().then(() => {
        sdkLoadTimeout = setTimeout(renderPaypalButtons, 300);
      }).catch(() => {
        // R47-F2: same as above — don't reset paypalInitStarted.
        // R54b-F4: set sdkLoadFailed so the retry button appears.
        error = t.pricing.paymentLoadFailed;
        sdkLoadFailed = true;
      });
    }
  });

  // Re-render PayPal buttons when billing cycle changes (monthly/annual).
  // Uses $effect with cleanup to prevent setTimeout firing after unmount.
  // Also clears sdkLoadTimeout from onMount — without this, both timers can
  // be pending simultaneously, causing renderPaypalButtons to fire twice and
  // produce a race condition (one render destroys the other's DOM mid-flight).
  let renderTimeout;
  $effect(() => {
    if (billingCycle) {
      paypalRendered = false;
      if (typeof window.paypal !== 'undefined') {
        clearTimeout(renderTimeout);
        clearTimeout(sdkLoadTimeout);
        renderTimeout = setTimeout(renderPaypalButtons, 100);
      }
    }
    return () => clearTimeout(renderTimeout);
  });

  // R54b-F4: manual retry for SDK load failure. Resets the init guard so
  // the $effect above re-runs loadPaypalScript. The previous design left
  // the user with a static error and no way to recover without a full
  // page reload (problematic if the failure was transient, e.g. the SDK
  // CDN was temporarily blocked by a network blip).
  function retrySdkLoad() {
    sdkLoadFailed = false;
    error = '';
    paypalInitStarted = false;
    paypalRendered = false;
  }
</script>

<svelte:head>
  <title>{t.meta.pricingTitle}</title>
  <meta name="description" content={t.meta.pricingDesc}>
  <meta name="robots" content="index, follow">
  <link rel="canonical" href="https://resume.takee.top/{lang}/pricing">
  <meta property="og:title" content={t.meta.pricingTitle}>
  <meta property="og:description" content={t.meta.pricingDesc}>
  <meta property="og:url" content="https://resume.takee.top/{lang}/pricing">
  <meta property="og:type" content="website">
</svelte:head>

<div class="pricing-page">
  <div class="pricing-header">
    <div class="orb orb-blue animate-float" aria-hidden="true" style="width:250px;height:250px;top:-15%;left:5%"></div>
    <div class="orb orb-purple animate-float" aria-hidden="true" style="width:200px;height:200px;bottom:-10%;right:10%;animation-delay:2s"></div>
    <div class="container" style="position:relative;text-align:center">
      <span class="section-badge"><span aria-hidden="true">💎</span> {t.pricing.title}</span>
      <h1 style="font-size:clamp(1.75rem,4vw,2.5rem);font-weight:800;margin:1rem 0 0.75rem">
        {t.pricing.choosePlan}
      </h1>
      <p style="color:var(--text-secondary);font-size:1rem;max-width:32rem;margin:0 auto">{t.pricing.subtitle}</p>
    </div>
  </div>

  <div class="container" style="padding:3rem 1.5rem;margin-top:-2rem">
    {#if error}
      <div class="error-banner" role="alert">{error}</div>
    {/if}
    {#if paymentSuccess}
      <div class="success-banner" role="status">{paymentSuccess}</div>
    {/if}
    {#if success}
      <div class="success-banner" role="status">{success}</div>
    {/if}

    <div class="billing-toggle" role="group" aria-label={t.pricing.choosePlan}>
      <button class="toggle-btn" class:active={billingCycle === 'monthly'} aria-pressed={billingCycle === 'monthly'} onclick={() => { billingCycle = 'monthly'; try { localStorage.setItem('billingCycle', 'monthly'); } catch (e) {} }}>
        {t.pricing.billingMonthly}
      </button>
      <button class="toggle-btn" class:active={billingCycle === 'annual'} aria-pressed={billingCycle === 'annual'} onclick={() => { billingCycle = 'annual'; try { localStorage.setItem('billingCycle', 'annual'); } catch (e) {} }}>
        {t.pricing.billingAnnual} <span class="save-badge">-{ANNUAL_DISCOUNT_PERCENT}%</span>
      </button>
    </div>

    <div class="pricing-grid">
      {#each plans as plan (plan.id)}
        <div class="pricing-card {plan.highlighted ? 'highlighted' : ''}">
          {#if plan.highlighted}
            <div class="popular-badge">{t.pricing.popular}</div>
          {/if}
          <div class="plan-header">
            <h3 class="plan-name">{plan.name}</h3>
            <div class="plan-price">
              {#if plan.id === 'free'}
                <span class="price-value">$0</span>
              {:else}
                <span class="price-currency">$</span>
                <span class="price-value">{billingCycle === 'monthly' ? plan.price : plan.annualPrice}</span>
                <span class="price-period">/{billingCycle === 'monthly' ? t.pricing.monthly : t.pricing.annual}</span>
              {/if}
            </div>
            {#if plan.annualSave && billingCycle === 'annual'}
              <div class="save-text">{t.pricing.save} ${plan.annualSave}</div>
            {/if}
          </div>
          <ul class="plan-features">
            {#each plan.features as feature (feature)}
              <li>
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true"><path d="M3 8l3.5 3.5L13 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                {feature}
              </li>
            {/each}
          </ul>
          {#if plan.highlighted}
            {#if !isLoggedIn}
              <a href="/{lang}/auth" class="btn btn-primary plan-cta">{t.pricing.loginToUpgrade}</a>
            {:else if paypalClientId}
              <div id="paypal-button-container" class="paypal-btn-wrap" bind:this={paypalContainer}>
                {#if !paypalRendered && !error}
                  <div class="paypal-loading-placeholder" role="status" aria-live="polite">{t.pricing.loadingPayment}</div>
                {/if}
                {#if sdkLoadFailed}
                  <button class="btn btn-secondary" onclick={retrySdkLoad}>{t.pricing.retry}</button>
                {/if}
              </div>
            {:else}
              <button class="btn btn-primary plan-cta" disabled>
                {t.pricing.paymentUnavailable}
              </button>
            {/if}
          {:else}
            <button class="btn btn-secondary plan-cta" disabled={plan.disabled}>
              {plan.cta}
            </button>
          {/if}
        </div>
      {/each}
    </div>

    <!-- Single-use AI Products -->
    <div class="products-section">
      <h2 class="products-title">{t.pricing.productsTitle}</h2>
      <p class="products-subtitle">{t.pricing.productsSubtitle}</p>
      <div class="products-grid">
        {#each t.pricing.products as product (product.name)}
          <div class="product-card">
            <div class="product-icon" aria-hidden="true">{product.icon}</div>
            <h3>{product.name}</h3>
            <p class="product-price">{product.price}</p>
            <p class="product-desc">{product.desc}</p>
          </div>
        {/each}
      </div>
    </div>
  </div>
</div>

<style>
  .pricing-page { overflow: hidden; }
  .pricing-header {
    position: relative; padding: 5rem 0 6rem;
    background: var(--gradient-hero); background-size: 200% 200%;
    animation: gradientShift 10s ease-in-out infinite;
  }
  .section-badge {
    display: inline-flex; align-items: center; gap: 0.375rem;
    padding: 0.375rem 0.875rem; border-radius: 9999px;
    background: var(--bg-glass); border: 1px solid var(--border);
    font-size: 0.8125rem; font-weight: 500; color: var(--primary);
    backdrop-filter: blur(8px);
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
    border: 1px solid rgba(16,185,129,0.2); color: #047857;
    font-size: 0.875rem; text-align: center;
  }
  .billing-toggle {
    display: flex; justify-content: center; gap: 0.25rem;
    margin-bottom: 2.5rem;
    background: var(--bg-glass); border: 1px solid var(--border);
    border-radius: 9999px; padding: 0.25rem;
    margin-left: auto; margin-right: auto;
  }
  .pricing-header .container { display: flex; flex-direction: column; align-items: center; }
  .toggle-btn {
    padding: 0.5rem 1.25rem; border-radius: 9999px; border: none;
    background: transparent; color: var(--text-secondary);
    font-size: 0.875rem; font-weight: 500; cursor: pointer;
    transition: all 0.2s; display: inline-flex; align-items: center; gap: 0.375rem;
  }
  .toggle-btn.active {
    background: var(--primary); color: white;
  }
  .save-badge {
    background: rgba(16,185,129,0.15); color: #047857;
    padding: 0.125rem 0.5rem; border-radius: 9999px;
    font-size: 0.75rem; font-weight: 600;
  }
  .pricing-grid {
    display: grid; grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem; align-items: stretch; max-width: 40rem; margin: 0 auto;
  }
  .pricing-card {
    background: var(--bg-glass); border: 1px solid var(--border);
    border-radius: var(--radius-lg); padding: 2rem;
    backdrop-filter: blur(16px); position: relative;
    display: flex; flex-direction: column;
    transition: all 0.35s cubic-bezier(0.4,0,0.2,1);
  }
  .pricing-card:hover {
    transform: translateY(-4px); box-shadow: var(--shadow-lg);
  }
  .pricing-card.highlighted {
    border-color: var(--primary);
    box-shadow: 0 8px 40px var(--primary-glow);
    transform: scale(1.02);
  }
  .pricing-card.highlighted:hover {
    transform: scale(1.02) translateY(-4px);
  }
  .popular-badge {
    position: absolute; top: -0.75rem; left: 50%; transform: translateX(-50%);
    padding: 0.25rem 1rem; border-radius: 9999px;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    color: white; font-size: 0.75rem; font-weight: 600;
    white-space: nowrap;
  }
  .plan-header { text-align: center; margin-bottom: 1.5rem; }
  .plan-name {
    font-size: 1.125rem; font-weight: 700; color: var(--text);
    margin-bottom: 0.75rem;
  }
  .plan-price { display: flex; align-items: baseline; justify-content: center; gap: 0.125rem; }
  .price-currency {
    font-size: 1.5rem; font-weight: 700; color: var(--text);
    align-self: flex-start; margin-top: 0.5rem;
  }
  .price-value {
    font-size: 2.75rem; font-weight: 800; color: var(--text);
    letter-spacing: -0.03em;
  }
  .price-period { font-size: 0.875rem; color: var(--text-secondary); }
  .save-text {
    text-align: center; margin-top: 0.5rem;
    font-size: 0.8125rem; color: #047857; font-weight: 600;
  }
  .plan-features {
    list-style: none; display: flex; flex-direction: column; gap: 0.75rem;
    margin-bottom: 2rem; flex: 1;
  }
  .plan-features li {
    display: flex; align-items: center; gap: 0.625rem;
    font-size: 0.875rem; color: var(--text-secondary);
  }
  .plan-features svg { color: var(--success-text); flex-shrink: 0; }
  .plan-cta { width: 100%; padding: 0.875rem; font-weight: 600; }
  .paypal-btn-wrap {
    width: 100%;
    min-height: 50px;
    display: flex;
    justify-content: center;
    align-items: center;
  }
  .paypal-loading-placeholder {
    display: flex; align-items: center; justify-content: center;
    width: 100%; padding: 0.875rem; border-radius: var(--radius);
    background: var(--bg-surface); color: var(--text-secondary);
    font-size: 0.8125rem; font-weight: 500;
  }
  @media (max-width: 768px) {
    .pricing-grid { grid-template-columns: 1fr; max-width: 24rem; margin: 0 auto; }
    .pricing-card.highlighted { transform: none; }
    .pricing-card.highlighted:hover { transform: translateY(-4px); }
  }

  .products-section {
    margin-top: 4rem; text-align: center;
  }
  .products-title {
    font-size: 1.5rem; font-weight: 700; color: var(--text);
    margin-bottom: 0.5rem;
  }
  .products-subtitle {
    font-size: 0.875rem; color: var(--text-secondary);
    margin-bottom: 2rem;
  }
  .products-grid {
    display: grid; grid-template-columns: repeat(4, 1fr);
    gap: 1rem; max-width: 56rem; margin: 0 auto;
  }
  .product-card {
    background: var(--bg-glass); border: 1px solid var(--border);
    border-radius: var(--radius); padding: 1.5rem 1rem;
    backdrop-filter: blur(8px);
    transition: all 0.3s;
  }
  .product-card:hover {
    transform: translateY(-2px); box-shadow: var(--shadow-md);
    border-color: var(--primary);
  }
  .product-icon {
    font-size: 2rem; margin-bottom: 0.75rem;
  }
  .product-card h3 {
    font-size: 0.9375rem; font-weight: 600; color: var(--text);
    margin-bottom: 0.375rem;
  }
  .product-price {
    font-size: 1.25rem; font-weight: 700; color: var(--primary);
    margin-bottom: 0.5rem;
  }
  .product-desc {
    font-size: 0.8125rem; color: var(--text-secondary);
    line-height: 1.4;
  }
  @media (max-width: 768px) {
    .products-grid { grid-template-columns: repeat(2, 1fr); }
  }
</style>
