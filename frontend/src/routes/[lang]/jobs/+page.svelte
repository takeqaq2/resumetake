<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';
  import { onMount } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  let jobs = $state([]);
  let loading = $state(true);
  let error = $state('');
  let search = $state('');
  let expandedId = $state(null);

  let filteredJobs = $derived(
    search.trim()
      ? jobs.filter(j =>
          j.title?.toLowerCase().includes(search.toLowerCase()) ||
          j.company?.toLowerCase().includes(search.toLowerCase()) ||
          j.location?.toLowerCase().includes(search.toLowerCase()) ||
          (j.tags || []).some(tag => tag.toLowerCase().includes(search.toLowerCase()))
        )
      : jobs
  );

  onMount(async () => {
    try {
      const res = await fetch('/api/v1/jobs');
      const data = await res.json();
      if (data.success && data.data) {
        jobs = data.data;
      } else {
        jobs = [];
      }
    } catch {
      error = lang === 'zh' ? '加载职位失败' : 'Failed to load jobs';
    } finally {
      loading = false;
    }
  });

  function toggleExpand(id) {
    expandedId = expandedId === id ? null : id;
  }

  function getTypeLabel(type, lang) {
    const map = {
      'full-time': { zh: '全职', en: 'Full-time', ja: '正社員', ko: '정규직', es: 'Tiempo completo', pt: 'Integral', fr: 'Temps plein', de: 'Vollzeit', ar: 'دوام كامل', hi: 'पूर्णकालिक' },
      'part-time': { zh: '兼职', en: 'Part-time', ja: 'パート', ko: '파트타임', es: 'Medio tiempo', pt: 'Meio período', fr: 'Mi-temps', de: 'Teilzeit', ar: 'دوام جزئي', hi: 'अंशकालिक' },
      'intern': { zh: '实习', en: 'Intern', ja: 'インターン', ko: '인턴', es: 'Pasantía', pt: 'Estágio', fr: 'Stage', de: 'Praktikum', ar: 'تدريب', hi: 'इंटर्नशिप' }
    };
    return map[type]?.[lang] || map[type]?.en || type;
  }
</script>

<svelte:head>
  <title>{t.jobs.title} - ResumeTake</title>
</svelte:head>

<div class="jobs-header">
  <div class="orb orb-blue animate-float" style="width:200px;height:200px;top:-20%;left:10%"></div>
  <div class="orb orb-purple animate-float" style="width:160px;height:160px;bottom:-10%;right:15%;animation-delay:2s"></div>
  <div class="container" style="position:relative">
    <h1 style="font-size:clamp(1.5rem,3vw,2rem);font-weight:700;margin-bottom:0.5rem">{t.jobs.title}</h1>
    <p style="color:var(--text-secondary);font-size:0.9375rem">{t.jobs.subtitle}</p>
  </div>
</div>

<div class="container" style="padding:2rem 1.5rem;margin-top:-1rem">
  <div class="jobs-search">
    <div class="search-icon">🔍</div>
    <input class="input search-input" placeholder={t.jobs.search} bind:value={search}>
  </div>

  {#if loading}
    <div class="jobs-loading">
      <div class="jobs-spinner"></div>
      <span>{lang === 'zh' ? '加载中...' : 'Loading...'}</span>
    </div>
  {:else if error}
    <div class="jobs-error">{error}</div>
  {:else if filteredJobs.length === 0}
    <div class="jobs-empty">
      <div style="font-size:3rem;margin-bottom:1rem;opacity:0.3">📋</div>
      <p>{lang === 'zh' ? '暂无职位信息' : 'No jobs available'}</p>
    </div>
  {:else}
    <div class="jobs-grid">
      {#each filteredJobs as job (job.id || job.title)}
        <div class="job-card card {expandedId === job.id ? 'expanded' : ''}" role="button" tabindex="0" onclick={() => toggleExpand(job.id)} onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') toggleExpand(job.id); }}>
          <div class="job-card-header">
            <div>
              <h3 class="job-title">{job.title}</h3>
              <p class="job-company">{job.company}</p>
            </div>
            {#if job.type}
              <span class="job-badge">{getTypeLabel(job.type, lang)}</span>
            {/if}
          </div>
          <div class="job-meta">
            {#if job.location}
              <span class="job-meta-item">📍 {job.location}</span>
            {/if}
            {#if job.salary}
              <span class="job-meta-item">💰 {job.salary}</span>
            {/if}
          </div>
          {#if job.tags?.length}
            <div class="job-tags">
              {#each job.tags.slice(0, 5) as tag}
                <span class="job-tag">{tag}</span>
              {/each}
            </div>
          {/if}
          {#if expandedId === job.id && job.description}
            <div class="job-desc">
              <p>{job.description}</p>
              {#if job.url}
                <a href={job.url} target="_blank" rel="noopener" class="btn btn-primary" style="margin-top:1rem;display:inline-flex" onclick={(e) => e.stopPropagation()}>
                  {t.jobs.apply} →
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
  .search-icon { font-size: 1.25rem; position: absolute; left: 1rem; z-index: 1; }
  .search-input { padding-left: 2.75rem !important; font-size: 0.9375rem; }
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
  @keyframes spin { to { transform: rotate(360deg); } }
  .jobs-empty {
    text-align: center; padding: 4rem 0;
    color: var(--text-secondary);
  }
  .jobs-error {
    text-align: center; padding: 2rem; color: #ef4444;
    background: rgba(239,68,68,0.08); border-radius: var(--radius);
    border: 1px solid rgba(239,68,68,0.2);
  }
  @media (max-width: 768px) {
    .jobs-grid { grid-template-columns: 1fr; }
  }
</style>
