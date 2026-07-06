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
    return () => window.removeEventListener('scroll', onScroll);
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
    "name":"ResumeTake",
    "url":"https://resume.takee.top/" + lang,
    "description": t.meta.description,
    "applicationCategory":"BusinessApplication",
    "operatingSystem":"Web",
    "inLanguage": lang,
    "offers":{"@type":"Offer","price":"0","priceCurrency":"USD"}
  })}</script>`}
</svelte:head>

<div class="scroll-progress" style="width:{scrollProgress}%"></div>

<!-- ===== HERO ===== -->
<section class="hero">
  <div class="hero-glow hero-glow-1"></div>
  <div class="hero-glow hero-glow-2"></div>
  <div class="hero-glow hero-glow-3"></div>

  <div class="container hero-inner">
    <div class="hero-left">
      <div class="hero-badge anim-1">
        <span class="badge-dot"></span>
        {t.hero.badge}
      </div>

      <h1 class="anim-2">
        {t.hero.title1}
        <span class="gradient-text">{t.hero.title2}</span>
      </h1>

      <p class="hero-desc anim-3">{t.hero.subtitle}</p>

      <div class="hero-tags anim-4">
        {#each ['ATS', 'AI', 'PDF', 'SEO'] as tag}
          <span class="hero-tag">{tag}</span>
        {/each}
      </div>

      <div class="hero-actions anim-5">
        <a href="/{lang}/editor" class="btn btn-primary btn-lg">
          {t.hero.cta}
          <svg width="18" height="18" viewBox="0 0 16 16" fill="none"><path d="M3 8h10M9 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </a>
        <a href="#features" class="btn btn-secondary btn-lg">{t.hero.learnMore}</a>
      </div>

      <div class="hero-note anim-6">
        {#each [lang === 'zh' ? '智能采集信息' : 'Smart parsing', lang === 'zh' ? '关键词优化' : 'Keyword optimization', lang === 'zh' ? '自动生成排版' : 'Auto formatting'] as note}
          <span class="hero-note-item">
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M3 8l3.5 3.5L13 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            {note}
          </span>
        {/each}
      </div>
    </div>

    <div class="hero-right anim-7">
      <div class="preview-card">
        <div class="preview-header">
          <span></span><span></span><span></span>
        </div>
        <div class="preview-body">
          <div class="preview-avatar">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none"><path d="M12 2a5 5 0 110 10 5 5 0 010-10zm0 12c-4.42 0-8 1.79-8 4v2h16v-2c0-2.21-3.58-4-8-4z" fill="currentColor" opacity="0.3"/></svg>
          </div>
          <div class="preview-info">
            <div class="preview-line preview-line-title"></div>
            <div class="preview-line preview-line-sub"></div>
          </div>
        </div>
        <div class="preview-score">
          <div class="preview-score-label">{lang === 'zh' ? '岗位匹配' : 'Match Score'}</div>
          <div class="preview-score-value">98%</div>
        </div>
        <div class="preview-lines">
          <div class="preview-line"></div>
          <div class="preview-line preview-line-short"></div>
          <div class="preview-line"></div>
        </div>
        <div class="preview-tags">
          <span class="preview-tag"></span>
          <span class="preview-tag"></span>
          <span class="preview-tag"></span>
        </div>
      </div>
    </div>
  </div>
</section>

<!-- ===== FEATURES ===== -->
<section id="features" class="section">
  <div class="container">
    <div class="section-header">
      <span class="section-badge">🚀 {t.features.title}</span>
      <h2>{t.features.subtitle}</h2>
    </div>
    <div class="features-grid">
      {#each t.features.items as f, i}
        <div class="feature-card" style="animation-delay:{i * 0.1}s">
          <div class="feature-icon">{f.icon}</div>
          <h3>{f.title}</h3>
          <p>{f.desc}</p>
        </div>
      {/each}
    </div>
  </div>
</section>

<!-- ===== HOW IT WORKS ===== -->
<section class="section section-alt">
  <div class="container">
    <div class="section-header">
      <span class="section-badge">⚡ {lang === 'zh' ? '三步搞定' : '3 Steps'}</span>
      <h2>{lang === 'zh' ? '轻松制作专业简历' : 'Create Your Resume Easily'}</h2>
    </div>
    <div class="steps-grid">
      <div class="step-card">
        <div class="step-num">1</div>
        <h3>{lang === 'zh' ? '粘贴简历内容' : 'Paste Resume'}</h3>
        <p>{lang === 'zh' ? '将你的简历内容粘贴到编辑器中，或上传文件自动识别' : 'Paste your resume content or upload a file for auto-parsing'}</p>
      </div>
      <div class="step-arrow">
        <svg width="24" height="24" viewBox="0 0 16 16" fill="none"><path d="M3 8h10M9 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </div>
      <div class="step-card">
        <div class="step-num">2</div>
        <h3>{lang === 'zh' ? 'AI智能优化' : 'AI Optimize'}</h3>
        <p>{lang === 'zh' ? 'AI自动匹配ATS关键词，优化工作经历，提升简历质量' : 'AI matches ATS keywords and optimizes your experience'}</p>
      </div>
      <div class="step-arrow">
        <svg width="24" height="24" viewBox="0 0 16 16" fill="none"><path d="M3 8h10M9 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </div>
      <div class="step-card">
        <div class="step-num">3</div>
        <h3>{lang === 'zh' ? '下载使用' : 'Download'}</h3>
        <p>{lang === 'zh' ? '一键导出专业简历，提升求职成功率' : 'Export your professional resume and boost your job search'}</p>
      </div>
    </div>
  </div>
</section>

<!-- ===== CTA ===== -->
<section class="section">
  <div class="container">
    <div class="cta-box">
      <h2>{t.cta.title}</h2>
      <p>{t.cta.subtitle}</p>
      <a href="/{lang}/editor" class="btn btn-primary btn-lg">
        {t.cta.button}
        <svg width="18" height="18" viewBox="0 0 16 16" fill="none"><path d="M3 8h10M9 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </a>
    </div>
  </div>
</section>

<style>
  /* ===== Animations ===== */
  .anim-1 { animation: fadeUp 0.6s 0.1s both; }
  .anim-2 { animation: fadeUp 0.6s 0.2s both; }
  .anim-3 { animation: fadeUp 0.6s 0.35s both; }
  .anim-4 { animation: fadeUp 0.6s 0.45s both; }
  .anim-5 { animation: fadeUp 0.6s 0.55s both; }
  .anim-6 { animation: fadeUp 0.6s 0.65s both; }
  .anim-7 { animation: fadeUp 0.6s 0.3s both; }
  @keyframes fadeUp {
    from { opacity: 0; transform: translateY(24px); }
    to { opacity: 1; transform: translateY(0); }
  }

  /* ===== Hero ===== */
  .hero {
    min-height: 90vh; display: flex; align-items: center;
    position: relative; overflow: hidden;
    background: var(--gradient-hero);
    background-size: 300% 300%;
    animation: gradientShift 12s ease-in-out infinite;
  }
  .hero-glow {
    position: absolute; border-radius: 50%; filter: blur(80px);
    opacity: 0.35; pointer-events: none;
  }
  .hero-glow-1 {
    width: 500px; height: 500px; top: -10%; left: -5%;
    background: rgba(37,99,235,0.18);
    animation: orb-float-1 14s ease-in-out infinite;
  }
  .hero-glow-2 {
    width: 400px; height: 400px; top: 50%; right: -8%;
    background: rgba(139,92,246,0.15);
    animation: orb-float-2 16s ease-in-out infinite;
  }
  .hero-glow-3 {
    width: 350px; height: 350px; bottom: -5%; left: 20%;
    background: rgba(236,72,153,0.1);
    animation: orb-float-1 18s ease-in-out infinite reverse;
  }
  .hero-inner {
    position: relative; z-index: 1;
    display: grid; grid-template-columns: 1fr 1fr;
    gap: 4rem; align-items: center;
    padding: 6rem 1.5rem 7rem;
  }
  .hero-left { max-width: 36rem; }
  .hero-badge {
    display: inline-flex; align-items: center; gap: 0.5rem;
    padding: 0.375rem 0.875rem; border-radius: 9999px;
    background: var(--bg-glass); border: 1px solid var(--border);
    font-size: 0.8125rem; font-weight: 500; color: var(--primary);
    backdrop-filter: blur(8px); margin-bottom: 1.25rem;
  }
  .badge-dot {
    width: 6px; height: 6px; border-radius: 50%;
    background: var(--primary); animation: pulse-dot 2s infinite;
  }
  @keyframes pulse-dot {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }
  h1 {
    font-size: clamp(2.5rem, 5vw, 3.75rem);
    font-weight: 800; line-height: 1.1;
    letter-spacing: -0.03em; margin-bottom: 1.25rem;
  }
  .gradient-text {
    background: linear-gradient(135deg, var(--primary), var(--accent));
    -webkit-background-clip: text; -webkit-text-fill-color: transparent;
    background-clip: text;
  }
  .hero-desc {
    font-size: 1.0625rem; color: var(--text-secondary);
    line-height: 1.7; margin-bottom: 1.5rem;
  }
  .hero-tags {
    display: flex; gap: 0.5rem; flex-wrap: wrap; margin-bottom: 2rem;
  }
  .hero-tag {
    padding: 0.25rem 0.625rem; border-radius: 9999px;
    background: var(--bg-glass); border: 1px solid var(--border);
    font-size: 0.75rem; font-weight: 500; color: var(--text-secondary);
    backdrop-filter: blur(8px);
  }
  .hero-actions { display: flex; gap: 0.75rem; flex-wrap: wrap; margin-bottom: 2rem; }
  .btn-lg { padding: 0.875rem 2rem; font-size: 1rem; font-weight: 600; }
  .hero-note {
    display: flex; gap: 1.25rem; flex-wrap: wrap;
  }
  .hero-note-item {
    display: inline-flex; align-items: center; gap: 0.375rem;
    font-size: 0.8125rem; color: var(--text-secondary);
  }
  .hero-note-item svg { color: #10b981; }

  /* ===== Preview Card ===== */
  .preview-card {
    background: var(--bg-glass); border: 1px solid var(--border);
    border-radius: var(--radius-lg); padding: 1.25rem;
    backdrop-filter: blur(16px); box-shadow: var(--shadow-lg);
    animation: float 6s ease-in-out infinite;
  }
  .preview-header {
    display: flex; gap: 0.375rem; margin-bottom: 1rem;
  }
  .preview-header span {
    width: 8px; height: 8px; border-radius: 50%;
    background: var(--border);
  }
  .preview-header span:nth-child(1) { background: #ef4444; }
  .preview-header span:nth-child(2) { background: #f59e0b; }
  .preview-header span:nth-child(3) { background: #10b981; }
  .preview-body {
    display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem;
  }
  .preview-avatar {
    width: 40px; height: 40px; border-radius: 50%;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    display: flex; align-items: center; justify-content: center;
    color: white; flex-shrink: 0;
  }
  .preview-line {
    height: 8px; border-radius: 4px; background: var(--border);
    margin-bottom: 0.375rem;
  }
  .preview-line-title { width: 80%; background: var(--text); opacity: 0.15; }
  .preview-line-sub { width: 50%; }
  .preview-score {
    display: flex; justify-content: space-between; align-items: center;
    padding: 0.625rem 0.75rem; border-radius: var(--radius);
    background: rgba(16,185,129,0.08); border: 1px solid rgba(16,185,129,0.15);
    margin-bottom: 0.75rem;
  }
  .preview-score-label { font-size: 0.75rem; color: var(--text-secondary); }
  .preview-score-value {
    font-size: 1.25rem; font-weight: 800; color: #059669;
  }
  .preview-lines { margin-bottom: 0.75rem; }
  .preview-line-short { width: 60%; }
  .preview-tags { display: flex; gap: 0.375rem; }
  .preview-tag {
    width: 48px; height: 20px; border-radius: 10px;
    background: var(--border); opacity: 0.5;
  }

  /* ===== Sections ===== */
  .section { padding: 5rem 0; }
  .section-alt { background: var(--bg-surface); }
  .section-header {
    text-align: center; margin-bottom: 3rem;
  }
  .section-badge {
    display: inline-flex; align-items: center; gap: 0.375rem;
    padding: 0.375rem 0.875rem; border-radius: 9999px;
    background: var(--bg-glass); border: 1px solid var(--border);
    font-size: 0.8125rem; font-weight: 500; color: var(--primary);
    margin-bottom: 0.75rem; backdrop-filter: blur(8px);
  }
  .section-header h2 {
    font-size: clamp(1.5rem, 3vw, 2rem);
    font-weight: 700; color: var(--text);
  }

  /* ===== Features ===== */
  .features-grid {
    display: grid; grid-template-columns: repeat(3, 1fr);
    gap: 1.5rem;
  }
  .feature-card {
    background: var(--bg-glass); border: 1px solid var(--border);
    border-radius: var(--radius-lg); padding: 2rem;
    text-align: center; backdrop-filter: blur(12px);
    transition: all 0.35s cubic-bezier(0.4,0,0.2,1);
  }
  .feature-card:hover {
    transform: translateY(-4px); box-shadow: var(--shadow-lg);
    border-color: rgba(37,99,235,0.15);
  }
  .feature-icon {
    font-size: 2.5rem; margin-bottom: 1rem;
    display: inline-flex; align-items: center; justify-content: center;
    width: 64px; height: 64px; border-radius: var(--radius-lg);
    background: var(--bg-surface);
  }
  .feature-card h3 {
    font-weight: 600; margin-bottom: 0.5rem;
    color: var(--text); font-size: 1.0625rem;
  }
  .feature-card p {
    color: var(--text-secondary); font-size: 0.9375rem; line-height: 1.65;
  }

  /* ===== Steps ===== */
  .steps-grid {
    display: flex; align-items: flex-start; justify-content: center;
    gap: 1rem; flex-wrap: wrap;
  }
  .step-card {
    flex: 1; min-width: 200px; max-width: 280px;
    text-align: center; padding: 1.5rem;
  }
  .step-num {
    width: 48px; height: 48px; border-radius: 50%;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    color: white; font-weight: 700; font-size: 1.25rem;
    display: inline-flex; align-items: center; justify-content: center;
    margin-bottom: 1rem; box-shadow: 0 4px 16px var(--primary-glow);
  }
  .step-card h3 {
    font-weight: 600; margin-bottom: 0.5rem;
    color: var(--text); font-size: 1.0625rem;
  }
  .step-card p {
    color: var(--text-secondary); font-size: 0.875rem; line-height: 1.6;
  }
  .step-arrow {
    display: flex; align-items: center; justify-content: center;
    padding-top: 1rem; color: var(--text-secondary); opacity: 0.4;
  }

  /* ===== CTA ===== */
  .cta-box {
    text-align: center; padding: 4rem 2rem;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    border-radius: var(--radius-lg); color: white;
    box-shadow: 0 8px 32px var(--primary-glow);
  }
  .cta-box h2 {
    font-size: clamp(1.5rem, 3vw, 2rem);
    font-weight: 700; margin-bottom: 0.75rem; color: white;
  }
  .cta-box p {
    opacity: 0.9; margin-bottom: 2rem; max-width: 32rem;
    margin-left: auto; margin-right: auto;
  }
  .cta-box .btn-primary {
    background: white; color: var(--primary);
    box-shadow: 0 4px 16px rgba(0,0,0,0.15);
  }
  .cta-box .btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 24px rgba(0,0,0,0.2);
  }

  /* ===== Responsive ===== */
  @media (max-width: 768px) {
    .hero-inner { grid-template-columns: 1fr; text-align: center; gap: 2rem; }
    .hero-left { max-width: none; }
    .hero-actions { justify-content: center; }
    .hero-note { justify-content: center; }
    .hero-right { display: none; }
    .features-grid { grid-template-columns: 1fr; }
    .steps-grid { flex-direction: column; align-items: center; }
    .step-arrow { transform: rotate(90deg); padding: 0; }
  }

  @keyframes gradientShift {
    0% { background-position: 0% 50%; }
    50% { background-position: 100% 50%; }
    100% { background-position: 0% 50%; }
  }
  @keyframes orb-float-1 {
    0%, 100% { transform: translate(0, 0); }
    33% { transform: translate(30px, -40px); }
    66% { transform: translate(-20px, 20px); }
  }
  @keyframes orb-float-2 {
    0%, 100% { transform: translate(0, 0); }
    33% { transform: translate(-40px, 20px); }
    66% { transform: translate(30px, -30px); }
  }
  @keyframes float {
    0%, 100% { transform: translateY(0px); }
    50% { transform: translateY(-12px); }
  }
</style>
