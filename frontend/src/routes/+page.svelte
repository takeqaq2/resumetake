<script>
  import { onMount } from 'svelte';
  import { LANGUAGES } from '$lib/i18n/index.js';

  let suggestedLang = $state('en');

  const markets = {
    zh: {
      headline: '面向中国求职者的 AI 简历优化',
      sub: '突出项目成果、量化指标与 ATS 关键词，适配互联网、校招与社招场景。',
      tone: '效率、结果导向、信息密度高',
      gradient: 'linear-gradient(135deg,#dc2626,#d97706)',
      bg: 'rgba(220,38,38,0.08)'
    },
    en: {
      headline: 'AI resumes for global hiring teams',
      sub: 'Clear achievements, ATS keyword matching, and concise professional storytelling.',
      tone: 'Direct, measurable, recruiter-friendly',
      gradient: 'linear-gradient(135deg,#2563eb,#7c3aed)',
      bg: 'rgba(37,99,235,0.08)'
    },
    ja: {
      headline: '日本市場向けの履歴書・職務経歴書最適化',
      sub: '誠実さ、構成の美しさ、職務経歴の具体性を重視します。',
      tone: '丁寧、信頼感、端正な表現',
      gradient: 'linear-gradient(135deg,#1b2a4a,#c53030)',
      bg: 'rgba(197,48,48,0.08)'
    },
    ko: {
      headline: '한국 채용 환경에 맞춘 AI 이력서',
      sub: '핵심 역량, 프로젝트 성과, 직무 적합도를 간결하게 강조합니다.',
      tone: '명확함, 세련됨, 역량 중심',
      gradient: 'linear-gradient(135deg,#6366f1,#ec4899)',
      bg: 'rgba(99,102,241,0.08)'
    },
    es: {
      headline: 'CV optimizado para mercados hispanohablantes',
      sub: 'Logros cuantificados, lectura rápida y palabras clave para filtros ATS.',
      tone: 'Cercano, claro, orientado a logros',
      gradient: 'linear-gradient(135deg,#ea580c,#0891b2)',
      bg: 'rgba(234,88,12,0.08)'
    },
    pt: {
      headline: 'Currículos para Brasil e mercados lusófonos',
      sub: 'Experiência objetiva, impacto mensurável e aderência à vaga.',
      tone: 'Prático, confiante, orientado a impacto',
      gradient: 'linear-gradient(135deg,#16a34a,#eab308)',
      bg: 'rgba(22,163,74,0.08)'
    },
    fr: {
      headline: 'CV optimisé pour recruteurs francophones',
      sub: 'Présentation élégante, compétences ciblées et progression professionnelle claire.',
      tone: 'Sobre, élégant, structuré',
      gradient: 'linear-gradient(135deg,#1e3a5f,#b8860b)',
      bg: 'rgba(30,58,95,0.08)'
    },
    de: {
      headline: 'Lebenslauf-Optimierung für den deutschen Markt',
      sub: 'Präzise Struktur, belastbare Ergebnisse und klare Qualifikationsnachweise.',
      tone: 'Präzise, sachlich, strukturiert',
      gradient: 'linear-gradient(135deg,#1f2937,#dc2626)',
      bg: 'rgba(31,41,55,0.08)'
    },
    ar: {
      headline: 'سير ذاتية مخصصة للأسواق العربية',
      sub: 'عرض واضح للخبرات والإنجازات والكلمات المفتاحية المناسبة للوظيفة.',
      tone: 'مهني، واضح، موثوق',
      gradient: 'linear-gradient(135deg,#16a34a,#ca8a04)',
      bg: 'rgba(22,163,74,0.08)'
    },
    hi: {
      headline: 'भारतीय नौकरी बाजार के लिए AI रिज़्यूमे',
      sub: 'कौशल, उपलब्धियों और ATS कीवर्ड को स्पष्ट और प्रभावी तरीके से दिखाएं।',
      tone: 'स्पष्ट, भरोसेमंद, कौशल-केंद्रित',
      gradient: 'linear-gradient(135deg,#ea580c,#7c3aed)',
      bg: 'rgba(234,88,12,0.08)'
    }
  };

  const order = ['zh', 'en', 'ja', 'ko', 'es', 'pt', 'fr', 'de', 'ar', 'hi'];

  onMount(() => {
    const lang = (navigator.language || 'en').toLowerCase();
    const match = order.find(code => lang.startsWith(code));
    if (match) suggestedLang = match;
  });
