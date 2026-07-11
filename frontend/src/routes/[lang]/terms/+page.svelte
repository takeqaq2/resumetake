<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';

  let lang = $derived($page.params.lang || 'en');
  let t = $derived(getTranslation(lang));
</script>

<svelte:head>
  <title>{t.terms.title}</title>
  <meta name="description" content={t.terms.description}>
  <link rel="canonical" href="https://resume.takee.top/{lang}/terms">
  <meta property="og:title" content={t.terms.title}>
  <meta property="og:description" content={t.terms.description}>
  <meta property="og:url" content="https://resume.takee.top/{lang}/terms">
  <meta property="og:type" content="website">
</svelte:head>

<section class="legal-page">
  <div class="container legal-container">
    <p class="eyebrow">ResumeTake</p>
    <h1>{t.terms.heading}</h1>
    <p class="updated"><time datetime="2026-07-09">{t.terms.updated}</time></p>

    <div class="legal-card">
      <h2 id="s1">{t.terms.s1Title}</h2>
      <p>{t.terms.s1Body}</p>

      <h2 id="s2">{t.terms.s2Title}</h2>
      <p>{t.terms.s2Body}</p>

      <h2 id="s3">{t.terms.s3Title}</h2>
      <p>{t.terms.s3Body}</p>

      <h2 id="s4">{t.terms.s4Title}</h2>
      <p>{t.terms.s4Body}</p>

      <h2 id="s5">{t.terms.s5Title}</h2>
      <p>{t.terms.s5Body}</p>

      <h2 id="s6">{t.terms.s6Title}</h2>
      <p>{t.terms.s6Body}</p>

      <h2 id="s7">{t.terms.s7Title}</h2>
      <p>{t.terms.contactPrefix} <a href="mailto:takeqaq@gmail.com">takeqaq@gmail.com</a>{t.terms.contactSuffix}</p>
    </div>
  </div>
</section>

<style>
  .legal-page { padding: 4rem 1.5rem; }
  .legal-container { max-width: 880px; }
  .eyebrow { color: var(--primary); font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; font-size: 0.75rem; }
  h1 { font-size: clamp(2rem, 4vw, 3rem); margin: 0.5rem 0; }
  .updated { color: var(--text-secondary); margin-bottom: 2rem; }
  .legal-card { background: var(--bg-surface); border: 1px solid var(--border); border-radius: var(--radius); padding: 2rem; line-height: 1.75; }
  h2 { font-size: 1.125rem; margin: 1.75rem 0 0.5rem; }
  h2:first-child { margin-top: 0; }
  p { color: var(--text-secondary); }
  a { color: var(--primary); }
</style>
