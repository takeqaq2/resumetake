<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';
  import { apiFetch, getToken } from '$lib/api.js';
  import { onMount } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  let resumeText = $state('');
  let targetJob = $state('');
  let jobDesc = $state('');
  let jobUrl = $state('');
  let isOptimizing = $state(false);
  let isUploading = $state(false);
  let isFetching = $state(false);
  let result = $state(null);
  let error = $state('');
  let startTime = $state(0);
  let dragOver = $state(false);
  let fileInput = $state();

  let showPerspective = $state(false);
  let perspectiveLoading = $state(false);
  let perspectiveResult = $state(null);
  let activePerspective = $state('original');
  let perspectiveTabRefs = $state([]);
  let perspectiveError = $state('');

  let modules = $state({
    ats: true, star: true, quant: true, summary: true, format: true
  });

  let usageCount = $state(0);
  let maxFreeUsage = $state(5);
  let copied = $state(false);
  let dirty = $state(false);
  let copiedTimer;
  let elapsed = $state(0);
  // Fix 3: track mount state so finally blocks in async functions can skip
  // writing to $state after the component is unmounted.
  let mounted = false;

  // AbortControllers for long-running AI requests — aborted on unmount so
  // pending 120s calls don't write to unmounted component state.
  let optimizeAbort = null;
  let perspectiveAbort = null;
  // R50-F2: AbortController for uploadFile/fetchJobUrl — these can take
  // several seconds (PDF upload, job URL scraping) and should be cancelled
  // when the user navigates away, same as optimize/perspective.
  let uploadAbort = null;
  let jobFetchAbort = null;

  const allModules = ['ats', 'star', 'quant', 'summary', 'format'];

  let allSelected = $derived(allModules.every(m => modules[m]));
  // Precompute the slice once per t change — using .slice() inline in {#each}
  // creates a new array reference on every reactive update, forcing Svelte
  // to re-diff the whole each block unnecessarily.
  let topProducts = $derived(t.pricing.products.slice(0, 3));

  function fmtUsed(used, max) {
    return t.editor.usedTodayPattern.replace('{used}', used).replace('{max}', max);
  }
  function fmtRemaining(remaining, max) {
    return t.editor.remainingPattern.replace('{remaining}', remaining).replace('{max}', max);
  }

  function toggleAll() {
    const val = !allSelected;
    allModules.forEach(m => modules[m] = val);
  }

  function copyOptimized() {
    if (!result?.optimized_content) return;
    const c = result.optimized_content;
    let text = '';
    if (c.summary) text += `${t.editor.summary}:\n${c.summary}\n\n`;
    if (c.experience?.length) {
      text += `${t.editor.experience}:\n`;
      c.experience.forEach(e => {
        // Match the fallback and separator used by the on-screen render:
        // "{position} — {company}" (em dash, not @), duration on its own line.
        const pos = e.position || e.title || '';
        const comp = e.company || e.org || '';
        text += `  ${pos} — ${comp}\n`;
        if (e.duration) text += `  ${e.duration}\n`;
        e.highlights?.forEach(h => text += `    - ${h}\n`);
        text += '\n';
      });
    }
    if (c.skills?.length) text += `${t.editor.skills}: ${c.skills.join(', ')}\n\n`;
    if (c.education?.length) {
      text += `${t.editor.education}:\n`;
      c.education.forEach(e => {
        // Match render: "{degree}, {major}" on line 1, "{school}" on line 2.
        const deg = e.degree || e.title || '';
        const sch = e.school || e.org || e.institution || '';
        text += `  ${deg}${e.major ? `, ${e.major}` : ''}\n`;
        text += `  ${sch}\n`;
      });
    }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(() => {
        copied = true;
        clearTimeout(copiedTimer);
        copiedTimer = setTimeout(() => copied = false, 2000);
      }).catch(() => {
        fallbackCopy(text);
      });
    } else {
      fallbackCopy(text);
    }
  }

  function fallbackCopy(text) {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);
    textarea.select();
    try { document.execCommand('copy'); copied = true; clearTimeout(copiedTimer); copiedTimer = setTimeout(() => copied = false, 2000); } catch {}
    document.body.removeChild(textarea);
  }

  onMount(() => {
    let cancelled = false;
    mounted = true;
    // R46-F4: removed the ?template= query param consumer — it only wrote to
    // localStorage ('selected_template') which was never read anywhere in the
    // codebase (R43-F4 removed the only reader, a removeItem call). The
    // templates page links still use ?template= but it is inert; keeping the
    // consumer gave the false impression that template selection persisted.
    const token = getToken();
    if (token) {
      apiFetch('/api/v1/auth/me', { skipAuth: true })
        .then(r => r.json()).then(d => {
          if (cancelled) return;
          if (d.success && d.data) {
            usageCount = d.data.usage_count || 0;
            maxFreeUsage = d.data.max_free_usage || 5;
          }
        }).catch(() => {});
    }
    let generated = null;
    try { generated = localStorage.getItem('generated_resume'); } catch (e) {}
    if (generated) {
      try {
        const r = JSON.parse(generated);
        let text = '';
        if (r.name) text += r.name + '\n';
        if (r.title) text += r.title + '\n';
        if (r.contact?.email) text += r.contact.email + '\n';
        if (r.summary) text += '\n' + r.summary + '\n';
        if (r.experience?.length) {
          r.experience.forEach(e => {
            text += `\n${e.position || e.title || ''} @ ${e.company || e.org || ''} (${e.duration || ''})\n`;
            (e.highlights || []).forEach(h => text += `- ${h}\n`);
          });
        }
        if (r.skills?.length) text += `\n${t.editor.skills}: ${r.skills.join(', ')}\n`;
        resumeText = text;
        try { localStorage.removeItem('generated_resume'); } catch (e) {}
      } catch {}
    }
    // Restore draft from previous session — prevents data loss on accidental
    // refresh/navigate-away. Only restore if no generated_resume was consumed.
    if (!resumeText) {
      try {
        const draft = localStorage.getItem('editor_draft');
        if (draft) {
          const d = JSON.parse(draft);
          resumeText = d.resumeText || '';
          targetJob = d.targetJob || '';
          jobDesc = d.jobDesc || '';
          jobUrl = d.jobUrl || '';
        }
      } catch (e) {}
    }
    // Restore optimization result from previous session (sessionStorage —
    // cleared when tab closes, unlike localStorage draft). Prevents losing
    // an AI result that cost a usage slot if the user accidentally refreshes.
    try {
      const savedResult = sessionStorage.getItem('editor_result');
      if (savedResult) result = JSON.parse(savedResult);
    } catch (e) {}
    // Warn before leaving only if the user has actually edited a field.
    // Without the dirty flag, restoring drafts from localStorage in onMount
    // would trigger the leave warning even when the user hasn't typed anything.
    function beforeUnloadHandler(e) {
      if (dirty) {
        // R48-F3: synchronously save the draft before the unload prompt —
        // the debounced $effect auto-save has a 1s delay, so input within
        // the last second would be lost if the user confirms leaving.
        // localStorage.setItem is synchronous, so this runs before unload.
        try {
          localStorage.setItem('editor_draft', JSON.stringify({ resumeText, targetJob, jobDesc, jobUrl }));
        } catch (err) {}
        e.preventDefault();
        e.returnValue = '';
      }
    }
    window.addEventListener('beforeunload', beforeUnloadHandler);
    return () => {
      cancelled = true;
      mounted = false;
      clearTimeout(copiedTimer);
      if (optimizeAbort) optimizeAbort.abort();
      if (perspectiveAbort) perspectiveAbort.abort();
      if (uploadAbort) uploadAbort.abort();
      if (jobFetchAbort) jobFetchAbort.abort();
      window.removeEventListener('beforeunload', beforeUnloadHandler);
    };
  });

  // Auto-save draft with debounce — protects against accidental page close.
  let draftTimer;
  $effect(() => {
    resumeText; targetJob; jobDesc; jobUrl;
    clearTimeout(draftTimer);
    draftTimer = setTimeout(() => {
      try {
        localStorage.setItem('editor_draft', JSON.stringify({ resumeText, targetJob, jobDesc, jobUrl }));
      } catch (e) {}
    }, 1000);
    return () => clearTimeout(draftTimer);
  });

  function scoreColor(score) {
    const s = score || 0;
    if (s < 50) return 'var(--error-text)';
    if (s < 75) return 'var(--warning-text)'; // amber-700, 4.8:1 contrast on light bg (WCAG AA)
    return 'var(--success-text)';
  }

  async function uploadFile(file) {
    if (!file) return;
    // R53-F1: guard against concurrent uploads — without this, a second
    // upload overwrites uploadAbort, leaving the first request unabortable,
    // and whichever finishes last wins the resumeText write (data race).
    // R53-F2: also block uploads while optimizing — the in-flight optimize
    // request carries the old resume text, so updating resumeText mid-flight
    // causes the UI to show new text while the result is for the old text.
    // R48-F1: also block while perspective analysis is running — same race
    // as optimize: the in-flight perspective request uses the old resumeText,
    // so updating it mid-flight causes UI to show new text + stale analysis.
    if (isUploading || isOptimizing || perspectiveLoading) return;
    const allowed = ['.txt', '.md', '.pdf'];
    const ext = '.' + file.name.split('.').pop().toLowerCase();
    if (!allowed.includes(ext)) {
      error = t.editor.uploadError;
      return;
    }
    // R56b-F1: also validate MIME type. Extension alone is spoofable
    // (rename malware.exe to malware.txt); while the backend re-validates
    // content, failing early on the client saves a round-trip and makes
    // the accepted types explicit. Browsers sometimes report a generic
    // or empty MIME for .md files (text/plain instead of text/markdown),
    // so we only REJECT when file.type is non-empty AND not in the
    // allowed set — an empty MIME falls through to the extension check.
    const allowedMime = ['text/plain', 'text/markdown', 'application/pdf'];
    if (file.type && !allowedMime.includes(file.type)) {
      error = t.editor.uploadError;
      return;
    }
    // PDF allows up to 2MB (backend MaxPDFBytes); text/md capped at 1MB.
    const maxSize = ext === '.pdf' ? 2 * 1024 * 1024 : 1024 * 1024;
    if (file.size > maxSize) {
      // i18n fileTooLarge already includes the per-type limits
      // ("text max 1MB, PDF max 2MB") — no need to append again.
      error = t.editor.fileTooLarge;
      return;
    }
    error = '';
    isUploading = true;
    try {
      uploadAbort = new AbortController();
      const formData = new FormData();
      formData.append('file', file);
      const res = await apiFetch('/api/v1/upload', {
        method: 'POST',
        body: formData,
        signal: uploadAbort.signal,
        // R41b-L7: 2MB PDF on a slow 512Kbps uplink takes ~32s, exceeding
        // the default 30s timeout. 60s matches the file-upload scenario.
        timeout: 60000
      });
      // R46-F3: guard against non-JSON responses (e.g. 502 gateway error
      // returning HTML). Without this, res.json() throws SyntaxError and
      // the user sees a misleading "upload error" instead of a server error.
      let data;
      try {
        data = await res.json();
      } catch {
        if (res.status === 429) {
          error = t.editor.rateLimited;
        } else if (!res.ok) {
          error = t.editor.serverError;
        } else {
          error = t.editor.uploadError;
        }
        return;
      }
      if (data.success && data.data) {
        resumeText = data.data.text;
        dirty = true;
      } else {
        error = data.message || t.editor.uploadError;
      }
    } catch (e) {
      if (e?.name === 'AbortError' && uploadAbort?.signal.aborted) return;
      error = t.editor.uploadError;
    } finally {
      if (!mounted) return;
      uploadAbort = null;
      isUploading = false;
    }
  }

  async function fetchJobUrl() {
    if (!jobUrl.trim()) return;
    // R40b-M2: guard against optimize in flight — updating targetJob/jobDesc
    // while optimize is running causes the result to be based on old values
    // while the UI shows new ones (same rationale as uploadFile's guard).
    // R48-F1: also block while perspective analysis is running.
    // Fix 2: also guard against re-entrant fetchJobUrl calls (isFetching).
    if (isFetching || isOptimizing || perspectiveLoading) return;
    // Fix 6: validate URL format before sending to backend — saves a
    // round-trip and gives a clearer message than a generic fetch error.
    try {
      const u = new URL(jobUrl.trim());
      if (u.protocol !== 'http:' && u.protocol !== 'https:') {
        error = t.editor.invalidUrl || 'Please enter a valid HTTP or HTTPS URL';
        return;
      }
    } catch {
      error = t.editor.invalidUrl || 'Please enter a valid URL';
      return;
    }
    error = '';
    isFetching = true;
    try {
      jobFetchAbort = new AbortController();
      const res = await apiFetch('/api/v1/scrape-job', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: jobUrl }),
        signal: jobFetchAbort.signal
      });
      // R46-F3: guard against non-JSON responses (e.g. 502 gateway error
      // returning HTML). Without this, res.json() throws SyntaxError and
      // the user sees a misleading "fetch error" instead of a server error.
      let data;
      try {
        data = await res.json();
      } catch {
        if (res.status === 429) {
          error = t.editor.rateLimited;
        } else if (!res.ok) {
          error = t.editor.serverError;
        } else {
          error = t.editor.fetchError;
        }
        return;
      }
      if (data.success && data.data) {
        if (data.data.title && !targetJob) targetJob = data.data.title;
        if (data.data.text && !jobDesc) jobDesc = data.data.text;
        dirty = true;
      } else {
        error = data.message || t.editor.fetchError;
      }
    } catch (e) {
      if (e?.name === 'AbortError' && jobFetchAbort?.signal.aborted) return;
      error = t.editor.fetchError;
    } finally {
      if (!mounted) return;
      jobFetchAbort = null;
      isFetching = false;
    }
  }

  async function optimize() {
    // Fix 5: concurrency guard — without this, a second optimize() call
    // (e.g. double-click or Ctrl+Enter while one is in flight) overwrites
    // optimizeAbort, leaving the first request unabortable, and whichever
    // finishes last wins the result write (data race, like uploadFile R53-F1).
    if (isOptimizing || isUploading || isFetching || perspectiveLoading) return;
    if (!resumeText.trim()) {
      error = t.editor.pasteFirst;
      return;
    }
    // R46-F5: distinguish "too short" from "empty" so the user knows to
    // add more content rather than wondering why their paste "didn't count".
    if (resumeText.trim().length < 50) {
      error = t.editor.contentTooShort;
      return;
    }
    const selected = allModules.filter(m => modules[m]);
    if (selected.length === 0) {
      error = t.editor.selectModule;
      return;
    }
    const token = getToken();
    if (!token) {
      error = t.editor.loginRequired;
      return;
    }
    error = '';
    isOptimizing = true;
    startTime = Date.now();
    elapsed = 0;
    // R50-F1: only hide perspective if not currently loading. If a perspective
    // request is in-flight, hiding the UI would discard the result and the
    // user would have to re-run it (consuming another usage credit).
    if (!perspectiveLoading) showPerspective = false;
    try {
      optimizeAbort = new AbortController();
      const res = await apiFetch('/api/v1/optimize', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        timeout: 120000,
        signal: optimizeAbort.signal,
        body: JSON.stringify({
          resume_text: resumeText,
          target_job: targetJob,
          job_description: jobDesc,
          modules: selected,
          lang
        })
      });
      let data;
      try {
        data = await res.json();
      } catch {
        if (res.status === 429) {
          error = t.editor.rateLimited;
        } else if (!res.ok) {
          error = t.editor.serverError;
        } else {
          error = t.editor.networkError;
        }
        return;
      }
      if (data.success && data.data) {
        result = data.data;
        try { sessionStorage.setItem('editor_result', JSON.stringify(data.data)); } catch (e) {}
        if (data.usage_count !== undefined) usageCount = data.usage_count;
        if (data.max_free_usage !== undefined) maxFreeUsage = data.max_free_usage;
      } else {
        if (res.status === 429) {
          error = t.editor.rateLimited;
        } else if (res.status === 403 && data.error === 'LIMIT_EXCEEDED') {
          error = t.editor.limitExceeded;
        } else if (res.status === 401) {
          error = t.editor.loginRequired;
        } else {
          error = data.error || t.editor.optimizeFailed;
        }
      }
    } catch (e) {
      // R43-F2: Distinguish timeout abort from user abort. apiFetch uses an
      // internal AbortController for timeout — when it fires, the caller's
      // signal (optimizeAbort.signal) is NOT aborted. Only when the user
      // navigates away (onMount cleanup calls optimizeAbort.abort()) is the
      // caller's signal aborted. So: if our signal is aborted, the user left
      // — silently return. Otherwise it's a timeout — show an error.
      if (e?.name === 'AbortError' && optimizeAbort?.signal.aborted) return;
      error = t.editor.networkError;
    } finally {
      if (!mounted) return;
      optimizeAbort = null;
      isOptimizing = false;
    }
  }

  async function analyzePerspective() {
    // R51-F1: prevent perspective analysis during upload/fetch — the
    // resumeText may change when upload completes, causing perspective
    // results to be based on stale text (same race R48-F1 fixed for
    // uploadFile, but the reverse guard was missing).
    // Fix 4: also block while optimize is running — same stale-text race.
    if (isUploading || isFetching || isOptimizing) return;
    if (!resumeText.trim()) {
      perspectiveError = t.editor.pasteFirst;
      return;
    }
    // R46-F5: add min-length check matching optimize() — perspective
    // analysis on <50 chars produces useless results and wastes a usage
    // credit. optimize() already had this check but used the wrong message.
    if (resumeText.trim().length < 50) {
      perspectiveError = t.editor.contentTooShort;
      return;
    }
    const token = getToken();
    if (!token) {
      perspectiveError = t.editor.loginRequiredShort;
      return;
    }
    perspectiveError = '';
    perspectiveResult = null;
    activePerspective = 'original'; // R51-F2: reset tab so new results show immediately
    perspectiveLoading = true;
    showPerspective = true;
    try {
      perspectiveAbort = new AbortController();
      const res = await apiFetch('/api/v1/perspective', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        timeout: 120000,
        signal: perspectiveAbort.signal,
        body: JSON.stringify({ resume_text: resumeText, target_job: targetJob, job_description: jobDesc, lang })
      });
      let data;
      try {
        data = await res.json();
      } catch {
        if (res.status === 429) {
          perspectiveError = t.editor.rateLimited;
        } else if (!res.ok) {
          perspectiveError = t.editor.serverError;
        } else {
          perspectiveError = t.editor.networkErrorShort;
        }
        return;
      }
      if (data.success && data.data) {
        perspectiveResult = data.data;
      } else {
        if (res.status === 429) {
          perspectiveError = t.editor.rateLimited;
        } else if (res.status === 403 && data.error === 'LIMIT_EXCEEDED') {
          perspectiveError = t.editor.limitExceeded;
        } else if (res.status === 401) {
          perspectiveError = t.editor.loginRequiredShort;
        } else {
          perspectiveError = data.error || t.editor.analysisFailed;
        }
      }
    } catch (e) {
      // R43-F2: Same timeout-vs-user-abort distinction as optimize().
      if (e?.name === 'AbortError' && perspectiveAbort?.signal.aborted) return;
      perspectiveError = t.editor.networkErrorShort;
    } finally {
      if (!mounted) return;
      perspectiveAbort = null;
      perspectiveLoading = false;
    }
  }

  $effect(() => {
    if (isOptimizing && startTime) {
      const iv = setInterval(() => { elapsed = parseFloat(((Date.now() - startTime) / 1000).toFixed(1)); }, 250);
      return () => clearInterval(iv);
    }
  });

  function handleUploadZoneClick(e) {
    // R41b-L5: also check isOptimizing — handleDrop checks it, but this
    // click handler didn't, letting users open the file picker only to
    // have uploadFile silently reject the selection.
    // R51-F5: also check perspectiveLoading for consistency with uploadFile.
    if (isUploading || isOptimizing || perspectiveLoading) return;
    if (fileInput && e.target !== fileInput) fileInput.click();
  }
  function handleDragOver(e) { e.preventDefault(); dragOver = true; }
  function handleDragLeave(e) { if (!e.currentTarget.contains(e.relatedTarget)) dragOver = false; }
  function handleDrop(e) { e.preventDefault(); dragOver = false; if (isUploading || isOptimizing || perspectiveLoading) return; if (e.dataTransfer.files.length > 0) uploadFile(e.dataTransfer.files[0]); }
  function handleFileChange(e) { if (e.target.files.length > 0) uploadFile(e.target.files[0]); e.target.value = ''; }
  function handleJobUrlKeydown(e) {
    // R38b-H1: ignore Ctrl/Meta+Enter — that combo triggers optimize()
    // via the window keydown handler. Without this guard, both fetchJobUrl()
    // and optimize() fire simultaneously when the user presses Ctrl+Enter
    // while focused in the job URL input.
    if (e.key === 'Enter' && !e.ctrlKey && !e.metaKey && !e.isComposing && !isFetching) fetchJobUrl();
  }
