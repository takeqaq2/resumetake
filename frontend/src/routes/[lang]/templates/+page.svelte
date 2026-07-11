<script>
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
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
  // R55b-F3: track mount state so purchaseTemplate's redirectTimer callback
  // doesn't fire goto() after the user has navigated away.
  let mounted = false;

  // R49-F1: display price for paid templates. Source of truth is
  // backend/handlers/product.go templatePrices map (all $2.99 USD).
  // If backend prices change, update this constant to match.
  // Payment is always charged in USD via PayPal regardless of locale.
  const templatePrice = '$2.99';

  // Svelte 5 declarative pattern: $effect runs client-side only, replaces
  // the old onMount querySelectorAll('.reveal') approach. bind:this collects
  // element refs without manual DOM lookup.
  $effect(() => {
    const cards = cardRefs.filter(Boolean);
    // R27-M2: wait until ALL card refs are populated before creating the
    // observer. Without this guard, each bind:this assignment triggers the
    // effect, creating+destroying N-1 observers before the final one sticks.
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
    if (token) {
      apiFetch("/api/v1/auth/me", { skipAuth: true }).then(res => res.ok ? res.json() : null).then(data => {
        if (cancelled) return;
        if (data?.data) user = data.data;
        userLoading = false;
      }).catch(() => { if (!cancelled) userLoading = false; });
    } else {
      userLoading = false;
    }
    return () => { cancelled = true; mounted = false; clearTimeout(redirectTimer); };
  });

  function isTemplateFree(tplId, index) {
    if (user?.plan === "pro" || user?.plan === "enterprise") return true;
    // R38b-L1: judge by template ID, not array index. "professional" is the
    // free template; relying on index===0 would break if i18n ever reorders
    // the items array. The index param is kept for backwards compat but
    // no longer determines free-ness.
    if (tplId === 'professional') return true;
    if (user?.purchased_templates?.includes(tplId)) return true;
    return false;
  }

  async function purchaseTemplate(tplId) {
    // R49-F3: guard against concurrent purchases. `purchasing` is set to
    // tplId when a purchase is in-flight, but the button's `disabled` attr
    // only disables the CURRENT template's button — users can click another
    // template's buy button while the first request is still pending,
    // creating duplicate PayPal orders. This early return prevents that.
    if (purchasing) return;
    const token = getToken();
    if (!token) {
      // R55-F3: carry redirect param so user returns to templates after login
      goto(`/${lang}/auth?redirect=${encodeURIComponent(`/${lang}/templates`)}`);
      return;
    }
    purchasing = tplId;
    error = '';
    success = '';
    let keepLock = false;
    try {
      // R52b-F1: Only call the API if the user has a plan that might include
      // this template (free/included case). For paid users without the
      // template, redirect to pricing WITHOUT creating an orphaned PayPal
      // order — the previous flow created an order via /purchase-template
      // but never captured it.
      if (user && (user.plan === 'pro' || user.plan === 'enterprise')) {
        const res = await apiFetch("/api/v1/purchase-template", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ template_id: tplId, lang }),
        });
        let result;
        try {
          result = await res.json();
        } catch {
          error = t.templates.networkError;
          return;
        }
        if (res.ok && result.success) {
          const data = result.data || {};
          if (data.order_id) {
            // Should not happen for Pro/Enterprise, but handle gracefully.
            success = t.templates.paymentRequired;
            keepLock = true;
            if (redirectTimer) clearTimeout(redirectTimer);
            redirectTimer = setTimeout(() => {
              if (!mounted) return;
              goto(`/${lang}/pricing`);
              purchasing = '';
            }, 1500);
          } else {
            // Free or included in plan — mark as purchased locally.
            if (user) {
              const purchased = user.purchased_templates || [];
              if (!purchased.includes(tplId)) purchased.push(tplId);
              user = { ...user, purchased_templates: purchased };
            }
            success = t.templates.purchaseSuccess;
          }
        } else {
          error = result.message || t.templates.purchaseFailed;
        }
      } else {
        // Free user — redirect to pricing to subscribe (Pro includes all templates).
        success = t.templates.paymentRequired;
        keepLock = true;
        if (redirectTimer) clearTimeout(redirectTimer);
        redirectTimer = setTimeout(() => {
          if (!mounted) return;
          goto(`/${lang}/pricing`);
          purchasing = '';
        }, 1500);
      }
    } catch (e) {
      error = t.templates.networkError;
    } finally {
      if (mounted && !keepLock) purchasing = '';
    }
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
      {@const isFree = isTemplateFree(tpl.id, i)}
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
          {:else if isFree}
            <span class="template-badge free">{t.templates.freeBadge}</span>
          {:else}
            <span class="template-badge price">{templatePrice}</span>
          {/if}
          {#if userLoading}
            <button class="template-use-btn" disabled style="opacity:0.5">…</button>
          {:else if isFree}
            <a href="/{lang}/editor" class="template-use-btn">{t.templates.use}</a>
          {:else}
            <button class="template-buy-btn" onclick={() => purchaseTemplate(tpl.id)} disabled={!!purchasing}>
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
