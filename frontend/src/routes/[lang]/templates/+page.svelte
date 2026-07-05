<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));
</script>

<svelte:head>
  <title>{t.meta.templatesTitle}</title>
  <meta name="description" content={t.meta.templatesDesc}>
  <meta name="keywords" content={t.meta.templatesKeywords}>
  <link rel="canonical" href="https://resume.takee.top/{lang}/templates">
  <meta property="og:title" content={t.meta.templatesTitle}>
  <meta property="og:description" content={t.meta.templatesDesc}>
</svelte:head>

<div class="container" style="padding:3rem 1.5rem">
  <div style="text-align:center;margin-bottom:3rem">
    <h1 style="font-size:2.25rem;font-weight:700;margin-bottom:0.75rem;color:var(--text)">{t.templates.title}</h1>
    <p style="color:var(--text-secondary)">{t.templates.subtitle}</p>
  </div>
  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:1.5rem">
    {#each t.templates.items as tpl}
      {@const gradients = {
        professional: 'linear-gradient(135deg,#3b82f6,#2563eb)',
        modern: 'linear-gradient(135deg,#8b5cf6,#7c3aed)',
        creative: 'linear-gradient(135deg,#ec4899,#db2777)',
        academic: 'linear-gradient(135deg,#10b981,#059669)',
        executive: 'linear-gradient(135deg,#374151,#111827)',
        minimal: 'linear-gradient(135deg,#f59e0b,#d97706)'
      }}
      <a href="/{lang}/editor?template={tpl.id}" class="card" style="padding:0;overflow:hidden;transition:transform 0.2s,box-shadow 0.2s;text-decoration:none">
        <div style="height:10rem;background:{gradients[tpl.id]||gradients.modern};display:flex;align-items:center;justify-content:center">
          <span style="color:rgba(255,255,255,0.2);font-size:3.5rem;font-weight:700">Aa</span>
        </div>
        <div style="padding:1.25rem">
          <h3 style="font-weight:600;margin-bottom:0.25rem;color:var(--text)">{tpl.name}</h3>
          <p style="font-size:0.875rem;color:var(--text-secondary)">{tpl.desc}</p>
        </div>
      </a>
    {/each}
  </div>
</div>
