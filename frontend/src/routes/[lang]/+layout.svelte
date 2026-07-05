<script>
  import '../../lib/styles/themes.css';
  import { page } from '$app/stores';
  import { getTranslation, LANGUAGES } from '$lib/i18n/index.js';
  import { afterNavigate } from '$app/navigation';

  let { children } = $props();
  let showLangMenu = $state(false);
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

<div data-lang={lang} dir={langInfo.dir}>
  <header style="position:sticky;top:0;z-index:50;backdrop-filter:blur(16px);-webkit-backdrop-filter:blur(16px);background:color-mix(in srgb,var(--bg) 85%,transparent);border-bottom:1px solid var(--border);transition:background 0.2s">
    <nav class="container" style="display:flex;align-items:center;justify-content:space-between;height:4rem">
      <a href="/{lang}" style="display:flex;align-items:center;gap:0.625rem;text-decoration:none">
        <div style="width:2.25rem;height:2.25rem;background:linear-gradient(135deg,var(--primary),var(--accent));border-radius:var(--radius);display:flex;align-items:center;justify-content:center;box-shadow:0 2px 8px rgba(59,130,246,0.3)">
          <span style="color:white;font-weight:700;font-size:1.125rem">R</span>
        </div>
        <span style="font-weight:600;font-size:1.125rem;color:var(--text)">ResumeTake</span>
      </a>
      <nav style="display:flex;align-items:center;gap:0.5rem">
        <a href="/{lang}" class="btn btn-secondary" style="padding:0.5rem 1rem">{t.nav.home}</a>
        <a href="/{lang}/editor" class="btn btn-primary" style="padding:0.5rem 1rem">{t.nav.start}</a>
        <div style="position:relative">
          <button class="btn btn-secondary" style="padding:0.5rem 0.75rem;font-size:0.8125rem" onclick={() => showLangMenu = !showLangMenu}>
            {langInfo.flag} {langInfo.name}
          </button>
          {#if showLangMenu}
            <div style="position:absolute;right:0;top:100%;margin-top:0.25rem;background:var(--bg);border:1px solid var(--border);border-radius:var(--radius);padding:0.375rem;min-width:10rem;box-shadow:0 8px 24px rgba(0,0,0,0.12);z-index:100">
              {#each Object.entries(LANGUAGES) as [code, info]}
                <button
                  style="display:flex;align-items:center;gap:0.5rem;width:100%;padding:0.5rem 0.75rem;border:none;background:none;color:{code===lang?'var(--primary)':'var(--text)'};font-size:0.8125rem;cursor:pointer;border-radius:0.25rem;text-align:left;font-weight:{code===lang?'600':'400'}"
                  onclick={() => switchLanguage(code)}
                  onmouseenter={(e) => e.target.style.background='var(--bg-surface)'}
                  onmouseleave={(e) => e.target.style.background='none'}
                >
                  <span>{info.flag}</span>
                  <span>{info.name}</span>
                </button>
              {/each}
            </div>
          {/if}
        </div>
      </nav>
    </nav>
  </header>

  <main style="flex:1;min-height:calc(100vh - 4rem)">
    {@render children()}
  </main>

  <footer style="background:var(--bg-surface);border-top:1px solid var(--border);padding:2.5rem 0;margin-top:4rem">
    <div class="container" style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:1rem">
      <div>
        <span style="color:var(--text-secondary);font-size:0.875rem">&copy; {new Date().getFullYear()} ResumeTake. {t.footer.copyright}</span>
      </div>
      <div style="display:flex;gap:1.5rem;color:var(--text-secondary);font-size:0.875rem">
        <a href="/{lang}/editor" style="color:var(--text-secondary);transition:color 0.2s">{t.footer.createResume}</a>
        <a href="/{lang}/templates" style="color:var(--text-secondary);transition:color 0.2s">{t.footer.templates}</a>
      </div>
    </div>
  </footer>
</div>

<svelte:window onclick={(e) => { if (showLangMenu && !e.target.closest('[data-lang-menu]')) showLangMenu = false; }} />
