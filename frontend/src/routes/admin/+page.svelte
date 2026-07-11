<script>
  import { page } from '$app/stores';
  import { getTranslation, LANGUAGES } from '$lib/i18n/index.js';
  import { onMount } from 'svelte';

  // R58-F2: prevent state writes after unmount. Admin page is ssr=false,
  // so navigating away unmounts the component — but in-flight fetch
  // callbacks can still fire and write to $state, triggering Svelte 5
  // state_unsafe_mutation warnings.
  let mounted = false;
  onMount(() => {
    mounted = true;
    return () => { mounted = false; };
  });

  let lang = $derived.by(() => {
    const m = $page.url.pathname.match(/^\/([a-z]{2})(?:\/|$)/);
    if (m && LANGUAGES[m[1]]) return m[1];
    return 'en';
  });
  let t = $derived(getTranslation(lang));

  let token = $state('');
  let loggedIn = $state(false);
  let stats = $state(null);
  let users = $state([]);
  let error = $state('');
  let loading = $state(false);
  let searchQuery = $state('');
  let currentPage = $state(1);
  let totalPages = $state(1);
  const perPage = 50;

  // R50-F14: extract plan label lookup to a function — the inline ternary
  // with optional chaining (?.) in the template caused a Svelte parse error.
  function planLabel(plan) {
    if (!plan) return t.admin.planFree || '-';
    const key = 'plan' + plan[0].toUpperCase() + plan.slice(1);
    return t.admin[key] || plan;
  }

  // Admin uses a separate ADMIN_TOKEN (not the user's localStorage token),
  // so apiFetch can't be used here — it would inject the wrong credentials.
  // This local helper adds an AbortController timeout so a stalled backend
  // doesn't leave the dashboard hanging forever (mirrors apiFetch's pattern).
  async function adminFetch(url, tok, timeout = 30000) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);
    try {
      return await fetch(url, {
        headers: { 'Authorization': 'Bearer ' + tok },
        signal: controller.signal
      });
    } finally {
      clearTimeout(timeoutId);
    }
  }

  async function login() {
    if (!token || loading) return;
    loading = true;
    error = '';
    try {
      const res = await adminFetch(`/api/admin/users?page=${currentPage}&per_page=${perPage}`, token);
      if (!mounted) return;
      let data;
      try { data = await res.json(); } catch {
        if (!mounted) return;
        error = res.status >= 500 ? t.admin.networkError : t.admin.loginFailed;
        return;
      }
      if (res.ok && data.success) {
        loggedIn = true;
        stats = data.stats;
        users = data.users || [];
        if (data.pagination) totalPages = data.pagination.total_pages || 1;
      } else {
        error = data.message || t.admin.loginFailed;
      }
    } catch (e) {
      if (!mounted) return;
      error = t.admin.networkError;
    } finally {
      if (mounted) loading = false;
    }
  }

  function logout() {
    loggedIn = false;
    token = '';
    stats = null;
    users = [];
    error = '';
    searchQuery = '';
    currentPage = 1;
    totalPages = 1;
    loading = false; // R58-F1: reset loading so login button isn't disabled during in-flight refresh
  }

  async function refresh() {
    loading = true;
    error = '';
    try {
      const res = await adminFetch(`/api/admin/users?page=${currentPage}&per_page=${perPage}`, token);
      if (res.status === 401) {
        // Token expired or revoked — force re-authentication instead of
        // leaving the user in a stale "logged in" state where every action fails.
        loggedIn = false;
        token = '';
        stats = null;
        users = [];
        currentPage = 1;
        totalPages = 1;
        error = t.admin.sessionExpired;
        return false;
      }
      let data;
      try { data = await res.json(); } catch {
        error = res.status >= 500 ? t.admin.networkError : t.admin.refreshFailed;
        return false;
      }
      if (res.ok && data.success) {
        stats = data.stats;
        users = data.users || [];
        if (data.pagination) totalPages = data.pagination.total_pages || 1;
        return true;
      } else {
        error = data.message || t.admin.refreshFailed;
        return false;
      }
    } catch (e) {
      if (!mounted) return false;
      error = t.admin.refreshFailed;
      return false;
    } finally {
      if (mounted) loading = false;
    }
  }

  // Clamp page into [1, totalPages] before fetching to avoid requesting an
  // out-of-range page when totalPages hasn't been refreshed yet (e.g. user
  // clicks "next" right after the last item on the final page was deleted).
  async function goToPage(pageNum) {
    if (loading) return;
    const target = Math.max(1, Math.min(pageNum, totalPages));
    if (target === currentPage) return;
    const prevPage = currentPage;
    currentPage = target;
    const ok = await refresh();
    // R56b-F3: if refresh failed (network/non-ok), the users array still
    // holds the previous page's data but currentPage shows the new page —
    // rollback so the page indicator matches the displayed data. The 401
    // path already resets currentPage to 1, so only roll back when still
    // logged in and the fetch did not succeed.
    if (!ok && loggedIn) currentPage = prevPage;
  }

  let filteredUsers = $derived(
    searchQuery
      ? users.filter(u => {
          const q = searchQuery.toLowerCase();
          return (u.email || '').toLowerCase().includes(q) || (u.name || '').toLowerCase().includes(q);
        })
      : users
  );
