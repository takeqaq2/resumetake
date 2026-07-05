<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';
  import { onMount } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  let resumeText = $state('');
  let targetJob = $state('');
  let jobDesc = $state('');
  let isOptimizing = $state(false);
  let result = $state(null);
  let error = $state('');
  let startTime = $state(0);

  let modules = $state({
    ats: true, star: true, quant: true, summary: true, format: true
  });

  const allModules = ['ats', 'star', 'quant', 'summary', 'format'];

  let allSelected = $derived(allModules.every(m => modules[m]));
  let noneSelected = $derived(allModules.every(m => !modules[m]));

  function toggleAll() {
    const val = !allSelected;
    allModules.forEach(m => modules[m] = val);
  }

  async function optimize() {
    if (!resumeText.trim()) {
      error = t.editor.pasteFirst;
      return;
    }
    const selected = allModules.filter(m => modules[m]);
    if (selected.length === 0) {
      error = lang === 'zh' ? '请至少选择一个优化模块' : lang === 'ja' ? '最適化モジュールを少なくとも1つ選択してください' : lang === 'ko' ? '최적화 모듈을 하나 이상 선택하세요' : lang === 'es' ? 'Selecciona al menos un módulo' : lang === 'pt' ? 'Selecione pelo menos um módulo' : lang === 'fr' ? 'Sélectionnez au moins un module' : lang === 'de' ? 'Wählen Sie mindestens ein Modul' : lang === 'ar' ? 'اختر وحدة تحسين واحدة على الأقل' : 'Please select at least one module';
      return;
    }
    error = '';
    isOptimizing = true;
    startTime = Date.now();
    try {
      const res = await fetch('/api/v1/optimize', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          resume_text: resumeText,
          target_job: targetJob,
          job_description: jobDesc,
          modules: selected,
          lang
        })
      });
      const data = await res.json();
      if (data.success && data.data) {
        result = data.data;
      } else {
        error = data.error || (lang === 'zh' ? '优化失败，请重试' : lang === 'ja' ? '最適化に失敗しました' : lang === 'ko' ? '최적화 실패' : 'Optimization failed');
      }
    } catch {
      error = lang === 'zh' ? '网络错误，请检查连接' : lang === 'ja' ? 'ネットワークエラー' : lang === 'ko' ? '네트워크 오류' : 'Network error';
    }
    finally { isOptimizing = false; }
  }

  let elapsed = $state(0);
  $effect(() => {
    if (isOptimizing && startTime) {
      const iv = setInterval(() => { elapsed = ((Date.now() - startTime) / 1000).toFixed(1); }, 100);
      return () => clearInterval(iv);
    }
  });
</script>

<svelte:head>
  <title>{t.meta.editorTitle}</title>
  <meta name="description" content={t.meta.editorDesc}>
  <meta name="keywords" content={t.meta.editorKeywords}>
  <link rel="canonical" href="https://resume.takee.top/{lang}/editor">
  <meta property="og:title" content={t.meta.editorTitle}>
  <meta property="og:description" content={t.meta.editorDesc}>
</svelte:head>

<div class="editor-header">
  <div class="orb orb-blue animate-float" style="width:200px;height:200px;top:-20%;left:10%"></div>
  <div class="orb orb-purple animate-float" style="width:160px;height:160px;bottom:-10%;right:15%;animation-delay:2s"></div>
  <div class="container" style="position:relative">
    <h1 class="anim-hero anim-hero-1" style="font-size:clamp(1.5rem,3vw,2rem);font-weight:700;margin-bottom:0.5rem;color:var(--text)">{t.editor.title}</h1>
    <p class="anim-hero anim-hero-2" style="color:var(--text-secondary);font-size:0.9375rem">{t.editor.subtitle}</p>
  </div>
</div>

