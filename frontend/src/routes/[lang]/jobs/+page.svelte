<script>
  import { page } from '$app/stores';
  import { env } from '$env/dynamic/public';
  import { getTranslation } from '$lib/i18n/index.js';
  import AdSlot from '$lib/AdSlot.svelte';
  import { onMount } from 'svelte';
  import { replaceState } from '$app/navigation';
  import { apiFetch } from '$lib/api.js';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  let jobs = $state([]);
  let loading = $state(true);
  let error = $state('');
  let search = $state('');
  let expandedId = $state(null);
  const jobsAdSlot = env.PUBLIC_AD_SLOT_JOBS || '';
  // R56b-F4: track whether the initial search value was hydrated from the
  // URL so the $effect that syncs search→URL doesn't overwrite the query
  // param on the first render before mount completes.
  let searchInitialized = false;

  let filteredJobs = $derived.by(() => {
    const q = search.trim().toLowerCase();
    if (!q) return jobs;
    return jobs.filter(j =>
      j.title?.toLowerCase().includes(q) ||
      j.company?.toLowerCase().includes(q) ||
      j.location?.toLowerCase().includes(q) ||
      (j.tags || []).some(tag => String(tag).toLowerCase().includes(q))
    );
  });

  // R56-F4: AbortController for the jobs fetch — aborted on unmount so the
  // in-flight request doesn't continue consuming bandwidth after navigation.
  let jobsAbort = null;

  // R56-F5: extracted from onMount so the retry button can re-invoke it.
  async function loadJobs() {
    if (jobsAbort) jobsAbort.abort();
    jobsAbort = new AbortController();
    loading = true;
    error = '';
    try {
      const res = await apiFetch('/api/v1/jobs', { skipAuth: true, signal: jobsAbort.signal });
      // R47-F5: guard against non-JSON responses (e.g. 502 gateway error
      // returning HTML). Without this, res.json() throws SyntaxError and
      // the user sees a generic "loadFailed" instead of a meaningful error.
      let data;
      try {
        data = await res.json();
      } catch {
        if (jobsAbort.signal.aborted) return;
        if (res.status === 429) {
          error = t.editor?.rateLimited || t.jobs.loadFailed;
        } else if (!res.ok) {
          error = t.editor?.serverError || t.jobs.loadFailed;
        } else {
          error = t.jobs.loadFailed;
        }
        return;
      }
      if (jobsAbort.signal.aborted) return;
      if (data.success && data.data) {
        jobs = data.data;
      } else {
        error = data.error || data.message || t.jobs.loadFailed;
      }
    } catch (e) {
      if (e?.name === 'AbortError' && jobsAbort?.signal.aborted) return;
      error = t.jobs.loadFailed;
    } finally {
      if (!jobsAbort.signal.aborted) loading = false;
    }
  }

  onMount(() => {
    // R56b-F4: hydrate search from the URL so a refreshed/shared link
    // restores the user's query. Using replaceState in the $effect below
    // keeps history clean (no new entry per keystroke).
    const q = $page.url.searchParams.get('q');
    if (q) search = q;
    searchInitialized = true;
    loadJobs();
    return () => { if (jobsAbort) jobsAbort.abort(); };
  });

  // R56b-F4: sync search→URL (debounced) so the query survives refresh and
  // is shareable. Debounce avoids hammering history.replaceState on every
  // keystroke. replaceState (not goto) keeps the back button usable.
  let searchSyncTimer = null;
  $effect(() => {
    if (!searchInitialized) return;
    const q = search.trim();
    if (searchSyncTimer) clearTimeout(searchSyncTimer);
    searchSyncTimer = setTimeout(() => {
      const url = new URL($page.url);
      if (q) url.searchParams.set('q', q);
      else url.searchParams.delete('q');
      replaceState(url, '');
    }, 300);
  });

  // R56b-F4: announce result count to screen readers. aria-live="polite"
  // on a visually-hidden span lets SR users know how many results matched
  // after they type, without moving focus.
  let resultsAnnouncement = $derived.by(() => {
    if (loading || error) return '';
    const count = filteredJobs.length;
    if (count === 0) return t.jobs.noResults;
    return count + ' / ' + jobs.length;
  });

  function toggleExpand(id) {
    expandedId = expandedId === id ? null : id;
  }

  const typeLabel = (type) => {
    const map = { 'full-time': t.jobs.fullTime, 'part-time': t.jobs.partTime, 'intern': t.jobs.intern };
    // R50-F4: unknown types (e.g. contract, remote) previously returned the
    // raw English enum string. Fall back to the localized "other" label.
    return map[type] || t.jobs.other;
  };