</script>

<svelte:head>
  <title>{t.admin.title}</title>
  <!-- R39b-F-L7: noindex + canonical is contradictory — search engines see
       noindex and ignore the canonical, but the presence of canonical on a
       noindex page can cause confusion in some crawlers. Remove it. -->
  <meta name="robots" content="noindex, nofollow">
</svelte:head>

<main id="main-content" class="admin-page" tabindex="-1">
  {#if !loggedIn}
    <div class="login-box">
      <h1><span aria-hidden="true">🔒</span> {t.admin.loginTitle}</h1>
      <p>{t.admin.enterToken}</p>
      <form class="login-form" onsubmit={(e) => { e.preventDefault(); login(); }}>
        <input
          type="password"
          placeholder={t.admin.tokenPlaceholder}
          aria-label={t.admin.tokenAriaLabel}
          autocomplete="off"
          bind:value={token}
        />
        <button type="submit" disabled={loading}>
          {loading ? t.admin.loadingButton : t.admin.loginButton}
        </button>
      </form>
      {#if error}
        <div class="error" role="alert">{error}</div>
      {/if}
    </div>
  {:else}
    <div class="dashboard">
      <div class="header">
        <h1><span aria-hidden="true">📊</span> {t.admin.dashboardTitle}</h1>
        <div class="header-actions">
          <button class="refresh-btn" onclick={refresh} disabled={loading}>
            {#if loading}
              {t.admin.refreshingButton}
            {:else}
              <span aria-hidden="true">🔄</span> {t.admin.refreshButton}
            {/if}
          </button>
          <button class="refresh-btn" onclick={logout}>
            <span aria-hidden="true">🚪</span> {t.nav.logout}
          </button>
        </div>
      </div>

      {#if stats}
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-value">{stats.total}</div>
            <div class="stat-label">{t.admin.totalUsers}</div>
          </div>
          <div class="stat-card pro">
            <div class="stat-value">{stats.pro}</div>
            <div class="stat-label">{t.admin.proUsers}</div>
          </div>
          <div class="stat-card enterprise">
            <div class="stat-value">{stats.enterprise}</div>
            <div class="stat-label">{t.admin.enterprise}</div>
          </div>
          <div class="stat-card free">
            <div class="stat-value">{stats.free}</div>
            <div class="stat-label">{t.admin.freeUsers}</div>
          </div>
          <div class="stat-card usage">
            <div class="stat-value">{stats.total_usage}</div>
            <div class="stat-label">{t.admin.totalOptimizations}</div>
          </div>
        </div>
      {/if}

      <div class="search-bar">
        <input
          type="text"
          placeholder={t.admin.searchPlaceholder}
          aria-label={t.admin.searchAriaLabel}
          bind:value={searchQuery}
        />
        <span style="font-size:0.75rem;color:var(--text-secondary);margin-left:0.5rem">{t.admin.searchCurrentPageOnly || 'Current page only'}</span>
        <span class="count">{filteredUsers.length} {t.admin.usersCount}</span>
      </div>

      <div class="user-table">
        <table aria-label={t.admin.tableAriaLabel}>
          <thead>
            <tr>
              <th scope="col">{t.admin.thEmail}</th>
              <th scope="col">{t.admin.thName}</th>
              <th scope="col">{t.admin.thPlan}</th>
              <th scope="col">{t.admin.thUsage}</th>
              <th scope="col">{t.admin.thCreated}</th>
            </tr>
          </thead>
          <tbody>
            {#each filteredUsers as user (user.id)}
              <tr>
                <td class="email">{user.email}</td>
                <td>{user.name}</td>
                <td>
                  <span class="plan-badge {user.plan || 'free'}">{planLabel(user.plan)}</span>
                </td>
                <td>{user.usage_count}</td>
                <td class="date">{user.created_at && !isNaN(new Date(user.created_at).getTime()) ? new Date(user.created_at).toLocaleDateString(lang, { year: 'numeric', month: 'short', day: 'numeric' }) : '-'}</td>
              </tr>
            {:else}
              <tr><td colspan="5" style="text-align:center;padding:2rem;color:var(--text-secondary)">{t.admin.noResults || 'No users found'}</td></tr>
            {/each}
          </tbody>
        </table>
      </div>

      {#if totalPages > 1}
        <nav class="pagination" aria-label={t.admin.paginationNav}>
          <button
            class="page-btn"
            onclick={() => goToPage(currentPage - 1)}
            disabled={loading || currentPage <= 1}
            aria-label={t.admin.prevPage}
          >
            <span aria-hidden="true">←</span>
          </button>
          <span class="page-info" aria-live="polite">
            {currentPage} / {totalPages}
          </span>
          <button
            class="page-btn"
            onclick={() => goToPage(currentPage + 1)}
            disabled={loading || currentPage >= totalPages}
            aria-label={t.admin.nextPage}
          >
            <span aria-hidden="true">→</span>
          </button>
        </nav>
      {/if}
    </div>
  {/if}
</main>

<style>
  .admin-page {
    min-height: 100vh;
    background: var(--bg);
    color: var(--text);
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    padding: 2rem;
  }

  .login-box {
    max-width: 400px;
    margin: 8rem auto;
    text-align: center;
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 16px;
    padding: 3rem;
  }

  .login-box h1 {
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
  }

  .login-box p {
    color: var(--text-secondary);
    margin-bottom: 1.5rem;
  }

  .login-form {
    display: flex;
    gap: 0.5rem;
  }

  .login-form input {
    flex: 1;
    padding: 0.75rem 1rem;
    background: var(--bg-glass);
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--text);
    font-size: 0.9rem;
  }

  .login-form button {
    padding: 0.75rem 1.5rem;
    background: var(--primary);
    border: none;
    border-radius: 8px;
    color: #fff;
    font-weight: 600;
    cursor: pointer;
  }

  .login-form button:hover { background: var(--primary-hover); }
  .login-form button:disabled { opacity: 0.5; }

  .error {
    margin-top: 1rem;
    padding: 0.75rem;
    background: rgba(239,68,68,0.1);
    border: 1px solid rgba(239,68,68,0.3);
    border-radius: 8px;
    color: #dc2626;
    font-size: 0.875rem;
  }

  .dashboard { max-width: 1200px; margin: 0 auto; }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  .header h1 { font-size: 1.75rem; }

  .header-actions { display: flex; gap: 0.5rem; }

  .refresh-btn {
    padding: 0.5rem 1rem;
    background: var(--bg-glass);
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--text);
    cursor: pointer;
    font-size: 0.875rem;
  }

  .refresh-btn:hover { background: var(--bg-surface); }
  .refresh-btn:disabled { opacity: 0.5; }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 1rem;
    margin-bottom: 2rem;
  }

  .stat-card {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 1.5rem;
    text-align: center;
  }

  .stat-card.pro { border-color: rgba(67,56,202,0.5); }
  .stat-card.enterprise { border-color: rgba(126,34,206,0.5); }
  .stat-card.free { border-color: rgba(4,120,87,0.5); }
  .stat-card.usage { border-color: rgba(180,83,9,0.5); }

  .stat-value {
    font-size: 2rem;
    font-weight: 800;
    margin-bottom: 0.25rem;
  }

  .stat-card.pro .stat-value { color: #4338ca; }
  .stat-card.enterprise .stat-value { color: #7e22ce; }
  .stat-card.free .stat-value { color: #047857; }
  .stat-card.usage .stat-value { color: #b45309; }

  .stat-label {
    font-size: 0.8rem;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .search-bar {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .search-bar input {
    flex: 1;
    padding: 0.75rem 1rem;
    background: var(--bg-glass);
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--text);
    font-size: 0.9rem;
  }

  .search-bar .count {
    color: var(--text-secondary);
    font-size: 0.875rem;
    white-space: nowrap;
  }

  .user-table {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 12px;
    overflow-x: auto;
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  th {
    text-align: left;
    padding: 1rem;
    background: var(--bg-glass);
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-secondary);
    border-bottom: 1px solid var(--border);
  }

  td {
    padding: 0.875rem 1rem;
    border-bottom: 1px solid var(--border);
    font-size: 0.9rem;
  }

  tr:hover { background: var(--bg-glass); }

  .email { color: var(--primary); }

  .date { color: var(--text-secondary); font-size: 0.8rem; }

  .plan-badge {
    display: inline-block;
    padding: 0.2rem 0.6rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
  }

  .plan-badge.free {
    background: rgba(4,120,87,0.15);
    color: #047857;
  }

  .plan-badge.pro {
    background: rgba(67,56,202,0.15);
    color: #4338ca;
  }

  .plan-badge.enterprise {
    background: rgba(126,34,206,0.15);
    color: #7e22ce;
  }
  /* Dark mode: use lighter shades for WCAG AA contrast on #111827 bg */
  :global([data-theme="dark"]) .stat-card.pro .stat-value { color: #818cf8; }
  :global([data-theme="dark"]) .stat-card.enterprise .stat-value { color: #c084fc; }
  :global([data-theme="dark"]) .stat-card.free .stat-value { color: #34d399; }
  :global([data-theme="dark"]) .stat-card.usage .stat-value { color: #fbbf24; }
  :global([data-theme="dark"]) .plan-badge.pro { color: #818cf8; background: rgba(129,140,248,0.15); }
  :global([data-theme="dark"]) .plan-badge.enterprise { color: #c084fc; background: rgba(192,132,252,0.15); }
  :global([data-theme="dark"]) .plan-badge.free { color: #34d399; background: rgba(52,211,153,0.15); }

  .pagination {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    margin-top: 1.25rem;
  }

  .page-btn {
    width: 2.25rem;
    height: 2.25rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--text);
    cursor: pointer;
    font-size: 1rem;
    transition: background 0.15s ease;
  }

  .page-btn:hover:not(:disabled) { background: var(--bg-glass); }
  .page-btn:disabled { opacity: 0.4; cursor: not-allowed; }

  .page-info {
    color: var(--text-secondary);
    font-size: 0.875rem;
    font-variant-numeric: tabular-nums;
    min-width: 4rem;
    text-align: center;
  }

  @media (max-width: 768px) {
    .stats-grid { grid-template-columns: repeat(2, 1fr); }
    .stats-grid > :last-child { grid-column: 1 / -1; }
    .admin-page { padding: 1rem; }
  }
</style>