</script>

<svelte:window onkeydown={(e) => {
  // R50-F1: also check perspectiveLoading to avoid interrupting an
  // in-flight perspective analysis (which would waste a usage credit).
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter' && !e.isComposing && !isOptimizing && !isUploading && !perspectiveLoading && !isFetching && resumeText.trim()) {
    e.preventDefault();
    optimize();
  }
}} />

<svelte:head>
  <title>{t.meta.editorTitle}</title>
  <meta name="description" content={t.meta.editorDesc}>
  <meta name="keywords" content={t.meta.editorKeywords}>
  <link rel="canonical" href="https://resume.takee.top/{lang}/editor">
  <meta property="og:title" content={t.meta.editorTitle}>
  <meta property="og:description" content={t.meta.editorDesc}>
  <meta property="og:url" content="https://resume.takee.top/{lang}/editor">
  <meta property="og:type" content="website">
</svelte:head>

<div class="editor-header">
  <div class="orb orb-blue animate-float" aria-hidden="true" style="width:200px;height:200px;top:-20%;left:10%"></div>
  <div class="orb orb-purple animate-float" aria-hidden="true" style="width:160px;height:160px;bottom:-10%;right:15%;animation-delay:2s"></div>
  <div class="container" style="position:relative">
    <h1 class="anim-hero anim-hero-1" style="font-size:clamp(1.5rem,3vw,2rem);font-weight:700;margin-bottom:0.5rem;color:var(--text)">{t.editor.title}</h1>
    <p class="anim-hero anim-hero-2" style="color:var(--text-secondary);font-size:0.9375rem">{t.editor.subtitle}</p>
  </div>
