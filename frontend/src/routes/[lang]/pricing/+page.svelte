<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  let plans = $derived([
    {
      id: 'free',
      name: t.pricing.free,
      price: '0',
      period: '',
      features: [
        lang === 'zh' ? '5次AI优化/月' : '5 AI optimizations/month',
        lang === 'zh' ? '基础模板' : 'Basic templates',
        lang === 'zh' ? '校招信息' : 'Campus recruitment info'
      ],
      cta: t.pricing.current,
      highlighted: false,
      disabled: true
    },
    {
      id: 'pro',
      name: t.pricing.pro,
      price: '¥29',
      period: t.pricing.monthly,
      features: [
        lang === 'zh' ? '无限AI优化' : 'Unlimited AI optimizations',
        lang === 'zh' ? '高级模板' : 'Premium templates',
        lang === 'zh' ? '4视角分析' : '4-perspective analysis',
        lang === 'zh' ? '0基础生成' : '0-basis resume generation',
        lang === 'zh' ? '简历导出' : 'Resume export'
      ],
      cta: t.pricing.upgrade,
      highlighted: true,
      disabled: false
    },
    {
      id: 'enterprise',
      name: t.pricing.enterprise,
      price: '¥99',
      period: t.pricing.monthly,
      features: [
        lang === 'zh' ? 'API接入' : 'API access',
        lang === 'zh' ? '团队管理' : 'Team management',
        lang === 'zh' ? '自定义模板' : 'Custom templates'
      ],
      cta: t.pricing.upgrade,
      highlighted: false,
      disabled: false
    }
  ]);
</script>

<svelte:head>
  <title>{t.pricing.title} - ResumeTake</title>
</svelte:head>

<div class="pricing-page">
  <div class="pricing-header">
    <div class="orb orb-blue animate-float" style="width:250px;height:250px;top:-15%;left:5%"></div>
    <div class="orb orb-purple animate-float" style="width:200px;height:200px;bottom:-10%;right:10%;animation-delay:2s"></div>
    <div class="container" style="position:relative;text-align:center">
      <span class="section-badge">💎 {t.pricing.title}</span>
      <h1 style="font-size:clamp(1.75rem,4vw,2.5rem);font-weight:800;margin:1rem 0 0.75rem">
        {lang === 'zh' ? '选择适合你的方案' : 'Choose Your Plan'}
      </h1>
      <p style="color:var(--text-secondary);font-size:1rem;max-width:32rem;margin:0 auto">{t.pricing.subtitle}</p>
    </div>
  </div>

  <div class="container" style="padding:3rem 1.5rem;margin-top:-2rem">
    <div class="pricing-grid">
      {#each plans as plan}
        <div class="pricing-card {plan.highlighted ? 'highlighted' : ''}">
          {#if plan.highlighted}
            <div class="popular-badge">{t.pricing.popular}</div>
          {/if}
          <div class="plan-header">
            <h3 class="plan-name">{plan.name}</h3>
            <div class="plan-price">
              <span class="price-value">{plan.price}</span>
              {#if plan.period}
                <span class="price-period">/{plan.period}</span>
              {/if}
            </div>
          </div>
          <ul class="plan-features">
            {#each plan.features as feature}
              <li>
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M3 8l3.5 3.5L13 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                {feature}
              </li>
            {/each}
          </ul>
          <button class="btn {plan.highlighted ? 'btn-primary' : 'btn-secondary'} plan-cta" disabled={plan.disabled}>
            {plan.cta}
          </button>
        </div>
      {/each}
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
  .pricing-grid {
    display: grid; grid-template-columns: repeat(3, 1fr);
    gap: 1.5rem; align-items: stretch;
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
  .plan-price { display: flex; align-items: baseline; justify-content: center; gap: 0.25rem; }
  .price-value {
    font-size: 2.5rem; font-weight: 800; color: var(--text);
    letter-spacing: -0.03em;
  }
  .price-period { font-size: 0.875rem; color: var(--text-secondary); }
  .plan-features {
    list-style: none; display: flex; flex-direction: column; gap: 0.75rem;
    margin-bottom: 2rem; flex: 1;
  }
  .plan-features li {
    display: flex; align-items: center; gap: 0.625rem;
    font-size: 0.875rem; color: var(--text-secondary);
  }
  .plan-features svg { color: #10b981; flex-shrink: 0; }
  .plan-cta { width: 100%; padding: 0.875rem; font-weight: 600; }
  @media (max-width: 768px) {
    .pricing-grid { grid-template-columns: 1fr; max-width: 24rem; margin: 0 auto; }
    .pricing-card.highlighted { transform: none; }
    .pricing-card.highlighted:hover { transform: translateY(-4px); }
  }
</style>
