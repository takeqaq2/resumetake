<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';
  import { onMount } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));
  let mounted = $state(false);
  onMount(() => { mounted = true; });
</script>

<svelte:head>
  <title>{t.meta.templatesTitle}</title>
  <meta name="description" content={t.meta.templatesDesc}>
  <meta name="keywords" content={t.meta.templatesKeywords}>
  <link rel="canonical" href="https://resume.takee.top/{lang}/templates">
  <meta property="og:title" content={t.meta.templatesTitle}>
  <meta property="og:description" content={t.meta.templatesDesc}>
</svelte:head>

<!-- Header -->
<div class="editor-header" style="padding:2.5rem 0 3rem;position:relative;overflow:hidden">
  <div class="orb orb-blue animate-float" style="width:200px;height:200px;top:-20%;left:10%"></div>
  <div class="orb orb-pink animate-float" style="width:160px;height:160px;bottom:-10%;right:15%;animation-delay:2s"></div>
  <div class="container" style="position:relative">
    <div class="{mounted ? 'animate-fade-in-up' : ''}" style="opacity:0;text-align:center">
      <h1 style="font-size:clamp(1.75rem,3.5vw,2.25rem);font-weight:700;margin-bottom:0.75rem;color:var(--text)">{t.templates.title}</h1>
      <p style="color:var(--text-secondary);max-width:32rem;margin:0 auto">{t.templates.subtitle}</p>
    </div>
  </div>
</div>

<div class="container" style="padding:2.5rem 1.5rem;margin-top:-1rem">
  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:1.5rem">
    {#each t.templates.items as tpl, i}
      {@const gradients = {
        professional: 'linear-gradient(135deg,#3b82f6,#2563eb)',
        modern: 'linear-gradient(135deg,#8b5cf6,#7c3aed)',
        creative: 'linear-gradient(135deg,#ec4899,#db2777)',
        academic: 'linear-gradient(135deg,#10b981,#059669)',
        executive: 'linear-gradient(135deg,#374151,#111827)',
        minimal: 'linear-gradient(135deg,#f59e0b,#d97706)'
      }}
      <a href="/{lang}/editor?template={tpl.id}" class="feature-card {mounted ? 'visible' : ''} reveal" style="padding:0;overflow:hidden;text-decoration:none;transition-delay:{i * 0.08}s;{mounted ? 'opacity:1;transform:none' : ''}">
        <div style="height:10rem;background:{gradients[tpl.id]||gradients.modern};display:flex;align-items:center;justify-content:center;position:relative;overflow:hidden">
          <div style="position:absolute;inset:0;background:radial-gradient(circle at 30% 40%, rgba(255,255,255,0.15) 0%, transparent 60%)"></div>
          <span style="color:rgba(255,255,255,0.25);font-size:3.5rem;font-weight:700;position:relative;z-index:1">Aa</span>
        </div>
        <div style="padding:1.25rem;text-align:left">
          <h3 style="font-weight:600;margin-bottom:0.375rem;color:var(--text)">{tpl.name}</h3>
          <p style="font-size:0.875rem;color:var(--text-secondary);line-height:1.5">{tpl.desc}</p>
        </div>
      </a>
    {/each}
  </div>
</div>
