<script>
  import '../../lib/styles/themes.css';
  import { page } from '$app/stores';
  import { getTranslation, LANGUAGES } from '$lib/i18n/index.js';
  import { afterNavigate } from '$app/navigation';

  let { children } = $props();
  let showLangMenu = $state(false);
  let scrolled = $state(false);
  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));
  let langInfo = $derived(LANGUAGES[lang] || LANGUAGES.en);

  afterNavigate(() => { showLangMenu = false; });

  function switchLanguage(newLang) {
    const path = $page.url.pathname;
    const segments = path.split('/');
    segments[1] = newLang;
    window.location.href = segments.join('/');
  }

  function onScroll() { scrolled = window.scrollY > 20; }
</script>

<svelte:head>
  <html lang={lang} dir={langInfo.dir}></html>
  <meta name="theme-color" content={lang === 'zh' ? '#dc2626' : lang === 'ja' ? '#1a1a2e' : lang === 'ko' ? '#6366f1' : '#2563eb'}>
  <link rel="icon" type="image/svg+xml" href="/favicon.svg">
  <meta property="og:locale" content={langInfo.locale}>
  {#each Object.entries(LANGUAGES) as [code, info]}
    <link rel="alternate" hreflang={code} href="https://resume.takee.top/{code}{$page.url.pathname.replace(/^\/[^\/]+/, '')}">
  {/each}
  <link rel="alternate" hreflang="x-default" href="https://resume.takee.top/en{$page.url.pathname.replace(/^\/[^\/]+/, '')}">
</svelte:head>

<svelte:window onscroll={onScroll} onclick={(e) => { if (showLangMenu && !e.target.closest('[data-lang-menu]')) showLangMenu = false; }} />

<div data-lang={lang} dir={langInfo.dir} style="min-height:100vh;display:flex;flex-direction:column">
  <header class="glass-header" style="{scrolled ? 'box-shadow:0 1px 12px rgba(0,0,0,0.06)' : ''}">
    <nav class="container" style="display:flex;align-items:center;justify-content:space-between;height:4rem">
      <a href="/{lang}" style="display:flex;align-items:center;gap:0.625rem;text-decoration:none">
        <div class="logo-icon">
          <span style="color:white;font-weight:700;font-size:1.125rem;position:relative;z-index:1">R</span>
        </div>
        <span style="font-weight:700;font-size:1.125rem;color:var(--text);letter-spacing:-0.02em">ResumeTake</span>
      </a>
      <nav style="display:flex;align-items:center;gap:0.5rem">
        <a href="/{lang}" class="btn btn-secondary" style="padding:0.5rem 1rem">{t.nav.home}</a>
        <a href="/{lang}/editor" class="btn btn-primary" style="padding:0.5rem 1.25rem;font-weight:600">
          <span>{t.nav.start}</span>
          <svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M3 8h10M9 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </a>
        <div data-lang-menu style="position:relative">
          <button class="btn btn-secondary" style="padding:0.5rem 0.75rem;font-size:0.8125rem;display:flex;align-items:center;gap:0.375rem" onclick={() => showLangMenu = !showLangMenu} aria-expanded={showLangMenu} aria-haspopup="true" aria-label={t.nav.language}>
            <span>{langInfo.flag}</span>
            <span>{langInfo.name}</span>
            <svg width="10" height="10" viewBox="0 0 10 10" fill="none" style="transition:transform 0.2s;{showLangMenu ? 'transform:rotate(180deg)' : ''}" aria-hidden="true"><path d="M2 4l3 3 3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
          </button>
          {#if showLangMenu}
            <div class="lang-menu" role="menu">
              {#each Object.entries(LANGUAGES) as [code, info]}
                <button class="{code===lang ? 'active' : ''}" onclick={() => switchLanguage(code)} role="menuitem">
                  <span style="font-size:1rem">{info.flag}</span>
                  <span>{info.name}</span>
                </button>
              {/each}
            </div>
          {/if}
        </div>
      </nav>
    </nav>
  </header>

  <main id="main-content" style="flex:1;min-height:calc(100vh - 4rem)">
    {@render children()}
  </main>

  <footer style="background:var(--bg-surface);border-top:1px solid var(--border);padding:2.5rem 0;margin-top:4rem">
    <div class="container" style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:1rem">
      <div style="display:flex;align-items:center;gap:0.625rem">
        <div class="logo-icon" style="width:1.75rem;height:1.75rem">
          <span style="color:white;font-weight:700;font-size:0.875rem;position:relative;z-index:1">R</span>
        </div>
        <span style="color:var(--text-secondary);font-size:0.875rem">&copy; {new Date().getFullYear()} ResumeTake. {t.footer.copyright}</span>
      </div>
      <div style="display:flex;gap:1.5rem;color:var(--text-secondary);font-size:0.875rem">
        <a href="/{lang}/editor" style="color:var(--text-secondary);transition:color 0.2s;position:relative" class="footer-link">{t.footer.createResume}</a>
        <a href="/{lang}/templates" style="color:var(--text-secondary);transition:color 0.2s;position:relative" class="footer-link">{t.footer.templates}</a>
      </div>
    </div>
  </footer>
</div>

<style>
  :global(.footer-link::after) {
    content: '';
    position: absolute;
    left: 0;
    bottom: -2px;
    width: 100%;
    height: 1.5px;
    background: var(--primary);
    transform: scaleX(0);
    transform-origin: left;
    transition: transform 0.25s ease;
  }
  :global(.footer-link:hover::after) {
    transform: scaleX(1);
  }
  :global(.footer-link:hover) {
    color: var(--primary) !important;
  }
</style>
