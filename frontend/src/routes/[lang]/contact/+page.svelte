<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';

  let lang = $derived($page.params.lang || 'en');
  let t = $derived(getTranslation(lang));
</script>

<svelte:head>
  <title>{t.contact.title}</title>
  <meta name="description" content={t.contact.description}>
  <link rel="canonical" href="https://resume.takee.top/{lang}/contact">
  <meta property="og:title" content={t.contact.title}>
  <meta property="og:description" content={t.contact.description}>
  <meta property="og:url" content="https://resume.takee.top/{lang}/contact">
  <meta property="og:type" content="website">
</svelte:head>

<section class="contact-page">
  <div class="container contact-container">
    <p class="eyebrow">ResumeTake</p>
    <h1>{t.contact.heading}</h1>
    <p class="lead">{t.contact.lead}</p>

    <div class="contact-card">
      <h2>{t.contact.emailLabel}</h2>
      <a href="mailto:takeqaq@gmail.com">takeqaq@gmail.com</a>
      <p>{t.contact.emailNote}</p>
    </div>
  </div>
</section>

<style>
  .contact-page { padding: 4rem 1.5rem; }
  .contact-container { max-width: 760px; }
  .eyebrow { color: var(--primary); font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; font-size: 0.75rem; }
  h1 { font-size: clamp(2rem, 4vw, 3rem); margin: 0.5rem 0; }
  .lead { color: var(--text-secondary); line-height: 1.7; margin-bottom: 2rem; }
  .contact-card { background: var(--bg-surface); border: 1px solid var(--border); border-radius: var(--radius); padding: 2rem; }
  h2 { font-size: 1.125rem; margin-bottom: 0.5rem; }
  a { color: var(--primary); font-weight: 700; }
  p { color: var(--text-secondary); margin-top: 1rem; }
</style>