</script>

<svelte:head>
  <title>{t.meta.jobsTitle}</title>
  <meta name="description" content={t.meta.jobsDesc}>
  <link rel="canonical" href="https://resume.takee.top/{lang}/jobs">
  <meta property="og:title" content={t.meta.jobsTitle}>
  <meta property="og:description" content={t.meta.jobsDesc}>
  <meta property="og:url" content="https://resume.takee.top/{lang}/jobs">
  <meta property="og:type" content="website">
</svelte:head>

<div class="jobs-header">
  <div class="orb orb-blue animate-float" aria-hidden="true" style="width:200px;height:200px;top:-20%;left:10%"></div>
  <div class="orb orb-purple animate-float" aria-hidden="true" style="width:160px;height:160px;bottom:-10%;right:15%;animation-delay:2s"></div>
  <div class="container" style="position:relative">
    <h1 style="font-size:clamp(1.5rem,3vw,2rem);font-weight:700;margin-bottom:0.5rem">{t.jobs.title}</h1>
    <p style="color:var(--text-secondary);font-size:0.9375rem">{t.jobs.subtitle}</p>
  </div>
</div>

<div class="container" style="padding:2rem 1.5rem;margin-top:-1rem">
  <div class="jobs-search">
    <div class="search-icon" aria-hidden="true">🔍</div>
    <input class="input search-input" placeholder={t.jobs.search} aria-label={t.jobs.search} bind:value={search}>
  </div>
  <!-- R56b-F4: visually-hidden live region announces result count to
       screen readers after the user types in the search box. -->
  <span class="sr-only" role="status" aria-live="polite">{resultsAnnouncement}</span>

  <AdSlot slot={jobsAdSlot} label={t.ads.label} />

  {#if loading}
    <div class="jobs-loading" role="status" aria-live="polite">
      <div class="jobs-spinner"></div>
      <span>{t.jobs.loading}</span>
    </div>
  {:else if error}
    <div class="jobs-error" role="alert">{error}</div>
    <div style="text-align:center;margin-top:1rem">
      <button class="btn btn-secondary" onclick={loadJobs}>{t.pricing?.retry || 'Retry'}</button>
    </div>
  {:else if filteredJobs.length === 0}
    <div class="jobs-empty">
      <div style="font-size:3rem;margin-bottom:1rem;opacity:0.3" aria-hidden="true">📋</div>
      <p>{jobs.length === 0 ? t.jobs.noJobs : t.jobs.noResults}</p>
    </div>
  {:else}
    <div class="jobs-grid">
      {#each filteredJobs as job, i (job.id || job.title + '-' + i)}
        {@const jobId = job.id || (job.title ? job.title + '-' + i : 'job-' + i)}
        <div class="job-card card {expandedId === jobId ? 'expanded' : ''}" role="button" tabindex="0" aria-expanded={expandedId === jobId} aria-controls={expandedId === jobId ? 'job-desc-' + jobId : undefined} aria-label={(job.title ? job.title + ' — ' : '') + (expandedId === jobId ? t.jobs.collapseAria : t.jobs.expandAria)} onclick={() => toggleExpand(jobId)} onkeydown={(e) => { if (e.currentTarget !== e.target) return; if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggleExpand(jobId); } else if (e.key === 'Escape' && expandedId === jobId) { e.preventDefault(); toggleExpand(jobId); } }}>
          <div class="job-card-header">
            <div>
              <h3 class="job-title">{job.title}</h3>
              <p class="job-company">{job.company}</p>
            </div>
            {#if job.type}
              <span class="job-badge">{typeLabel(job.type)}</span>
            {/if}
          </div>
          <div class="job-meta">
            {#if job.location}
              <span class="job-meta-item"><span aria-hidden="true">📍</span> {job.location}</span>
            {/if}
            {#if job.salary}
              <span class="job-meta-item"><span aria-hidden="true">💰</span> {job.salary}</span>
            {/if}
          </div>
          {#if job.tags?.length}
            <div class="job-tags">
              {#each job.tags.slice(0, 5) as tag (tag)}
                <span class="job-tag">{tag}</span>
              {/each}
            </div>
          {/if}
          {#if expandedId === jobId && job.description}
            <div id="job-desc-{jobId}" class="job-desc" role="region" aria-label={t.jobs.descriptionLabel}>
              <p>{job.description}</p>
              {#if job.url && /^https?:\/\//i.test(job.url)}
                <a href={job.url} target="_blank" rel="noopener noreferrer" class="btn btn-primary" style="margin-top:1rem;display:inline-flex" onclick={(e) => e.stopPropagation()}>
                  {t.jobs.apply} <span aria-hidden="true">→</span>
                </a>
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .jobs-header {
    position: relative; overflow: hidden; padding: 2rem 0 3rem;
    background: var(--gradient-hero); background-size: 200% 200%;
    animation: gradientShift 10s ease-in-out infinite;
  }
  .jobs-search {
    display: flex; align-items: center; gap: 0.75rem;
    margin-bottom: 2rem; position: relative;
  }
  .search-icon { font-size: 1.25rem; position: absolute; inset-inline-start: 1rem; z-index: 1; }
  .search-input { padding-inline-start: 2.75rem; font-size: 0.9375rem; }
  .jobs-grid {
    display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
    gap: 1.25rem;
  }
  .job-card {
    cursor: pointer; transition: all 0.3s; padding: 1.5rem;
  }
  .job-card.expanded { border-color: var(--primary); }
  .job-card-header {
    display: flex; justify-content: space-between; align-items: flex-start;
    gap: 1rem; margin-bottom: 0.75rem;
  }
  .job-title {
    font-weight: 600; font-size: 1.0625rem; color: var(--text);
    margin-bottom: 0.25rem;
  }
  .job-company { font-size: 0.875rem; color: var(--primary); font-weight: 500; }
  .job-badge {
    padding: 0.25rem 0.75rem; border-radius: 9999px; font-size: 0.75rem;
    font-weight: 500; white-space: nowrap;
    background: rgba(37,99,235,0.08); color: var(--primary);
    border: 1px solid rgba(37,99,235,0.15);
  }
  .job-meta {
    display: flex; gap: 1rem; flex-wrap: wrap; margin-bottom: 0.75rem;
  }
  .job-meta-item {
    font-size: 0.8125rem; color: var(--text-secondary);
  }
  .job-tags { display: flex; flex-wrap: wrap; gap: 0.375rem; }
  .job-tag {
    padding: 0.1875rem 0.5rem; border-radius: 9999px;
    background: var(--bg-surface); border: 1px solid var(--border);
    font-size: 0.75rem; color: var(--text-secondary);
  }
  .job-desc {
    margin-top: 1rem; padding-top: 1rem;
    border-top: 1px solid var(--border);
    font-size: 0.875rem; color: var(--text-secondary); line-height: 1.6;
  }
  .jobs-loading {
    display: flex; flex-direction: column; align-items: center;
    gap: 1rem; padding: 4rem 0; color: var(--text-secondary);
  }
  .jobs-spinner {
    width: 32px; height: 32px; border: 3px solid var(--border);
    border-top-color: var(--primary); border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
  .jobs-empty {
    text-align: center; padding: 4rem 0;
    color: var(--text-secondary);
  }
  .jobs-error {
    text-align: center; padding: 2rem; color: var(--error-text);
    background: rgba(239,68,68,0.08); border-radius: var(--radius);
    border: 1px solid rgba(239,68,68,0.2);
  }
  @media (max-width: 768px) {
    .jobs-grid { grid-template-columns: 1fr; }
  }
  @media (prefers-reduced-motion: reduce) {
    .jobs-header { animation: none !important; }
    .jobs-spinner { display: none; }
  }
</style>
