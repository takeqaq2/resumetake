<script>
  import { page } from '$app/stores';
  import { getTranslation, LANGUAGES } from '$lib/i18n/index.js';
  import { onMount } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));
  let mounted = $state(false);
  onMount(() => { mounted = true; });
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

<section style="position:relative;overflow:hidden">
  <div style="position:absolute;inset:0;background:linear-gradient(135deg,rgba(59,130,246,0.05) 0%,rgba(139,92,246,0.05) 100%)"></div>
  <div class="container" style="position:relative;text-align:center;padding:6rem 1.5rem 7rem">
    <div class="hero-badge" style="display:inline-flex;align-items:center;gap:0.5rem;padding:0.375rem 1rem;border-radius:9999px;font-size:0.8125rem;font-weight:500;margin-bottom:2rem">
      <span style="font-size:0.875rem">✨</span> {t.hero.badge}
    </div>
    <h1 style="font-size:clamp(2.5rem,5.5vw,4rem);font-weight:800;line-height:1.08;margin-bottom:1.5rem;letter-spacing:-0.02em;color:var(--text)">
      {t.hero.title1}<br>
      <span class="gradient-text" style="background:linear-gradient(135deg,var(--primary),var(--accent));-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text">{t.hero.title2}</span>
    </h1>
    <p style="font-size:1.125rem;color:var(--text-secondary);max-width:36rem;margin:0 auto 2.5rem;line-height:1.7">
      {t.hero.subtitle}
    </p>
    <div style="display:flex;gap:1rem;justify-content:center;flex-wrap:wrap">
      <a href="/{lang}/editor" class="btn btn-primary" style="padding:0.875rem 2rem;font-size:1rem">{t.hero.cta}</a>
      <a href="#features" class="btn btn-secondary" style="padding:0.875rem 2rem;font-size:1rem">{t.hero.learnMore}</a>
    </div>
  </div>
</section>

<section id="features" style="padding:5rem 0;background:var(--bg-surface)">
  <div class="container">
    <div style="text-align:center;margin-bottom:3rem">
      <h2 style="font-size:2rem;font-weight:700;margin-bottom:0.75rem;color:var(--text)">{t.features.title}</h2>
      <p style="color:var(--text-secondary);max-width:32rem;margin:0 auto">{t.features.subtitle}</p>
    </div>
    <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:1.5rem">
      {#each t.features.items as f}
        <div class="card" style="text-align:center;padding:2rem;transition:transform 0.2s,box-shadow 0.2s;cursor:default">
          <div style="font-size:2.5rem;margin-bottom:1rem">{f.icon}</div>
          <h3 style="font-weight:600;margin-bottom:0.5rem;color:var(--text)">{f.title}</h3>
          <p style="color:var(--text-secondary);font-size:0.9375rem;line-height:1.6">{f.desc}</p>
        </div>
      {/each}
    </div>
  </div>
</section>

<section style="padding:5rem 0">
  <div class="container" style="text-align:center">
    <h2 style="font-size:2rem;font-weight:700;margin-bottom:0.75rem;color:var(--text)">{t.cta.title}</h2>
    <p style="color:var(--text-secondary);margin-bottom:2.5rem">{t.cta.subtitle}</p>
    <a href="/{lang}/editor" class="btn btn-primary" style="padding:0.875rem 2.5rem;font-size:1rem">{t.cta.button}</a>
  </div>
</section>
