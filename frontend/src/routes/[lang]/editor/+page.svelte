<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';
  import { onMount } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  let resume = $state({ name: '', email: '', phone: '', summary: '' });
  let targetJob = $state('');
  let jobDescription = $state('');
  let isOptimizing = $state(false);
  let result = $state(null);
  let tab = $state('edit');
  let mounted = $state(false);

  onMount(() => { mounted = true; });

  async function optimize() {
    isOptimizing = true;
    try {
      const res = await fetch('/api/v1/optimize', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ resume_content: resume, target_job: targetJob, job_description: jobDescription, lang })
      });
      const data = await res.json();
      result = data.data;
      tab = 'result';
    } catch { alert(lang === 'zh' ? '优化失败，请重试' : 'Optimization failed, please try again'); }
    finally { isOptimizing = false; }
  }
</script>

<svelte:head>
  <title>{t.meta.editorTitle}</title>
  <meta name="description" content={t.meta.editorDesc}>
  <meta name="keywords" content={t.meta.editorKeywords}>
  <link rel="canonical" href="https://resume.takee.top/{lang}/editor">
  <meta property="og:title" content={t.meta.editorTitle}>
  <meta property="og:description" content={t.meta.editorDesc}>
</svelte:head>

<!-- Editor Header -->
<div class="editor-header" style="padding:2.5rem 0 3rem;position:relative;overflow:hidden">
  <div class="orb orb-blue animate-float" style="width:200px;height:200px;top:-20%;left:10%"></div>
  <div class="orb orb-purple animate-float" style="width:160px;height:160px;bottom:-10%;right:15%;animation-delay:2s"></div>
  <div class="container" style="position:relative">
    <div class="{mounted ? 'animate-fade-in-up' : ''}" style="opacity:0">
      <h1 style="font-size:clamp(1.5rem,3vw,2rem);font-weight:700;margin-bottom:0.5rem;color:var(--text)">{t.editor.title}</h1>
      <p style="color:var(--text-secondary);font-size:0.9375rem">{t.editor.subtitle}</p>
    </div>
  </div>
</div>