</script>

<svelte:head>
  <title>ResumeTake - Choose Your Language | AI Resume Optimizer</title>
  <meta name="description" content="Choose your language and market style for ResumeTake AI resume optimization. Supports Chinese, English, Japanese, Korean, Spanish, Portuguese, French, German, Arabic, and Hindi.">
  <link rel="canonical" href="https://resume.takee.top/">
  <meta property="og:title" content="ResumeTake - Choose Your Language | AI Resume Optimizer">
  <meta property="og:description" content="Choose your language and market style for ResumeTake AI resume optimization. Supports Chinese, English, Japanese, Korean, Spanish, Portuguese, French, German, Arabic, and Hindi.">
  <meta property="og:url" content="https://resume.takee.top/">
  <meta property="og:type" content="website">
  <meta property="og:image" content="https://resume.takee.top/og-image.jpg">
  <meta property="og:image:width" content="1200">
  <meta property="og:image:height" content="630">
  <meta property="og:image:alt" content="ResumeTake - AI Resume Builder">
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content="ResumeTake - Choose Your Language | AI Resume Optimizer">
  <meta name="twitter:description" content="Choose your language and market style for ResumeTake AI resume optimization.">
  <meta name="twitter:image" content="https://resume.takee.top/og-image.jpg">
  <link rel="alternate" hreflang="zh" href="https://resume.takee.top/zh">
  <link rel="alternate" hreflang="en" href="https://resume.takee.top/en">
  <link rel="alternate" hreflang="ja" href="https://resume.takee.top/ja">
  <link rel="alternate" hreflang="ko" href="https://resume.takee.top/ko">
  <link rel="alternate" hreflang="es" href="https://resume.takee.top/es">
  <link rel="alternate" hreflang="pt" href="https://resume.takee.top/pt">
  <link rel="alternate" hreflang="fr" href="https://resume.takee.top/fr">
  <link rel="alternate" hreflang="de" href="https://resume.takee.top/de">
  <link rel="alternate" hreflang="ar" href="https://resume.takee.top/ar">
  <link rel="alternate" hreflang="hi" href="https://resume.takee.top/hi">
  <link rel="alternate" hreflang="x-default" href="https://resume.takee.top/en">
</svelte:head>