<div class="container" style="padding:2rem 1.5rem;margin-top:-1rem">
  <div class="editor-grid">
    <!-- Left: Input -->
    <div class="editor-left">
      <!-- Resume Text -->
      <div class="editor-card anim-hero anim-hero-3">
        <label for="resume-text" class="label" style="font-weight:600;color:var(--text);font-size:0.9375rem;margin-bottom:0.75rem;display:block">📋 {t.editor.pasteResume}</label>
        <textarea
          id="resume-text"
          class="input resume-textarea"
          rows="10"
          placeholder={t.editor.pasteResumePlaceholder}
          bind:value={resumeText}
          style="resize:vertical;min-height:180px;font-size:0.875rem;line-height:1.6"
        ></textarea>
        <p style="font-size:0.75rem;color:var(--text-secondary);margin-top:0.5rem;opacity:0.7">💡 {t.editor.pasteResumeHint}</p>
      </div>

      <!-- Target Job -->
      <div class="editor-card anim-hero anim-hero-4">
        <label for="target-job" class="label" style="font-weight:600;color:var(--text);font-size:0.9375rem;margin-bottom:0.75rem;display:block">🎯 {t.editor.targetJob}</label>
        <input id="target-job" class="input" placeholder={t.editor.targetJobPlaceholder} bind:value={targetJob}>
        <div style="margin-top:0.75rem">
          <label for="job-desc" class="label" style="font-size:0.8125rem">{t.editor.jobDesc}</label>
          <textarea id="job-desc" class="input" rows="3" placeholder={t.editor.jobDescPlaceholder} bind:value={jobDesc} style="resize:vertical;font-size:0.8125rem"></textarea>
        </div>
      </div>

      <!-- Optimization Modules -->
      <div class="editor-card anim-hero anim-hero-5">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:0.75rem">
          <span id="opt-modules-label" style="font-weight:600;color:var(--text);font-size:0.9375rem">⚙️ {t.editor.optModules}</span>
          <button class="btn-link" onclick={toggleAll}>{allSelected ? t.editor.deselectAll : t.editor.selectAll}</button>
        </div>
        <p style="font-size:0.75rem;color:var(--text-secondary);margin-bottom:0.75rem">{t.editor.optModulesHint}</p>
        <div style="display:flex;flex-direction:column;gap:0.5rem">
          {#each allModules as m}
            <label class="module-item {modules[m] ? 'active' : ''}">
              <input type="checkbox" bind:checked={modules[m]} style="width:1.125rem;height:1.125rem;accent-color:var(--primary);cursor:pointer">
              <div>
                <span class="module-label">{t.editor[`module_${m}`]}</span>
                <span class="module-desc">{t.editor[`module_${m}_desc`]}</span>
              </div>
            </label>
          {/each}
        </div>
      </div>

      {#if error}
        <div class="error-msg"><span>⚠️</span> {error}</div>
      {/if}

      {#if isOptimizing}
        <div class="optimizing-card">
          <div class="optimizing-spinner"></div>
          <span>{t.editor.optimizing}</span>
          <span style="opacity:0.6;font-size:0.8125rem">{elapsed}s</span>
        </div>
      {:else}
        <button class="optimize-btn" onclick={optimize}>
          <span style="position:relative;z-index:1;display:flex;align-items:center;gap:0.5rem">{t.editor.optimizeBtn}</span>
        </button>
      {/if}

      {#if result}
        <div class="success-msg">
          <span>✅</span> {t.editor.optimized}
          {#if elapsed > 0}<span style="opacity:0.6;font-size:0.8125rem;margin-left:0.5rem">{t.editor.optimizedTime}: {elapsed}s</span>{/if}
        </div>
      {/if}
    </div>

    <!-- Right: Preview / Result -->
    <div class="editor-right">
      {#if result}
        <div class="editor-card">
          <div class="editor-tabs">
            <button class="tab-btn active">{t.editor.resultTab}</button>
          </div>

          <div style="display:flex;flex-direction:column;gap:1.25rem">
            <div class="result-score">
              <div>
                <span style="font-weight:600;color:#059669">{t.editor.atsScore}</span>
                <p style="font-size:0.8125rem;color:#059669;opacity:0.7;margin-top:0.125rem">{lang === 'zh' ? 'AI分析结果' : 'AI Analysis Result'}</p>
              </div>
              <span style="font-size:2rem;font-weight:800;color:#059669">{result.ats_score || 0}%</span>
            </div>

            {#if result.keywords?.length}
              <div>
                <h4 style="font-weight:500;margin-bottom:0.75rem;color:var(--text);display:flex;align-items:center;gap:0.375rem">
                  <span>🔑</span> {t.editor.keywords}
                </h4>
                <div style="display:flex;flex-wrap:wrap;gap:0.5rem">
                  {#each result.keywords as kw}
                    <span class="keyword-tag">{kw}</span>
                  {/each}
                </div>
              </div>
            {/if}

            {#if result.suggestions?.length}
              <div>
                <h4 style="font-weight:500;margin-bottom:0.75rem;color:var(--text);display:flex;align-items:center;gap:0.375rem">
                  <span>💡</span> {t.editor.suggestions}
                </h4>
                <ul style="list-style:none;display:flex;flex-direction:column;gap:0.625rem">
                  {#each result.suggestions as s}
                    <li style="font-size:0.9375rem;color:var(--text-secondary);display:flex;gap:0.625rem;align-items:flex-start;line-height:1.5">
                      <span style="color:var(--primary);margin-top:0.125rem;flex-shrink:0">→</span>
                      <span>{s}</span>
                    </li>
                  {/each}
                </ul>
              </div>
            {/if}

            {#if result.optimized_content}
              <div>
                <h4 style="font-weight:500;margin-bottom:0.75rem;color:var(--text);display:flex;align-items:center;gap:0.375rem">
                  <span>📄</span> {lang === 'zh' ? '优化后内容' : lang === 'ja' ? '最適化済みコンテンツ' : lang === 'ko' ? '최적화된 내용' : 'Optimized Content'}
                </h4>
                <div class="optimized-content">
                  {#if result.optimized_content.summary}
                    <div style="margin-bottom:1rem">
                      <h5 style="font-size:0.8125rem;font-weight:600;color:var(--primary);text-transform:uppercase;letter-spacing:0.05em;margin-bottom:0.375rem">{lang === 'zh' ? '个人简介' : lang === 'ja' ? '自己PR' : lang === 'ko' ? '자기소개' : 'Summary'}</h5>
                      <p style="font-size:0.875rem;line-height:1.6;color:var(--text)">{result.optimized_content.summary}</p>
                    </div>
                  {/if}
                  {#if result.optimized_content.experience?.length}
                    <div style="margin-bottom:1rem">
                      <h5 style="font-size:0.8125rem;font-weight:600;color:var(--primary);text-transform:uppercase;letter-spacing:0.05em;margin-bottom:0.375rem">{lang === 'zh' ? '工作经历' : lang === 'ja' ? '職務経歴' : lang === 'ko' ? '업무 경험' : 'Experience'}</h5>
                      {#each result.optimized_content.experience as exp}
                        <div style="margin-bottom:0.75rem;padding:0.75rem;background:var(--bg-surface);border-radius:var(--radius);border:1px solid var(--border)">
                          <p style="font-weight:600;font-size:0.875rem;color:var(--text)">{exp.position || exp.title} — {exp.company || exp.org}</p>
                          {#if exp.duration}<p style="font-size:0.75rem;color:var(--text-secondary);margin-top:0.125rem">{exp.duration}</p>{/if}
                          {#if exp.highlights?.length}
                            <ul style="margin-top:0.5rem;padding-left:1.25rem">
                              {#each exp.highlights as h}
                                <li style="font-size:0.8125rem;color:var(--text-secondary);line-height:1.5;margin-bottom:0.25rem">{h}</li>
                              {/each}
                            </ul>
                          {/if}
                        </div>
                      {/each}
                    </div>
                  {/if}
                  {#if result.optimized_content.skills?.length}
                    <div>
                      <h5 style="font-size:0.8125rem;font-weight:600;color:var(--primary);text-transform:uppercase;letter-spacing:0.05em;margin-bottom:0.375rem">{lang === 'zh' ? '技能' : lang === 'ja' ? 'スキル' : lang === 'ko' ? '기술' : 'Skills'}</h5>
                      <div style="display:flex;flex-wrap:wrap;gap:0.375rem">
                        {#each result.optimized_content.skills as skill}
                          <span class="keyword-tag">{skill}</span>
                        {/each}
                      </div>
                    </div>
                  {/if}
                </div>
              </div>
            {/if}
          </div>
        </div>
      {:else}
        <div class="editor-card empty-state">
          <div style="text-align:center;padding:3rem 1rem;color:var(--text-secondary)">
            <div style="font-size:3.5rem;margin-bottom:1rem;opacity:0.3">✨</div>
            <p style="font-size:0.9375rem;font-weight:500;margin-bottom:0.375rem">{t.editor.emptyResult}</p>
            <p style="font-size:0.8125rem;opacity:0.5">{t.editor.pasteResumeHint}</p>
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .anim-hero { animation: heroFadeIn 0.5s cubic-bezier(0.16, 1, 0.3, 1) both; }
  .anim-hero-1 { animation-delay: 0.05s; }
  .anim-hero-2 { animation-delay: 0.15s; }
  .anim-hero-3 { animation-delay: 0.2s; }
  .anim-hero-4 { animation-delay: 0.3s; }
  .anim-hero-5 { animation-delay: 0.4s; }
  @keyframes heroFadeIn {
    from { opacity: 0; transform: translateY(16px); }
    to { opacity: 1; transform: translateY(0); }
  }
  .editor-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
    align-items: start;
  }
  .editor-left { display: flex; flex-direction: column; gap: 1rem; }
  .editor-right { position: sticky; top: 5rem; }
  .resume-textarea { font-family: 'SF Mono', 'Fira Code', monospace; }
  .module-item {
    display: flex; align-items: center; gap: 0.75rem;
    padding: 0.625rem 0.75rem; border-radius: var(--radius);
    border: 1px solid var(--border); cursor: pointer;
    transition: all 0.2s;
  }
  .module-item:hover { border-color: var(--primary); background: rgba(37,99,235,0.03); }
  .module-item.active { border-color: var(--primary); background: rgba(37,99,235,0.05); }
  .module-label { display: block; font-size: 0.875rem; font-weight: 500; color: var(--text); }
  .module-desc { display: block; font-size: 0.75rem; color: var(--text-secondary); margin-top: 0.125rem; }
  .btn-link {
    background: none; border: none; color: var(--primary);
    font-size: 0.75rem; cursor: pointer; font-weight: 500;
    padding: 0.25rem 0.5rem; border-radius: var(--radius);
    transition: all 0.2s;
  }
  .btn-link:hover { background: rgba(37,99,235,0.08); }
  .optimize-btn {
    width: 100%; padding: 1rem; font-size: 1rem; font-weight: 600;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    color: white; border: none; border-radius: var(--radius-lg);
    cursor: pointer; position: relative; overflow: hidden;
    box-shadow: 0 4px 20px var(--primary-glow);
    transition: all 0.3s;
  }
  .optimize-btn:hover { transform: translateY(-2px); box-shadow: 0 6px 30px var(--primary-glow), 0 0 60px var(--accent-glow); }
  .optimizing-card {
    display: flex; align-items: center; justify-content: center; gap: 0.75rem;
    padding: 1rem; border-radius: var(--radius-lg);
    background: linear-gradient(135deg, var(--primary), var(--accent));
    color: white; font-weight: 600; font-size: 0.9375rem;
  }
  .optimizing-spinner {
    width: 18px; height: 18px; border: 2px solid rgba(255,255,255,0.3);
    border-top-color: white; border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
  .error-msg {
    padding: 0.75rem 1rem; border-radius: var(--radius);
    background: rgba(239,68,68,0.08); border: 1px solid rgba(239,68,68,0.2);
    color: #ef4444; font-size: 0.875rem;
    display: flex; align-items: center; gap: 0.5rem;
  }
  .success-msg {
    padding: 0.75rem 1rem; border-radius: var(--radius);
    background: rgba(16,185,129,0.08); border: 1px solid rgba(16,185,129,0.2);
    color: #059669; font-size: 0.875rem; font-weight: 500;
    display: flex; align-items: center; gap: 0.5rem;
  }
  .optimized-content {
    padding: 1rem; border-radius: var(--radius);
    background: var(--bg-surface); border: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .empty-state { min-height: 400px; display: flex; align-items: center; justify-content: center; }
  .editor-tabs {
    display: flex; gap: 0; border-bottom: 1px solid var(--border); margin-bottom: 1.25rem;
  }
  @media (max-width: 768px) {
    .editor-grid { grid-template-columns: 1fr !important; }
    .editor-right { position: static; }
  }
</style>