<div class="container" style="padding:2rem 1.5rem;margin-top:-1rem">
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:2rem" class="editor-grid">
    <!-- Left: Input -->
    <div style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="editor-card {mounted ? 'animate-fade-in-up delay-1' : ''}" style="opacity:0">
        <h2 style="font-weight:600;margin-bottom:1.25rem;color:var(--text);display:flex;align-items:center;gap:0.5rem">
          <span style="font-size:1.125rem">📋</span> {t.editor.basicInfo}
        </h2>
        <div style="display:flex;flex-direction:column;gap:1rem">
          <div>
            <label class="label" for="rname">{t.editor.name}</label>
            <input id="rname" class="input" placeholder={t.editor.namePlaceholder} bind:value={resume.name}>
          </div>
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:1rem">
            <div>
              <label class="label" for="remail">{t.editor.email}</label>
              <input id="remail" type="email" class="input" placeholder={t.editor.emailPlaceholder} bind:value={resume.email}>
            </div>
            <div>
              <label class="label" for="rphone">{t.editor.phone}</label>
              <input id="rphone" type="tel" class="input" placeholder={t.editor.phonePlaceholder} bind:value={resume.phone}>
            </div>
          </div>
          <div>
            <label class="label" for="rsummary">{t.editor.summary}</label>
            <textarea id="rsummary" class="input" rows="4" placeholder={t.editor.summaryPlaceholder} bind:value={resume.summary} style="resize:vertical"></textarea>
          </div>
        </div>
      </div>

      <div class="editor-card {mounted ? 'animate-fade-in-up delay-2' : ''}" style="opacity:0">
        <h2 style="font-weight:600;margin-bottom:1.25rem;color:var(--text);display:flex;align-items:center;gap:0.5rem">
          <span style="font-size:1.125rem">🎯</span> {t.editor.targetJob}
        </h2>
        <div style="display:flex;flex-direction:column;gap:1rem">
          <div>
            <label class="label" for="tjob">{t.editor.targetJob}</label>
            <input id="tjob" class="input" placeholder={t.editor.targetJobPlaceholder} bind:value={targetJob}>
          </div>
          <div>
            <label class="label" for="jdesc">{t.editor.jobDesc}</label>
            <textarea id="jdesc" class="input" rows="4" placeholder={t.editor.jobDescPlaceholder} bind:value={jobDescription} style="resize:vertical"></textarea>
          </div>
        </div>
      </div>

      <button class="optimize-btn {mounted ? 'animate-fade-in-up delay-3' : ''}" style="opacity:0" onclick={optimize} disabled={isOptimizing}>
        {#if isOptimizing}
          <span style="display:flex;align-items:center;gap:0.5rem;position:relative;z-index:1">
            <span class="spinner"></span> {t.editor.optimizing}
          </span>
        {:else}
          <span style="display:flex;align-items:center;gap:0.5rem;position:relative;z-index:1">
            <span>✨</span> {t.editor.optimizeBtn}
          </span>
        {/if}
      </button>
    </div>

    <!-- Right: Preview/Result -->
    <div>
      <div class="editor-card {mounted ? 'animate-fade-in-up delay-2' : ''}" style="min-height:500px;opacity:0">
        <div style="display:flex;gap:0;border-bottom:1px solid var(--border);margin-bottom:1.25rem">
          {#each [{id:'edit',label:t.editor.previewTab},{id:'result',label:t.editor.resultTab}] as t2}
            <button class="tab-btn {tab===t2.id ? 'active' : ''}" onclick={()=>tab=t2.id}>{t2.label}</button>
          {/each}
        </div>

        {#if tab==='edit'}
          <div class="animate-fade-in" style="padding:1rem 0">
            <h3 style="font-size:1.25rem;font-weight:600;color:var(--text)">{resume.name||t.editor.defaultName}</h3>
            <p style="font-size:0.875rem;color:var(--text-secondary);margin-top:0.25rem">{resume.email||'email@example.com'} | {resume.phone||'+1 234 567 890'}</p>
            <div style="margin-top:1.5rem;padding-top:1.5rem;border-top:1px solid var(--border)">
              <h4 style="font-weight:500;margin-bottom:0.5rem;color:var(--text);font-size:0.875rem;text-transform:uppercase;letter-spacing:0.05em">{t.editor.summary}</h4>
              <p style="line-height:1.7;color:var(--text-secondary);font-size:0.9375rem">{resume.summary||t.editor.defaultSummary}</p>
            </div>
          </div>
        {:else if result}
          <div class="animate-fade-in" style="display:flex;flex-direction:column;gap:1.25rem">
            <div class="result-score">
              <div>
                <span style="font-weight:600;color:#059669">{t.editor.atsScore}</span>
                <p style="font-size:0.8125rem;color:#059669;opacity:0.7;margin-top:0.125rem">{lang === 'zh' ? 'AI分析结果' : 'AI Analysis Result'}</p>
              </div>
              <span style="font-size:2rem;font-weight:800;color:#059669">{result.ats_score}%</span>
            </div>
            <div>
              <h4 style="font-weight:500;margin-bottom:0.75rem;color:var(--text);display:flex;align-items:center;gap:0.375rem">
                <span>🔑</span> {t.editor.keywords}
              </h4>
              <div style="display:flex;flex-wrap:wrap;gap:0.5rem">
                {#each result.keywords||[] as kw}
                  <span class="keyword-tag">{kw}</span>
                {/each}
              </div>
            </div>
            <div>
              <h4 style="font-weight:500;margin-bottom:0.75rem;color:var(--text);display:flex;align-items:center;gap:0.375rem">
                <span>💡</span> {t.editor.suggestions}
              </h4>
              <ul style="list-style:none;display:flex;flex-direction:column;gap:0.625rem">
                {#each result.suggestions||[] as s}
                  <li style="font-size:0.9375rem;color:var(--text-secondary);display:flex;gap:0.625rem;align-items:flex-start;line-height:1.5">
                    <span style="color:var(--primary);margin-top:0.125rem;flex-shrink:0">→</span>
                    <span>{s}</span>
                  </li>
                {/each}
              </ul>
            </div>
          </div>
        {:else}
          <div class="animate-fade-in" style="text-align:center;padding:4rem 0;color:var(--text-secondary)">
            <div style="font-size:3rem;margin-bottom:1rem;opacity:0.4;animation:floatSlow 3s ease-in-out infinite">✨</div>
            <p style="font-size:0.9375rem">{t.editor.emptyResult}</p>
            <p style="font-size:0.8125rem;opacity:0.5;margin-top:0.375rem">{lang === 'zh' ? '填写信息后点击优化按钮' : 'Fill in the form and click Optimize'}</p>
          </div>
        {/if}
      </div>
      <button class="btn btn-secondary" style="width:100%;padding:0.875rem;margin-top:1rem;font-size:0.9375rem;display:flex;align-items:center;justify-content:center;gap:0.5rem">
        <span>📄</span> {t.editor.exportPdf}
      </button>
    </div>
  </div>
</div>

<style>
  .spinner {
    display: inline-block;
    width: 16px; height: 16px;
    border: 2px solid rgba(255,255,255,0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
  @media (max-width: 768px) {
    :global(.editor-grid) {
      grid-template-columns: 1fr !important;
    }
  }
</style>
