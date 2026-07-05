<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  let resume = $state({ name: '', email: '', phone: '', summary: '' });
  let targetJob = $state('');
  let jobDescription = $state('');
  let isOptimizing = $state(false);
  let result = $state(null);
  let tab = $state('edit');

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

<div class="container" style="padding:2.5rem 1.5rem">
  <h1 style="font-size:1.875rem;font-weight:700;margin-bottom:0.5rem;color:var(--text)">{t.editor.title}</h1>
  <p style="color:var(--text-secondary);margin-bottom:2rem">{t.editor.subtitle}</p>

  <div style="display:grid;grid-template-columns:1fr 1fr;gap:2rem">
    <div style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="card">
        <h2 style="font-weight:600;margin-bottom:1.25rem;color:var(--text)">{t.editor.basicInfo}</h2>
        <div style="display:flex;flex-direction:column;gap:1rem">
          <div><label class="label" for="rname">{t.editor.name}</label><input id="rname" class="input" placeholder={t.editor.namePlaceholder} bind:value={resume.name}></div>
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:1rem">
            <div><label class="label" for="remail">{t.editor.email}</label><input id="remail" type="email" class="input" placeholder={t.editor.emailPlaceholder} bind:value={resume.email}></div>
            <div><label class="label" for="rphone">{t.editor.phone}</label><input id="rphone" type="tel" class="input" placeholder={t.editor.phonePlaceholder} bind:value={resume.phone}></div>
          </div>
          <div><label class="label" for="rsummary">{t.editor.summary}</label><textarea id="rsummary" class="input" rows="4" placeholder={t.editor.summaryPlaceholder} bind:value={resume.summary}></textarea></div>
        </div>
      </div>
      <div class="card">
        <h2 style="font-weight:600;margin-bottom:1.25rem;color:var(--text)">{t.editor.targetJob}</h2>
        <div style="display:flex;flex-direction:column;gap:1rem">
          <div><label class="label" for="tjob">{t.editor.targetJob}</label><input id="tjob" class="input" placeholder={t.editor.targetJobPlaceholder} bind:value={targetJob}></div>
          <div><label class="label" for="jdesc">{t.editor.jobDesc}</label><textarea id="jdesc" class="input" rows="4" placeholder={t.editor.jobDescPlaceholder} bind:value={jobDescription}></textarea></div>
        </div>
      </div>
      <button class="btn btn-primary" style="width:100%;padding:0.875rem;font-size:1rem" onclick={optimize} disabled={isOptimizing}>
        {isOptimizing ? t.editor.optimizing : t.editor.optimizeBtn}
      </button>
    </div>

    <div>
      <div class="card" style="min-height:500px">
        <div style="display:flex;gap:1rem;border-bottom:1px solid var(--border);margin-bottom:1.25rem">
          {#each [{id:'edit',label:t.editor.previewTab},{id:'result',label:t.editor.resultTab}] as t2}
            <button class="btn" style="padding:0.5rem 1rem;border-bottom:2px solid {tab===t2.id?'var(--primary)':'transparent'};color:{tab===t2.id?'var(--primary)':'var(--text-secondary)'};border-radius:0;border-left:none;border-right:none;border-top:none" onclick={()=>tab=t2.id}>{t2.label}</button>
          {/each}
        </div>
        {#if tab==='edit'}
          <div>
            <h3 style="font-size:1.25rem;font-weight:600;color:var(--text)">{resume.name||t.editor.defaultName}</h3>
            <p style="font-size:0.875rem;color:var(--text-secondary);margin-top:0.25rem">{resume.email} | {resume.phone}</p>
            <p style="margin-top:1rem;line-height:1.7;color:var(--text-secondary)">{resume.summary||t.editor.defaultSummary}</p>
          </div>
        {:else if result}
          <div style="display:flex;flex-direction:column;gap:1.25rem">
            <div style="display:flex;justify-content:space-between;align-items:center;padding:1rem;border-radius:var(--radius);background:rgba(16,185,129,0.08);border:1px solid rgba(16,185,129,0.2)">
              <span style="font-weight:500;color:#059669">{t.editor.atsScore}</span>
              <span style="font-size:1.5rem;font-weight:700;color:#059669">{result.ats_score}%</span>
            </div>
            <div>
              <h4 style="font-weight:500;margin-bottom:0.5rem;color:var(--text)">{t.editor.keywords}</h4>
              <div style="display:flex;flex-wrap:wrap;gap:0.5rem">
                {#each result.keywords||[] as kw}
                  <span style="padding:0.25rem 0.75rem;border-radius:9999px;background:rgba(37,99,235,0.08);color:var(--primary);font-size:0.8125rem;border:1px solid rgba(37,99,235,0.15)">{kw}</span>
                {/each}
              </div>
            </div>
            <div>
              <h4 style="font-weight:500;margin-bottom:0.5rem;color:var(--text)">{t.editor.suggestions}</h4>
              <ul style="list-style:none;display:flex;flex-direction:column;gap:0.5rem">
                {#each result.suggestions||[] as s}
                  <li style="font-size:0.875rem;color:var(--text-secondary);display:flex;gap:0.5rem"><span style="color:var(--primary)">•</span>{s}</li>
                {/each}
              </ul>
            </div>
          </div>
        {:else}
          <div style="text-align:center;padding:4rem 0;color:var(--text-secondary)">
            <div style="font-size:2.5rem;margin-bottom:0.75rem;opacity:0.5">✨</div>
            <p style="font-size:0.875rem">{t.editor.emptyResult}</p>
          </div>
        {/if}
      </div>
      <button class="btn btn-primary" style="width:100%;padding:0.875rem;margin-top:1rem;font-size:1rem">{t.editor.exportPdf}</button>
    </div>
  </div>
</div>
