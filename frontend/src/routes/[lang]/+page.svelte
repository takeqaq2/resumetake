<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';
  import { onMount } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));
  let scrollProgress = $state(0);

  onMount(() => {
    const onScroll = () => {
      const h = document.documentElement.scrollHeight - window.innerHeight;
      scrollProgress = h > 0 ? (window.scrollY / h) * 100 : 0;
    };
    window.addEventListener('scroll', onScroll, { passive: true });

    document.querySelectorAll('.reveal').forEach(el => el.classList.add('js-ready'));
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(e => { if (e.isIntersecting) { e.target.classList.add('visible'); observer.unobserve(e.target); } });
    }, { threshold: 0.1 });
    document.querySelectorAll('.reveal.js-ready').forEach(el => observer.observe(el));

    return () => { window.removeEventListener('scroll', onScroll); observer.disconnect(); };
  });
</script>

<svelte:head>
  <title>{t.meta.title}</title>
  <meta name="description" content={t.meta.description}>
  <meta name="keywords" content={t.meta.keywords}>
  <link rel="canonical" href="https://resume.takee.top/{lang}">
  <meta property="og:title" content={t.meta.title}>
  <meta property="og:description" content={t.meta.description}>
  <meta property="og:url" content="https://resume.takee.top/{lang}">
  <meta property="og:type" content="website">
  <meta property="og:image" content="https://resume.takee.top/og-image.png">
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content={t.meta.title}>
  <meta name="twitter:description" content={t.meta.description}>
  {@html `<script type="application/ld+json">${JSON.stringify({
    "@context":"https://schema.org",
    "@type":"WebApplication",
    "name":"ResumeTake - " + (lang === 'zh' ? 'AI简历优化工具' : lang === 'ja' ? 'AI履歴書作成ツール' : lang === 'ko' ? 'AI 이력서 작성 도구' : lang === 'ar' ? 'منشئ السيرة الذاتية بالذكاء الاصطناعي' : 'AI Resume Builder'),
    "url":"https://resume.takee.top/" + lang,
    "description": t.meta.description,
    "applicationCategory":"BusinessApplication",
    "operatingSystem":"Web",
    "inLanguage": lang,
    "offers":{"@type":"Offer","price":"0","priceCurrency": lang === 'zh' ? 'CNY' : lang === 'ja' ? 'JPY' : lang === 'ko' ? 'KRW' : lang === 'es' ? 'EUR' : lang === 'pt' ? 'BRL' : lang === 'fr' ? 'EUR' : lang === 'de' ? 'EUR' : lang === 'ar' ? 'SAR' : lang === 'hi' ? 'INR' : 'USD'},
    "featureList": t.features.items.map(i => i.title)
  })}</script>`}
  {@html `<script type="application/ld+json">${JSON.stringify({
    "@context":"https://schema.org",
    "@type":"FAQPage",
    "mainEntity": t.faq.map(f => ({
      "@type":"Question",
      "name": f.q,
      "acceptedAnswer":{"@type":"Answer","text":f.a}
    }))
  })}</script>`}
</svelte:head>

<div class="scroll-progress" style="width:{scrollProgress}%"></div>