<main class="root-page" id="main-content" tabindex="-1">
  <section class="hero-shell">
    <div class="mesh mesh-a"></div>
    <div class="mesh mesh-b"></div>
    <div class="hero-copy">
      <div class="brand-mark" aria-hidden="true">R</div>
      <p class="eyebrow">ResumeTake Global</p>
      <h1>Choose the resume experience that fits your market</h1>
      <p class="subtitle">Different countries read resumes differently. Pick a language to get localized tone, structure, keyword strategy, and visual style.</p>
      <div class="quick-actions">
        <a class="btn btn-primary" href="/{suggestedLang}">Use Suggested: {LANGUAGES[suggestedLang].name}</a>
        <a class="btn btn-secondary" href="/en">Continue in English</a>
      </div>
    </div>
  </section>

  <section class="language-section" aria-label="Choose language">
    <div class="section-heading">
      <span>10 Languages</span>
      <h2>Localized for how hiring works in each region</h2>
    </div>

    <div class="language-grid">
      {#each order as code (code)}
        {@const lang = LANGUAGES[code]}
        {@const market = markets[code]}
        <a class="language-card {code === suggestedLang ? 'suggested' : ''}" href="/{code}" dir={lang.dir} style="--card-gradient:{market.gradient};--card-bg:{market.bg}">
          {#if code === suggestedLang}
            <span class="suggested-badge">Suggested</span>
          {/if}
          <div class="card-top">
            <span class="flag" aria-hidden="true">{lang.flag}</span>
            <div>
              <h3>{lang.name}</h3>
              <p>{lang.country}</p>
            </div>
          </div>
          <h4>{market.headline}</h4>
          <p class="market-sub">{market.sub}</p>
          <div class="tone-pill">{market.tone}</div>
        </a>
      {/each}
    </div>
  </section>
</main>

<style>
  .root-page { min-height: 100vh; background: var(--bg); overflow: hidden; }
  .hero-shell {
    position: relative; min-height: 52vh; display: grid; place-items: center;
    padding: 5rem 1.5rem 3rem; background: var(--gradient-hero);
  }
  .mesh { position: absolute; border-radius: 999px; filter: blur(70px); opacity: 0.35; pointer-events: none; }
  .mesh-a { width: 420px; height: 420px; left: -120px; top: -120px; background: var(--primary); }
  .mesh-b { width: 360px; height: 360px; right: -80px; bottom: -120px; background: var(--accent); }
  .hero-copy { position: relative; z-index: 1; max-width: 760px; text-align: center; }
  .brand-mark {
    width: 56px; height: 56px; margin: 0 auto 1rem; border-radius: 18px;
    display: grid; place-items: center; color: white; font-weight: 900; font-size: 1.5rem;
    background: linear-gradient(135deg, var(--primary), var(--accent)); box-shadow: 0 16px 40px var(--primary-glow);
  }
  .eyebrow { color: var(--primary); font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; font-size: 0.78rem; }
  h1 { font-size: clamp(2.35rem, 6vw, 4.75rem); line-height: 0.98; letter-spacing: -0.055em; margin: 0.65rem 0 1rem; }
  .subtitle { color: var(--text-secondary); font-size: clamp(1rem, 2vw, 1.2rem); line-height: 1.75; max-width: 650px; margin: 0 auto; }
  .quick-actions { display: flex; gap: 0.85rem; justify-content: center; flex-wrap: wrap; margin-top: 1.75rem; }
  .language-section { max-width: 1180px; margin: 0 auto; padding: 3rem 1.5rem 5rem; }
  .section-heading { display: flex; justify-content: space-between; align-items: end; gap: 1rem; margin-bottom: 1.5rem; }
  .section-heading span { color: var(--primary); font-weight: 700; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.08em; }
  .section-heading h2 { max-width: 680px; font-size: clamp(1.4rem, 3vw, 2.3rem); line-height: 1.15; text-align: right; }
  .language-grid { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 1rem; }
  .language-card {
    position: relative; min-height: 250px; padding: 1.1rem; border: 1px solid var(--border); border-radius: 1.25rem;
    color: var(--text); background: linear-gradient(180deg, var(--card-bg), var(--bg-glass)); overflow: hidden;
    box-shadow: var(--shadow-sm); transition: transform 0.25s ease, box-shadow 0.25s ease, border-color 0.25s ease;
  }
  .language-card:before { content: ''; position: absolute; inset: 0 0 auto 0; height: 4px; background: var(--card-gradient); }
  .language-card:hover { transform: translateY(-5px); box-shadow: var(--shadow-lg); border-color: color-mix(in srgb, var(--primary) 40%, var(--border)); }
  .language-card.suggested { outline: 2px solid var(--primary); outline-offset: 2px; }
  .suggested-badge {
    position: absolute; top: 0.7rem; right: 0.7rem; padding: 0.25rem 0.55rem; border-radius: 999px;
    font-size: 0.7rem; font-weight: 700; color: white; background: var(--card-gradient);
  }
  .card-top { display: flex; gap: 0.75rem; align-items: center; margin-bottom: 1rem; }
  .flag { font-size: 2rem; }
  .card-top h3 { font-size: 1.05rem; margin: 0; }
  .card-top p { margin: 0.1rem 0 0; color: var(--text-secondary); font-size: 0.8rem; }
  .language-card h4 { font-size: 1rem; line-height: 1.35; margin: 0 0 0.55rem; }
  .market-sub { color: var(--text-secondary); font-size: 0.86rem; line-height: 1.58; }
  .tone-pill {
    position: absolute; left: 1rem; right: 1rem; bottom: 1rem; padding: 0.55rem 0.7rem; border-radius: 0.85rem;
    background: var(--bg-glass); border: 1px solid var(--border); color: var(--text-secondary); font-size: 0.76rem;
  }
  [dir="rtl"] .suggested-badge { right: auto; left: 0.7rem; }
  @media (max-width: 1040px) { .language-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
  @media (max-width: 720px) {
    .hero-shell { min-height: auto; padding-top: 4rem; }
    .section-heading { display: block; }
    .section-heading h2 { text-align: left; margin-top: 0.5rem; }
    .language-grid { grid-template-columns: 1fr; }
    .language-card { min-height: 220px; }
  }
</style>
