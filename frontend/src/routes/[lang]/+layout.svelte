<script>
  import { page } from '$app/stores';
  import { getTranslation, LANGUAGES } from '$lib/i18n/index.js';
  import AdSenseScript from '$lib/AdSenseScript.svelte';
  import CookieNotice from '$lib/CookieNotice.svelte';
  import { afterNavigate, goto } from '$app/navigation';
  import { onMount, tick } from 'svelte';

  let { children } = $props();
  let showLangMenu = $state(false);
  let showMobileMenu = $state(false);
  let scrolled = $state(false);
  let isLoggedIn = $state(false);
  let langMenuRef = $state();
  let langTriggerBtn = $state();
  let mainEl = $state();
  // UI1: manual theme override. 'auto' follows OS preference (no data-theme attr).
  let theme = $state('auto');
  let themeReady = $state(false);
  // R55-F5: avoid SSR hydration mismatch — new Date().getFullYear() may
  // differ between server (UTC) and client (local timezone) around midnight
  // on New Year's Eve. Initialize in onMount for client-side consistency.
  let footerYear = $state(0);
  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));
  let langInfo = $derived(LANGUAGES[lang] || LANGUAGES.en);

  afterNavigate(async () => {
    showLangMenu = false;
    // Refresh login state on every navigation so the header reflects
    // login/logout performed on other routes (e.g. /auth) without a full
    // page reload. onMount only runs once and would otherwise leave the
    // header showing a stale "Login"/"Logout" button.
    try { isLoggedIn = !!localStorage.getItem('token'); } catch (e) {}
    // Close the mobile menu if open and wait for the DOM to flush so the
    // inert attribute on <main> is removed before we focus it (otherwise
    // focus silently fails and the mobile-menu close effect focuses the
    // hamburger button instead).
    if (showMobileMenu) {
      showMobileMenu = false;
      await tick();
    }
    // R38b-M1: skip focus management when navigating via language switch —
    // switchLanguage restores focus to the lang menu trigger button via rAF,
    // and afterNavigate firing after that rAF would override it, stranding
    // keyboard users' focus on #main-content instead of back on the menu.
    if (!skipNextFocusManagement && mainEl && !showMobileMenu) mainEl.focus({ preventScroll: true });
    skipNextFocusManagement = false;
  });

  // R38b-M1: set by switchLanguage so the next afterNavigate skips focus
  // management (see comment above).
  let skipNextFocusManagement = false;

  function switchLanguage(newLang) {
    const wasLangMenuOpen = showLangMenu;
    showLangMenu = false;
    const path = $page.url.pathname;
    const segments = path.split('/');
    segments[1] = newLang;
    // Preserve query string and hash so URL-dependent state (e.g.
    // ?payment=success on pricing) survives language switching.
    // R55-F6: update language prefix in ?redirect= param so post-login
    // redirect matches the newly selected language.
    let search = $page.url.search;
    if (search.includes('redirect=')) {
      search = search.replace(/(redirect=)(%2F)[a-z]{2}/, `$1$2${newLang}`);
    }
    skipNextFocusManagement = true;
    goto(segments.join('/') + search + $page.url.hash);
    // WAI-ARIA Menu Button: after selecting an item, return focus to the
    // trigger button. Only when the desktop lang menu was open (mobile menu
    // language buttons are handled by the mobile-menu close effect).
    if (wasLangMenuOpen) {
      requestAnimationFrame(() => { if (langTriggerBtn) langTriggerBtn.focus(); });
    }
  }

  onMount(() => {
    try {
      const saved = localStorage.getItem('theme');
      if (saved === 'light' || saved === 'dark') theme = saved;
    } catch (e) {}
    themeReady = true;
    footerYear = new Date().getFullYear();
    // R45-F5: rAF throttle — scroll fires at 60-120Hz on mobile; without
    // throttling, each event reads window.scrollY and writes to $state.
    // Matches the pattern used by routes/[lang]/+page.svelte (ticking flag).
    let scrollTicking = false;
    let scrollRafId = null;
    function onScroll() {
      if (scrollTicking) return;
      scrollTicking = true;
      scrollRafId = requestAnimationFrame(() => {
        scrolled = window.scrollY > 20;
        scrollTicking = false;
        scrollRafId = null;
      });
    }
    function onAuthChange() { try { isLoggedIn = !!localStorage.getItem('token'); } catch (e) {} }
    window.addEventListener('scroll', onScroll, { passive: true });
    // lib/api.js dispatches this when a 401 is received and the token is
    // cleared — keeps the header's Login/Logout button in sync without waiting
    // for the next navigation.
    window.addEventListener('auth:unauthorized', onAuthChange);
    window.addEventListener('storage', onAuthChange);
    // Close mobile menu when resizing to desktop
    const desktopMQ = window.matchMedia('(min-width: 769px)');
    const onDesktopChange = (e) => { if (e.matches) showMobileMenu = false; };
    desktopMQ.addEventListener('change', onDesktopChange);
    try { isLoggedIn = !!localStorage.getItem('token'); } catch (e) {}
    onScroll();
    return () => {
      // R55b-F5: cancel pending rAF — without this, a scroll event fired
      // just before unmount would schedule a callback that writes to
      // `scrolled` $state on an unmounted component.
      if (scrollRafId !== null) cancelAnimationFrame(scrollRafId);
      window.removeEventListener('scroll', onScroll);
      window.removeEventListener('auth:unauthorized', onAuthChange);
      window.removeEventListener('storage', onAuthChange);
      desktopMQ.removeEventListener('change', onDesktopChange);
    };
  });

  $effect(() => {
    if (typeof document !== 'undefined') {
      document.documentElement.lang = lang || 'en';
      document.documentElement.dir = langInfo.dir || 'ltr';
      // Set data-lang on <html> so themes.css rules like
      // [data-lang="zh"] body { font-family: ... } match (data-lang was only
      // set on a <div> inside <body>, never on an ancestor of <body>).
      document.documentElement.dataset.lang = lang || 'en';
      // Skip link is server-rendered in app.html and doesn't update on
      // client-side language navigation — sync its text here.
      const skipLink = document.querySelector('.skip-link');
      if (skipLink) skipLink.textContent = t.nav.skipToContent || 'Skip to content';
    }
  });

  // UI1: apply manual theme override to <html> and persist. The first run is
  // skipped until onMount above restores the saved choice, so we don't
  // clobber the pre-paint theme set by app.html's inline script (FOUC guard).
  $effect(() => {
    if (typeof document === 'undefined' || !themeReady) return;
    if (theme === 'auto') {
      delete document.documentElement.dataset.theme;
      try { localStorage.removeItem('theme'); } catch (e) {}
    } else {
      document.documentElement.dataset.theme = theme;
      try { localStorage.setItem('theme', theme); } catch (e) {}
      // Update theme-color meta to match the active theme (app.html only
      // declares media-based metas, which don't reflect a manual override).
      document.querySelectorAll('meta[name="theme-color"]').forEach(m => {
        m.setAttribute('content', theme === 'dark' ? '#0b0f1a' : '#2563eb');
      });
    }
  });
  // One-time restore of the saved preference now happens in onMount above.

  // Lock body scroll when mobile menu is open to prevent background scroll
  // on touch devices (the overlay is position:fixed but body still scrolls).
  $effect(() => {
    if (typeof document === 'undefined') return;
    document.body.style.overflow = showMobileMenu ? 'hidden' : '';
    return () => { document.body.style.overflow = ''; };
  });

  // Cycle auto → dark → light → auto to match the 3-state toggle button.
  // R42b-M1: close mobile menu after theme switch — in the mobile menu,
  // toggling theme changes the icon but leaves the menu open, requiring an
  // extra tap to close. Desktop header doesn't have this issue (no menu).
  function toggleTheme() {
    theme = theme === 'auto' ? 'dark' : theme === 'dark' ? 'light' : 'auto';
    showMobileMenu = false;
  }

  function logout() {
    // Invalidate token server-side (fire-and-forget). Even if the request
    // fails, we clear localStorage and redirect — the token is useless once
    // the user intends to log out.
    try {
      const token = localStorage.getItem('token');
      if (token) {
        // R32-L4: add a short timeout so a hanging backend doesn't leave
        // an idle connection open indefinitely.
        const ctrl = new AbortController();
        const timer = setTimeout(() => ctrl.abort(), 5000);
        fetch('/api/v1/auth/logout', {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${token}` },
          signal: ctrl.signal
        }).catch(() => {}).finally(() => clearTimeout(timer));
      }
    } catch (e) {}
    try { localStorage.removeItem('token'); } catch (e) {}
    // Clear user-specific cached data so the next login (potentially a
    // different user on a shared device) doesn't see the previous user's
    // generated resume or template selection.
    try { localStorage.removeItem('generated_resume'); } catch (e) {}
    // R43-F4: Removed localStorage.removeItem('selected_template') — no code
    // in the codebase ever setItem('selected_template'), making it dead code.
    isLoggedIn = false;
    goto(`/${lang}`);
  }

  function toggleLangMenu(e) {
    e.stopPropagation();
    showLangMenu = !showLangMenu;
    // WAI-ARIA Menu Button: when menu opens, move focus to the first item.
    // langMenuRef is bound inside a {#if showLangMenu} block, so it's only
    // populated after Svelte flushes the DOM update. Reading it synchronously
    // here would always be undefined on first open. Defer the ref access to
    // the rAF callback (runs after DOM flush).
    if (showLangMenu) {
      requestAnimationFrame(() => {
        if (!langMenuRef) return;
        const firstItem = langMenuRef.querySelector('[role="menuitem"]');
        if (firstItem) firstItem.focus();
      });
    }
  }

  function onDocClick(e) {
    if (showLangMenu && !e.target.closest('[data-lang-menu]')) {
      showLangMenu = false;
      if (langTriggerBtn) langTriggerBtn.focus();
    }
  }

  // WAI-ARIA Menu keyboard navigation: Arrow keys move between items,
  // Home/End jump to first/last, Escape closes and returns focus to trigger.
  function onLangMenuKeydown(e) {
    const items = langMenuRef ? Array.from(langMenuRef.querySelectorAll('[role="menuitem"]')) : [];
    if (items.length === 0) return;
    const currentIndex = items.indexOf(document.activeElement);
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        items[(currentIndex + 1) % items.length].focus();
        break;
      case 'ArrowUp':
        e.preventDefault();
        if (currentIndex === -1) {
          items[items.length - 1].focus();
        } else {
          items[(currentIndex - 1 + items.length) % items.length].focus();
        }
        break;
      case 'Home':
        e.preventDefault();
        items[0].focus();
        break;
      case 'End':
        e.preventDefault();
        items[items.length - 1].focus();
        break;
      case 'Escape': {
        e.preventDefault();
        showLangMenu = false;
        const trigger = e.currentTarget.closest('[data-lang-menu]').querySelector('button[aria-haspopup="menu"]');
        if (trigger) trigger.focus();
        break;
      }
      case 'Tab':
        // WAI-ARIA Menu Button: Tab closes the menu and moves focus to the
        // next tabbable element on the page (don't preventDefault).
        showLangMenu = false;
        break;
    }
  }

  // Mobile menu focus trap: when the menu is open, Tab cycles within the
  // menu items instead of escaping to the page behind the overlay.
  let hamburgerBtn = $state();
  let mobileMenuPanel = $state();

  function trapMobileMenuFocus(e) {
    if (e.key !== 'Tab' || !mobileMenuPanel) return;
    const focusable = mobileMenuPanel.querySelectorAll('a, button, [tabindex]:not([tabindex="-1"])');
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault();
      first.focus();
    }
  }

  // Track previous state so we can detect "just closed" — when the menu
  // closes, the {#if showMobileMenu} block is removed from the DOM and
  // mobileMenuPanel becomes undefined before this effect reruns, so we
  // can't check it in the close branch.
  let prevMobileMenuOpen = false;
  $effect(() => {
    const isOpen = showMobileMenu;
    if (isOpen && mobileMenuPanel) {
      const focusable = mobileMenuPanel.querySelectorAll('a, button');
      if (focusable.length > 0) focusable[0].focus();
    } else if (!isOpen && prevMobileMenuOpen) {
      // Menu just closed — return focus to the hamburger button (WAI-ARIA
      // Dialog pattern). Don't check mobileMenuPanel (already gone).
      if (hamburgerBtn) hamburgerBtn.focus();
    }
    prevMobileMenuOpen = isOpen;
  });
</script>

<AdSenseScript />

<svelte:head>
  <meta property="og:locale" content={langInfo.locale}>
  {#if !$page.url.pathname.includes('/auth')}
  {#each Object.entries(LANGUAGES) as [code, info] (code)}
    <link rel="alternate" hreflang={code} href="https://resume.takee.top/{code}{$page.url.pathname.replace(/^\/[^\/]+/, '').replace(/\/+$/, '')}">
  {/each}
  <link rel="alternate" hreflang="x-default" href="https://resume.takee.top/en{$page.url.pathname.replace(/^\/[^\/]+/, '').replace(/\/+$/, '')}">
  {/if}
  {#if lang === 'zh'}
    <link href="https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;500;700&display=swap" rel="stylesheet">
  {:else if lang === 'ja'}
    <link href="https://fonts.googleapis.com/css2?family=Noto+Sans+JP:wght@400;500;700&display=swap" rel="stylesheet">
  {:else if lang === 'ko'}
    <link href="https://fonts.googleapis.com/css2?family=Noto+Sans+KR:wght@400;500;700&display=swap" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/static/pretendard.min.css" rel="stylesheet">
  {:else if lang === 'ar'}
    <link href="https://fonts.googleapis.com/css2?family=Cairo:wght@400;600;700&family=Tajawal:wght@400;500;700&display=swap" rel="stylesheet">
  {:else if lang === 'hi'}
    <link href="https://fonts.googleapis.com/css2?family=Noto+Sans+Devanagari:wght@400;500;700&display=swap" rel="stylesheet">
  {/if}
</svelte:head>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div data-lang={lang} dir={langInfo.dir} style="min-height:100vh;display:flex;flex-direction:column" role="presentation" onclick={onDocClick}>
  <header class="glass-header" style="{scrolled ? 'box-shadow:0 1px 12px rgba(0,0,0,0.06)' : ''}">
    <div class="container" style="display:flex;align-items:center;justify-content:space-between;height:4rem">
      <a href="/{lang}" inert={showMobileMenu || undefined} style="display:flex;align-items:center;gap:0.625rem;text-decoration:none">
        <div class="logo-icon">
          <span aria-hidden="true" style="color:white;font-weight:700;font-size:1.125rem;position:relative;z-index:1">R</span>
        </div>
        <span style="font-weight:700;font-size:1.125rem;color:var(--text);letter-spacing:-0.02em">ResumeTake</span>
      </a>
      <nav class="desktop-nav" aria-label={t.nav.menu} inert={showMobileMenu || undefined} style="display:flex;align-items:center;gap:0.5rem">
        <a href="/{lang}" class="btn btn-secondary" style="padding:0.5rem 1rem" aria-current={$page.url.pathname === `/${lang}` ? 'page' : undefined}>{t.nav.home}</a>
        <a href="/{lang}/editor" class="btn btn-secondary" style="padding:0.5rem 1rem" aria-current={$page.url.pathname.startsWith(`/${lang}/editor`) ? 'page' : undefined}>{t.nav.optimize}</a>
        <a href="/{lang}/generate" class="btn btn-secondary" style="padding:0.5rem 1rem" aria-current={$page.url.pathname.startsWith(`/${lang}/generate`) ? 'page' : undefined}>{t.nav.generate}</a>
        <a href="/{lang}/jobs" class="btn btn-secondary" style="padding:0.5rem 1rem" aria-current={$page.url.pathname.startsWith(`/${lang}/jobs`) ? 'page' : undefined}>{t.nav.jobs}</a>
        <a href="/{lang}/pricing" class="btn btn-secondary" style="padding:0.5rem 1rem" aria-current={$page.url.pathname.startsWith(`/${lang}/pricing`) ? 'page' : undefined}>{t.nav.pricing}</a>
        {#if isLoggedIn}
          <button class="btn btn-secondary" style="padding:0.5rem 1rem;font-size:0.8125rem" onclick={logout}>{t.nav.logout}</button>
        {:else}
          <a href="/{lang}/auth" class="btn btn-primary" style="padding:0.5rem 1.25rem;font-weight:600">
            <span>{t.nav.login}</span>
          </a>
        {/if}
        <button class="btn btn-secondary theme-toggle" onclick={toggleTheme} aria-label={theme === 'auto' ? t.nav.themeAuto : theme === 'dark' ? t.nav.themeDark : t.nav.themeLight} title={theme === 'auto' ? t.nav.themeAuto : theme === 'dark' ? t.nav.themeDark : t.nav.themeLight} style="padding:0.5rem 0.625rem;font-size:0.8125rem;display:flex;align-items:center;justify-content:center">
          {#if theme === 'dark'}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
          {:else if theme === 'light'}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/></svg>
          {:else}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M12 3v18" fill="currentColor"/><path d="M12 3a9 9 0 0 0 0 18z" fill="currentColor" stroke="none"/></svg>
          {/if}
        </button>
        <div data-lang-menu style="position:relative">
          <button class="btn btn-secondary" style="padding:0.5rem 0.75rem;font-size:0.8125rem;display:flex;align-items:center;gap:0.375rem" bind:this={langTriggerBtn} onclick={toggleLangMenu} onkeydown={(e) => { if (e.key === 'Escape') showLangMenu = false; }} aria-expanded={showLangMenu} aria-haspopup="menu" aria-label={t.nav.language}>
            <span>{langInfo.flag}</span>
            <span>{langInfo.name}</span>
            <svg width="10" height="10" viewBox="0 0 10 10" fill="none" style="transition:transform 0.2s;{showLangMenu ? 'transform:rotate(180deg)' : ''}" aria-hidden="true"><path d="M2 4l3 3 3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
          </button>
          {#if showLangMenu}
            <div class="lang-menu" role="menu" bind:this={langMenuRef}>
              {#each Object.entries(LANGUAGES) as [code, info] (code)}
                <button class="{code===lang ? 'active' : ''}" aria-current={code===lang ? 'true' : undefined} onclick={() => switchLanguage(code)} onkeydown={onLangMenuKeydown} role="menuitem" tabindex="-1">
                  <span style="font-size:1rem" aria-hidden="true">{info.flag}</span>
                  <span>{info.name}</span>
                </button>
              {/each}
            </div>
          {/if}
        </div>
      </nav>
      <button class="btn btn-secondary hamburger-btn" bind:this={hamburgerBtn} onclick={() => showMobileMenu = !showMobileMenu} aria-label={showMobileMenu ? t.nav.closeMenu : t.nav.menu} aria-expanded={showMobileMenu}>
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          {#if showMobileMenu}
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          {:else}
            <line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>
          {/if}
        </svg>
      </button>
    </div>
  </header>

  {#if showMobileMenu}
    <div class="mobile-menu" role="dialog" aria-modal="true" tabindex="-1" aria-label={t.nav.menu} onclick={() => showMobileMenu = false} onkeydown={(e) => { if (e.key === 'Escape') { e.preventDefault(); showMobileMenu = false; } trapMobileMenuFocus(e); }}>
      <div class="mobile-menu-inner" bind:this={mobileMenuPanel} role="presentation" onclick={(e) => e.stopPropagation()}>
        <button class="mobile-close-btn" type="button" onclick={() => showMobileMenu = false} aria-label={t.nav.closeMenu} style="position:absolute;top:0.75rem;right:0.75rem;background:none;border:none;font-size:1.5rem;color:var(--text-secondary);cursor:pointer;padding:0.25rem 0.5rem;line-height:1">×</button>
        <a href="/{lang}" class="btn btn-secondary mobile-link" aria-current={$page.url.pathname === `/${lang}` ? 'page' : undefined}>{t.nav.home}</a>
        <a href="/{lang}/editor" class="btn btn-secondary mobile-link" aria-current={$page.url.pathname.startsWith(`/${lang}/editor`) ? 'page' : undefined}>{t.nav.optimize}</a>
        <a href="/{lang}/generate" class="btn btn-secondary mobile-link" aria-current={$page.url.pathname.startsWith(`/${lang}/generate`) ? 'page' : undefined}>{t.nav.generate}</a>
        <a href="/{lang}/jobs" class="btn btn-secondary mobile-link" aria-current={$page.url.pathname.startsWith(`/${lang}/jobs`) ? 'page' : undefined}>{t.nav.jobs}</a>
        <a href="/{lang}/pricing" class="btn btn-secondary mobile-link" aria-current={$page.url.pathname.startsWith(`/${lang}/pricing`) ? 'page' : undefined}>{t.nav.pricing}</a>
        {#if isLoggedIn}
          <button class="btn btn-secondary mobile-link" onclick={logout}>{t.nav.logout}</button>
        {:else}
          <a href="/{lang}/auth" class="btn btn-primary mobile-link" style="font-weight:600">{t.nav.login}</a>
        {/if}
        <button class="btn btn-secondary mobile-link" onclick={toggleTheme} style="justify-content:center;gap:0.5rem">
          {#if theme === 'dark'}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
            <span>{t.nav.themeDark}</span>
          {:else if theme === 'light'}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/></svg>
            <span>{t.nav.themeLight}</span>
          {:else}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M12 3a9 9 0 0 0 0 18z" fill="currentColor" stroke="none"/></svg>
            <span>{t.nav.themeAuto}</span>
          {/if}
        </button>
        <div class="mobile-lang-row">
          {#each Object.entries(LANGUAGES) as [code, info] (code)}
            <button class="btn btn-secondary {code===lang ? 'active-lang' : ''}" aria-current={code===lang ? 'true' : undefined} style="font-size:0.8125rem;gap:0.25rem;display:flex;align-items:center" onclick={() => switchLanguage(code)}>
              <span aria-hidden="true">{info.flag}</span>
              <span>{info.name}</span>
            </button>
          {/each}
        </div>
      </div>
    </div>
  {/if}

  <main id="main-content" bind:this={mainEl} tabindex="-1" inert={showMobileMenu || undefined} style="flex:1;min-height:calc(100vh - 4rem)">
    {@render children()}
  </main>

  <footer inert={showMobileMenu || undefined} style="background:var(--bg-surface);border-top:1px solid var(--border);padding:2.5rem 0;margin-top:4rem">
    <div class="container" style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:1rem">
      <div style="display:flex;align-items:center;gap:0.625rem">
        <div class="logo-icon" style="width:1.75rem;height:1.75rem">
          <span aria-hidden="true" style="color:white;font-weight:700;font-size:0.875rem;position:relative;z-index:1">R</span>
        </div>
        <span style="color:var(--text-secondary);font-size:0.875rem">&copy; {footerYear || new Date().getFullYear()} ResumeTake. {t.footer.copyright}</span>
      </div>
      <nav aria-label={t.footer.footerNav} style="display:flex;gap:1.5rem;color:var(--text-secondary);font-size:0.875rem;flex-wrap:wrap">
        <a href="/{lang}/editor" style="color:var(--text-secondary);transition:color 0.2s;position:relative" class="footer-link">{t.footer.createResume}</a>
        <a href="/{lang}/jobs" style="color:var(--text-secondary);transition:color 0.2s;position:relative" class="footer-link">{t.nav.jobs}</a>
        <a href="/{lang}/pricing" style="color:var(--text-secondary);transition:color 0.2s;position:relative" class="footer-link">{t.nav.pricing}</a>
        <a href="/{lang}/templates" style="color:var(--text-secondary);transition:color 0.2s;position:relative" class="footer-link">{t.footer.templates}</a>
        <a href="/{lang}/privacy" style="color:var(--text-secondary);transition:color 0.2s;position:relative" class="footer-link">{t.footer.privacy}</a>
        <a href="/{lang}/terms" style="color:var(--text-secondary);transition:color 0.2s;position:relative" class="footer-link">{t.footer.terms}</a>
        <a href="/{lang}/contact" style="color:var(--text-secondary);transition:color 0.2s;position:relative" class="footer-link">{t.footer.contact}</a>
      </nav>
    </div>
  </footer>

  <CookieNotice lang={lang} />
</div>

<style>
  :global(.footer-link::after) {
    content: '';
    position: absolute;
    inset-inline-start: 0;
    bottom: -2px;
    width: 100%;
    height: 1.5px;
    background: var(--primary);
    transform: scaleX(0);
    transform-origin: inline-start;
    transition: transform 0.25s ease;
  }
  :global(.footer-link:hover::after) {
    transform: scaleX(1);
  }
  :global(.footer-link:hover) {
    color: var(--primary) !important;
  }
  .hamburger-btn {
    display: none;
    padding: 0.5rem;
  }
  .mobile-menu {
    position: fixed;
    inset: 0;
    top: 4rem;
    background: rgba(0, 0, 0, 0.4);
    z-index: 99;
    animation: fadeIn 0.15s ease;
  }
  .mobile-menu-inner {
    background: var(--bg-surface);
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    border-bottom: 1px solid var(--border);
  }
  .mobile-link {
    width: 100%;
    text-align: center;
    justify-content: center;
  }
  .mobile-lang-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.375rem;
    margin-top: 0.5rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--border);
  }
  :global(.active-lang) {
    background: var(--primary) !important;
    color: white !important;
    border-color: var(--primary) !important;
  }
  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  /* Visually mark the current page link for sighted users */
  .btn-secondary[aria-current="page"] {
    background: var(--primary);
    color: white;
    border-color: var(--primary);
  }
  @media (max-width: 768px) {
    .desktop-nav {
      display: none !important;
    }
    .hamburger-btn {
      display: flex !important;
    }
  }
</style>