</div>

<div class="container" style="padding:2rem 1.5rem;margin-top:-1rem">
  <div class="editor-grid">
    <!-- Left: Input -->
    <div class="editor-left">
      <!-- Resume Text / Upload -->
      <div class="editor-card anim-hero anim-hero-3">
        <label for="resume-text" class="label" style="font-weight:600;color:var(--text);font-size:0.9375rem;margin-bottom:0.75rem;display:block"><span aria-hidden="true">📋</span> {t.editor.pasteResume}</label>

        <div id="upload-zone" class="upload-zone {dragOver ? 'drag-over' : ''}" role="button" tabindex={isUploading ? -1 : 0} aria-label={t.editor.uploadFile} aria-busy={isUploading} aria-disabled={isUploading || isOptimizing || perspectiveLoading}
          onclick={handleUploadZoneClick}
          onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); handleUploadZoneClick(e); } }}
          ondragover={handleDragOver}
          ondragleave={handleDragLeave}
          ondrop={handleDrop}>
          <input bind:this={fileInput} type="file" accept=".txt,.md,.pdf,text/plain,text/markdown,application/pdf" style="display:none" onchange={handleFileChange}>
          {#if isUploading}
            <div class="upload-spinner"></div>
            <span class="upload-text">{t.editor.uploading}</span>
          {:else}
            <div class="upload-icon" aria-hidden="true">📁</div>
            <span class="upload-text">{t.editor.dragDrop}</span>
            <span class="upload-hint">{t.editor.uploadHint}</span>
          {/if}
        </div>

        <div class="divider-or"><span>{t.editor.orUploadResume}</span></div>

        <textarea
          id="resume-text"
          class="input resume-textarea"
          rows="10"
          required
          minlength="50"
          maxlength="15000"
          placeholder={t.editor.pasteResumePlaceholder}
          bind:value={resumeText}
          oninput={() => dirty = true}
          style="resize:vertical;min-height:180px;font-size:0.875rem;line-height:1.6"
        ></textarea>
        <p style="font-size:0.75rem;color:var(--text-secondary);margin-top:0.5rem"><span aria-hidden="true">💡</span> {t.editor.pasteResumeHint}</p>
      </div>

      <!-- Target Job + URL Fetch -->
      <div class="editor-card anim-hero anim-hero-4">
        <label for="target-job" class="label" style="font-weight:600;color:var(--text);font-size:0.9375rem;margin-bottom:0.75rem;display:block"><span aria-hidden="true">🎯</span> {t.editor.targetJob}</label>
        <input id="target-job" class="input" placeholder={t.editor.targetJobPlaceholder} bind:value={targetJob} oninput={() => dirty = true}>

        <div class="url-fetch-row" style="margin-top:0.75rem">
          <input
            id="job-url-input"
            class="input url-input"
            placeholder={t.editor.jobUrlPlaceholder}
            aria-label={t.editor.jobUrlAria}
            bind:value={jobUrl}
            onkeydown={handleJobUrlKeydown}
          >
          <button id="fetch-job-btn" class="fetch-btn" disabled={isFetching || isOptimizing || !jobUrl.trim()} onclick={fetchJobUrl}>
            {#if isFetching}
              <span class="fetch-spinner"></span>
            {:else}
              <span aria-hidden="true">🔗</span>
            {/if}
            <span>{isFetching ? t.editor.fetching : t.editor.fetchJobUrl}</span>
          </button>
        </div>

        <div class="divider-or"><span>{t.editor.orPasteUrl}</span></div>

        <div>
          <label for="job-desc" class="label" style="font-size:0.8125rem">{t.editor.jobDesc}</label>
          <textarea id="job-desc" class="input" rows="3" placeholder={t.editor.jobDescPlaceholder} bind:value={jobDesc} oninput={() => dirty = true} style="resize:vertical;font-size:0.8125rem"></textarea>
        </div>
      </div>

      <!-- Optimization Modules -->
      <div class="editor-card anim-hero anim-hero-5">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:0.75rem">
          <span id="opt-modules-label" style="font-weight:600;color:var(--text);font-size:0.9375rem"><span aria-hidden="true">⚙️</span> {t.editor.optModules}</span>
          <button id="toggle-all-btn" class="btn-link" onclick={toggleAll} aria-pressed={allSelected}>{allSelected ? t.editor.deselectAll : t.editor.selectAll}</button>
        </div>
        <p style="font-size:0.75rem;color:var(--text-secondary);margin-bottom:0.75rem">{t.editor.optModulesHint}</p>
        <div role="group" aria-labelledby="opt-modules-label" style="display:flex;flex-direction:column;gap:0.5rem">
          {#each allModules as m (m)}
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
        <div class="error-msg" role="alert"><span aria-hidden="true">⚠️</span> {error}</div>
      {/if}

      {#if !isOptimizing && !result && usageCount > 0}
        <div style="font-size:0.8125rem;color:var(--text-secondary);text-align:center;padding:0.5rem;background:var(--bg-surface);border-radius:var(--radius)">
          {fmtUsed(usageCount, maxFreeUsage)}
        </div>
      {/if}

      {#if isOptimizing}
        <div class="optimizing-card" role="status" aria-live="polite">
          <div class="optimizing-spinner" aria-hidden="true"></div>
          <span>{t.editor.optimizing}</span>
          <!-- R58-F-L1: aria-hidden so SR doesn't announce elapsed every
               250ms update — the "Optimizing..." text alone is sufficient
               context; the timer is visual-only. -->
          <span style="color:var(--text-secondary);font-size:0.8125rem" aria-hidden="true">{elapsed}s</span>
        </div>
      {:else}
        <button id="optimize-btn" class="optimize-btn" onclick={optimize} disabled={perspectiveLoading || isUploading || isFetching}>
          <span style="position:relative;z-index:1;display:flex;align-items:center;gap:0.5rem">{t.editor.optimizeBtn}</span>
        </button>
      {/if}

      {#if result}
        <div class="success-msg" role="status">
          <span aria-hidden="true">✅</span> {t.editor.optimized}
          {#if elapsed > 0}<span style="color:var(--text-secondary);font-size:0.8125rem;margin-inline-start:0.5rem">{t.editor.optimizedTime}: {elapsed}s</span>{/if}
          {#if maxFreeUsage > 0}
            <span style="color:var(--text-secondary);font-size:0.75rem;margin-inline-start:auto">{fmtRemaining(Math.max(0, maxFreeUsage - usageCount), maxFreeUsage)}</span>
          {/if}
        </div>
      {/if}

      <button class="perspective-btn" onclick={analyzePerspective} disabled={perspectiveLoading || isOptimizing || isUploading || isFetching || !resumeText.trim()}>
        {#if perspectiveLoading}
          <span class="perspective-spinner"></span>
        {:else}
          <span aria-hidden="true">🔍</span>
        {/if}
        <span>{t.perspective.title}</span>
      </button>

      <!-- AI Tools (Single-use products) -->
      <div class="ai-tools-section">
        <h4 class="ai-tools-title">
          <span aria-hidden="true">🛠️</span> {t.editor.aiTools}
        </h4>
        <div class="ai-tools-grid">
          {#each topProducts as product (product.name)}
            <a class="ai-tool-btn" href={`/${lang}/pricing`} style="text-decoration:none;color:inherit">
              <span class="tool-icon" aria-hidden="true">{product.icon}</span>
              <div class="tool-info">
                <span class="tool-name">{product.name}</span>
                <span class="tool-price">{product.price}</span>
                <span class="tool-desc" style="font-size:0.6875rem;color:var(--text-secondary);line-height:1.3">{product.desc}</span>
              </div>
            </a>
          {/each}
        </div>
      </div>
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
                <span style="font-weight:600;color:var(--success-text)">{t.editor.atsScore}</span>
                <p style="font-size:0.8125rem;color:{scoreColor(result.ats_score)};margin-top:0.125rem">{t.editor.aiAnalysisResult}</p>
              </div>
              <span style="font-size:2rem;font-weight:800;color:{scoreColor(result.ats_score)}">{result.ats_score || 0}%</span>
            </div>

            {#if result.keywords?.length}
              <div>
                <h4 style="font-weight:500;margin-bottom:0.75rem;color:var(--text);display:flex;align-items:center;gap:0.375rem">
                  <span aria-hidden="true">🔑</span> {t.editor.keywords}
                </h4>
                <div style="display:flex;flex-wrap:wrap;gap:0.5rem">
                  {#each result.keywords as kw, i (i + '-' + kw)}
                    <span class="keyword-tag">{kw}</span>
                  {/each}
                </div>
              </div>
            {/if}

            {#if result.suggestions?.length}
              <div>
                <h4 style="font-weight:500;margin-bottom:0.75rem;color:var(--text);display:flex;align-items:center;gap:0.375rem">
                  <span aria-hidden="true">💡</span> {t.editor.suggestions}
                </h4>
                <ul style="list-style:none;display:flex;flex-direction:column;gap:0.625rem">
                  {#each result.suggestions as s, i (i + '-' + s)}
                    <li style="font-size:0.9375rem;color:var(--text-secondary);display:flex;gap:0.625rem;align-items:flex-start;line-height:1.5">
                      <span style="color:var(--primary);margin-top:0.125rem;flex-shrink:0" aria-hidden="true">→</span>
                      <span>{s}</span>
                    </li>
                  {/each}
                </ul>
              </div>
            {/if}

            {#if result.optimized_content}
              <div>
                <h4 style="font-weight:500;margin-bottom:0.75rem;color:var(--text);display:flex;align-items:center;gap:0.375rem">
                  <span aria-hidden="true">📄</span> {t.editor.optimizedContent}
                </h4>
                <div class="optimized-content">
                  {#if result.optimized_content.summary}
                    <div style="margin-bottom:1rem">
                      <h5 style="font-size:0.8125rem;font-weight:600;color:var(--primary);text-transform:uppercase;letter-spacing:0.05em;margin-bottom:0.375rem">{t.editor.summary}</h5>
                      <p style="font-size:0.875rem;line-height:1.6;color:var(--text)">{result.optimized_content.summary}</p>
                    </div>
                  {/if}
                  {#if result.optimized_content.experience?.length}
                    <div style="margin-bottom:1rem">
                      <h5 style="font-size:0.8125rem;font-weight:600;color:var(--primary);text-transform:uppercase;letter-spacing:0.05em;margin-bottom:0.375rem">{t.editor.experience}</h5>
                      {#each result.optimized_content.experience as exp, i (i + '-' + (exp.company || exp.org || ''))}
                        <div style="margin-bottom:0.75rem;padding:0.75rem;background:var(--bg-surface);border-radius:var(--radius);border:1px solid var(--border)">
                          <p style="font-weight:600;font-size:0.875rem;color:var(--text)">{[exp.position || exp.title, exp.company || exp.org].filter(Boolean).join(' — ')}</p>
                          {#if exp.duration}<p style="font-size:0.75rem;color:var(--text-secondary);margin-top:0.125rem">{exp.duration}</p>{/if}
                          {#if exp.highlights?.length}
                            <ul style="margin-top:0.5rem;padding-inline-start:1.25rem">
                              {#each exp.highlights as h, hi (hi + '-' + h)}
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
                      <h5 style="font-size:0.8125rem;font-weight:600;color:var(--primary);text-transform:uppercase;letter-spacing:0.05em;margin-bottom:0.375rem">{t.editor.skills}</h5>
                      <div style="display:flex;flex-wrap:wrap;gap:0.375rem">
                        {#each result.optimized_content.skills as skill, si (si + '-' + skill)}
                          <span class="keyword-tag">{skill}</span>
                        {/each}
                      </div>
                    </div>
                  {/if}
                  {#if result.optimized_content.education?.length}
                    <div>
                      <h5 style="font-size:0.8125rem;font-weight:600;color:var(--primary);text-transform:uppercase;letter-spacing:0.05em;margin-bottom:0.375rem">{t.editor.education}</h5>
                      {#each result.optimized_content.education as edu, i (i + '-' + (edu.school || edu.org || ''))}
                        <div style="padding:0.5rem 0;font-size:0.875rem;color:var(--text)">
                          <p style="font-weight:600">{[edu.degree || edu.title, edu.major].filter(Boolean).join(', ')}</p>
                          <p style="color:var(--text-secondary);font-size:0.8125rem">{edu.school || edu.org || edu.institution}</p>
                        </div>
                      {/each}
                    </div>
                  {/if}
                </div>
              </div>
            {/if}

            {#if result.optimized_content}
              <button onclick={copyOptimized} style="width:100%;padding:0.75rem;border-radius:var(--radius);border:1px solid var(--primary);background:transparent;color:var(--primary);font-weight:600;cursor:pointer;font-size:0.875rem;transition:all 0.2s">
                {copied ? t.editor.copySuccess : t.editor.copyBtn}
              </button>
            {/if}
          </div>
        </div>
      {:else}
        <div class="editor-card empty-state">
          <div style="text-align:center;padding:3rem 1rem;color:var(--text-secondary)">
            <div style="font-size:3.5rem;margin-bottom:1rem;opacity:0.3" aria-hidden="true">✨</div>
            <p style="font-size:0.9375rem;font-weight:500;margin-bottom:0.375rem">{t.editor.emptyResult}</p>
            <p style="font-size:0.8125rem;color:var(--text-secondary)">{t.editor.pasteResumeHint}</p>
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>

{#if showPerspective}
  <div class="container" style="padding:2rem 1.5rem">
    <div class="perspective-section">
      <h2 style="font-size:1.25rem;font-weight:700;margin-bottom:1.25rem;color:var(--text)">{t.perspective.title}</h2>

      {#if perspectiveError}
        <div class="error-msg" role="alert"><span aria-hidden="true">⚠️</span> {perspectiveError}</div>
      {/if}

      {#if perspectiveLoading}
        <div class="perspective-loading" role="status" aria-live="polite">
          <div class="perspective-loading-spinner"></div>
          <span>{t.editor.analyzing}</span>
        </div>
      {:else if perspectiveResult}
        <div class="perspective-tabs" role="tablist" aria-label={t.editor.analysisLabel}>
          {#each ['original', 'optimized', 'imagined', 'desired'] as key, i (key)}
            <button class="tab-btn {activePerspective === key ? 'active' : ''}" role="tab" id="perspective-tab-{key}" aria-controls="perspective-panel" aria-selected={activePerspective === key} tabindex={activePerspective === key ? 0 : -1} bind:this={perspectiveTabRefs[i]} onclick={() => activePerspective = key} onkeydown={(e) => { const keys = ['original', 'optimized', 'imagined', 'desired']; const idx = keys.indexOf(key); let newIdx = -1; const rtl = document.documentElement.dir === 'rtl'; if (e.key === 'ArrowRight') { e.preventDefault(); newIdx = rtl ? (idx + 3) % 4 : (idx + 1) % 4; } else if (e.key === 'ArrowLeft') { e.preventDefault(); newIdx = rtl ? (idx + 1) % 4 : (idx + 3) % 4; } else if (e.key === 'Home') { e.preventDefault(); newIdx = 0; } else if (e.key === 'End') { e.preventDefault(); newIdx = 3; } if (newIdx >= 0) { activePerspective = keys[newIdx]; perspectiveTabRefs[newIdx]?.focus(); } }}>
              {t.perspective[key]}
            </button>
          {/each}
        </div>

        <div class="perspective-cards" role="tabpanel" id="perspective-panel" aria-labelledby="perspective-tab-{activePerspective}">
          {#if perspectiveResult[activePerspective]}
            {@const p = perspectiveResult[activePerspective]}
            {@const badgeStyles = {
              original: { bg: 'rgba(100,116,139,0.1)', color: 'var(--text-secondary)' },
              optimized: { bg: 'rgba(16,185,129,0.1)', color: 'var(--success-text)' },
              imagined: { bg: 'rgba(139,92,246,0.1)', color: 'var(--accent)' },
              desired: { bg: 'rgba(37,99,235,0.1)', color: 'var(--primary)' }
            }}
            {@const bs = badgeStyles[activePerspective] || badgeStyles.original}
            <div class="perspective-card">
              <div class="perspective-score-badge" style="background:{bs.bg};color:{bs.color}">
                {#if p.score !== undefined}
                  <span style="font-size:1.5rem;font-weight:800">{p.score}</span>
                  <span style="font-size:0.75rem;color:var(--text-secondary)">/100</span>
                {/if}
              </div>
              {#if p.summary}
                <p class="perspective-summary">{p.summary}</p>
              {/if}
              {#if p.experience_highlights?.length}
                <div class="perspective-section-inner">
                  <h4>{t.editor.experienceHighlights}</h4>
                  <ul>
                    {#each p.experience_highlights as h, hi (hi + '-' + h)}
                      <li>{h}</li>
                    {/each}
                  </ul>
                </div>
              {/if}
              {#if p.skills?.length}
                <div class="perspective-section-inner">
                  <h4>{t.editor.skillsLabel}</h4>
                  <div class="perspective-skills">
                    {#each p.skills as skill, si (si + '-' + skill)}
                      <span class="keyword-tag">{skill}</span>
                    {/each}
                  </div>
                </div>
              {/if}
              {#if p.analysis}
                <div class="perspective-section-inner">
                  <h4>{t.editor.analysisLabel}</h4>
                  <p class="perspective-analysis">{p.analysis}</p>
                </div>
              {/if}
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>
{/if}

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

  .upload-zone {
    border: 2px dashed var(--border);
    border-radius: var(--radius-lg);
    padding: 1.5rem;
    text-align: center;
    cursor: pointer;
    transition: all 0.3s;
    background: var(--bg-surface);
  }
  .upload-zone:hover, .upload-zone.drag-over {
    border-color: var(--primary);
    background: rgba(37,99,235,0.04);
  }
  .upload-icon { font-size: 2rem; margin-bottom: 0.5rem; }
  .upload-text { display: block; font-size: 0.875rem; font-weight: 500; color: var(--text); }
  .upload-hint { display: block; font-size: 0.75rem; color: var(--text-secondary); margin-top: 0.25rem; }
  .upload-spinner {
    width: 24px; height: 24px; border: 2px solid var(--border);
    border-top-color: var(--primary); border-radius: 50%;
    animation: spin 0.6s linear infinite; margin: 0 auto 0.5rem;
  }

  .divider-or {
    display: flex; align-items: center; gap: 0.75rem;
    margin: 0.75rem 0; font-size: 0.75rem; color: var(--text-secondary);
  }
  .divider-or::before, .divider-or::after {
    content: ''; flex: 1; height: 1px; background: var(--border);
  }
  /* UI3: RTL mirror for upload zone alignment (:global ancestor on <html>) */
  :global([dir="rtl"]) .upload-zone { text-align: right; }

  .url-fetch-row { display: flex; gap: 0.5rem; }
  .url-input { flex: 1; }
  .fetch-btn {
    display: flex; align-items: center; gap: 0.375rem;
    padding: 0.625rem 0.875rem; border-radius: var(--radius);
    background: var(--primary); color: white; border: none;
    font-size: 0.8125rem; font-weight: 500; cursor: pointer;
    white-space: nowrap; transition: all 0.2s;
  }
  .fetch-btn:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
  .fetch-btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .fetch-spinner {
    width: 14px; height: 14px; border: 2px solid rgba(255,255,255,0.3);
    border-top-color: white; border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }

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
  .error-msg {
    padding: 0.75rem 1rem; border-radius: var(--radius);
    background: rgba(239,68,68,0.08); border: 1px solid rgba(239,68,68,0.2);
    color: var(--error-text); font-size: 0.875rem;
    display: flex; align-items: center; gap: 0.5rem;
  }
  .success-msg {
    padding: 0.75rem 1rem; border-radius: var(--radius);
    background: rgba(16,185,129,0.08); border: 1px solid rgba(16,185,129,0.2);
    color: var(--success-text); font-size: 0.875rem; font-weight: 500;
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
    .url-fetch-row { flex-direction: column; }
    .perspective-cards { grid-template-columns: 1fr; }
  }

  .perspective-btn {
    width: 100%; padding: 0.875rem; font-size: 0.9375rem; font-weight: 600;
    background: var(--bg-glass); color: var(--text);
    border: 2px dashed var(--border); border-radius: var(--radius-lg);
    cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 0.5rem;
    transition: all 0.3s;
  }
  .perspective-btn:hover:not(:disabled) {
    border-color: var(--primary); background: rgba(37,99,235,0.04);
  }
  .perspective-btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .perspective-spinner {
    width: 16px; height: 16px; border: 2px solid var(--border);
    border-top-color: var(--primary); border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
  .perspective-section {
    background: var(--bg-glass); border: 1px solid var(--border);
    border-radius: var(--radius-lg); padding: 2rem;
    backdrop-filter: blur(16px);
  }
  .perspective-tabs {
    display: flex; gap: 0.5rem; flex-wrap: wrap; margin-bottom: 1.5rem;
    border-bottom: 1px solid var(--border); padding-bottom: 0.75rem;
  }
  .perspective-cards {
    display: grid; grid-template-columns: repeat(2, 1fr); gap: 1.25rem;
  }
  .perspective-card {
    background: var(--bg-surface); border: 1px solid var(--border);
    border-radius: var(--radius-lg); padding: 1.5rem;
    animation: fadeInUp 0.4s ease;
  }
  .perspective-score-badge {
    display: inline-flex; align-items: baseline; gap: 0.25rem;
    padding: 0.5rem 1rem; border-radius: var(--radius);
    margin-bottom: 1rem;
  }
  .perspective-summary {
    font-size: 0.9375rem; line-height: 1.6; color: var(--text);
    margin-bottom: 1rem;
  }
  .perspective-section-inner {
    margin-top: 1rem; padding-top: 1rem;
    border-top: 1px solid var(--border);
  }
  .perspective-section-inner h4 {
    font-size: 0.8125rem; font-weight: 600; color: var(--primary);
    text-transform: uppercase; letter-spacing: 0.05em;
    margin-bottom: 0.5rem;
  }
  .perspective-section-inner ul {
    list-style: none; padding: 0;
  }
  .perspective-section-inner li {
    font-size: 0.875rem; color: var(--text-secondary); line-height: 1.5;
    padding: 0.25rem 0; padding-inline-start: 1rem; position: relative;
  }
  .perspective-section-inner li::before {
    content: '→'; position: absolute; inset-inline-start: 0; color: var(--primary);
  }
  :global([dir="rtl"]) .perspective-section-inner li::before { content: '←'; }
  .perspective-skills { display: flex; flex-wrap: wrap; gap: 0.375rem; }
  .perspective-analysis {
    font-size: 0.875rem; color: var(--text-secondary); line-height: 1.6;
  }
  .perspective-loading {
    display: flex; flex-direction: column; align-items: center;
    gap: 1rem; padding: 3rem 0; color: var(--text-secondary);
  }
  .perspective-loading-spinner {
    width: 32px; height: 32px; border: 3px solid var(--border);
    border-top-color: var(--primary); border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
  @keyframes fadeInUp {
    from { opacity: 0; transform: translateY(12px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .ai-tools-section {
    margin-top: 0.5rem;
    padding: 1rem;
    background: var(--bg-glass);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    backdrop-filter: blur(8px);
  }
  .ai-tools-title {
    display: flex; align-items: center; gap: 0.5rem;
    font-size: 0.875rem; font-weight: 600; color: var(--text);
    margin-bottom: 0.75rem;
  }
  .ai-tools-grid {
    display: flex; flex-direction: column; gap: 0.5rem;
  }
  .ai-tool-btn {
    display: flex; align-items: center; gap: 0.75rem;
    padding: 0.625rem 0.75rem;
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    cursor: pointer;
    transition: all 0.2s;
    text-align: start;
  }
  .ai-tool-btn:hover {
    border-color: var(--primary);
    background: var(--bg-hover, rgba(99,102,241,0.05));
  }
  .tool-icon { font-size: 1.25rem; }
  .tool-info { display: flex; flex-direction: column; }
  .tool-name { font-size: 0.8125rem; font-weight: 500; color: var(--text); }
  .tool-price { font-size: 0.75rem; color: var(--primary); font-weight: 600; }
</style>