<section class="hero-mesh" style="min-height:92vh;display:flex;align-items:center;justify-content:center;position:relative">
  <div class="orb orb-blue animate-float" style="width:300px;height:300px;top:10%;left:5%"></div>
  <div class="orb orb-purple animate-float" style="width:250px;height:250px;top:60%;right:8%;animation-delay:2s"></div>
  <div class="orb orb-pink animate-float" style="width:200px;height:200px;bottom:15%;left:15%;animation-delay:4s"></div>
  <div class="orb orb-green animate-float" style="width:180px;height:180px;top:20%;right:25%;animation-delay:1s"></div>

  <div class="container hero-content" style="position:relative;text-align:center;padding:6rem 1.5rem 7rem">
    <div class="hero-badge anim-hero anim-hero-1">
      <span style="font-size:0.9375rem">✨</span> {t.hero.badge}
    </div>

    <h1 class="anim-hero anim-hero-2" style="font-size:clamp(2.5rem,5.5vw,4.2rem);font-weight:800;line-height:1.08;margin-bottom:1.5rem;letter-spacing:-0.03em;color:var(--text)">
      {t.hero.title1}<br>
      <span class="gradient-text">{t.hero.title2}</span>
    </h1>

    <p class="anim-hero anim-hero-3" style="font-size:1.125rem;color:var(--text-secondary);max-width:38rem;margin:0 auto 2.5rem;line-height:1.7">
      {t.hero.subtitle}
    </p>

    <div class="anim-hero anim-hero-4" style="display:flex;gap:1rem;justify-content:center;flex-wrap:wrap">
      <a href="/{lang}/editor" class="btn btn-primary" style="padding:0.875rem 2.25rem;font-size:1rem;font-weight:600">
        <span>{t.hero.cta}</span>
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" style="transition:transform 0.3s"><path d="M3 8h10M9 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </a>
      <a href="#features" class="btn btn-secondary" style="padding:0.875rem 2.25rem;font-size:1rem">{t.hero.learnMore}</a>
    </div>

    <div class="anim-hero anim-hero-5" style="margin-top:3rem">
      <div style="display:flex;gap:1.5rem;justify-content:center;flex-wrap:wrap;opacity:0.5">
        {#each ['ATS', 'AI', 'PDF', 'SEO'] as tag}
          <span style="padding:0.25rem 0.75rem;border-radius:9999px;background:var(--bg-glass);border:1px solid var(--border);font-size:0.75rem;color:var(--text-secondary);backdrop-filter:blur(8px)">{tag}</span>
        {/each}
      </div>
    </div>
  </div>
</section>

<section id="features" style="padding:6rem 0;position:relative">
  <div class="container">
    <div class="reveal" style="text-align:center;margin-bottom:4rem">
      <div class="hero-badge" style="margin-bottom:1rem">
        <span style="font-size:0.875rem">🚀</span> {t.features.title}
      </div>
      <h2 style="font-size:clamp(1.75rem,3.5vw,2.25rem);font-weight:700;margin-bottom:0.75rem;color:var(--text)">
        {t.features.subtitle}
      </h2>
    </div>
    <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:1.5rem">
      {#each t.features.items as f, i}
        <div class="feature-card reveal" style="transition-delay:{i * 0.1}s">
          <div class="feature-icon">{f.icon}</div>
          <h3 style="font-weight:600;margin-bottom:0.5rem;color:var(--text);font-size:1.0625rem">{f.title}</h3>
          <p style="color:var(--text-secondary);font-size:0.9375rem;line-height:1.65">{f.desc}</p>
        </div>
      {/each}
    </div>
  </div>
</section>

<section id="cta" style="padding:4rem 0">
  <div class="container">
    <div class="cta-section reveal">
      <div style="position:relative;z-index:1">
        <h2 style="font-size:clamp(1.5rem,3vw,2rem);font-weight:700;margin-bottom:0.75rem">{t.cta.title}</h2>
        <p style="opacity:0.9;margin-bottom:2rem;max-width:32rem;margin-left:auto;margin-right:auto">{t.cta.subtitle}</p>
        <a href="/{lang}/editor" class="btn btn-primary" style="padding:0.875rem 2.5rem;font-size:1rem;font-weight:600">
          {t.cta.button}
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M3 8h10M9 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </a>
      </div>
    </div>
  </div>
</section>

<style>
  .anim-hero {
    animation: heroFadeIn 0.7s cubic-bezier(0.16, 1, 0.3, 1) both;
  }
  .anim-hero-1 { animation-delay: 0.1s; }
  .anim-hero-2 { animation-delay: 0.2s; }
  .anim-hero-3 { animation-delay: 0.35s; }
  .anim-hero-4 { animation-delay: 0.5s; }
  .anim-hero-5 { animation-delay: 0.65s; }
  @keyframes heroFadeIn {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
  }
  .reveal {
    opacity: 0;
    transform: translateY(30px);
    transition: all 0.7s cubic-bezier(0.16, 1, 0.3, 1);
  }
  .reveal.visible {
    opacity: 1;
    transform: translateY(0);
  }
</style>
