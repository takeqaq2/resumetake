<script>
  import { page } from '$app/stores';
  import { env } from '$env/dynamic/public';
  import { getTranslation } from '$lib/i18n/index.js';
  import { apiFetch } from '$lib/api.js';
  import AdSlot from '$lib/AdSlot.svelte';
  import { onMount, tick, untrack } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  // R50-F13: initialize with the welcome message so SSR renders it
  // immediately. Previously messages was [] until onMount ran (client-only),
  // causing a first-paint flash of an empty chat on slow connections.
  // R52-F2: use a deterministic id instead of uuid() — uuid() produces
  // different values on server vs client, causing SSR hydration key mismatch.
  let messages = $state([{ id: 'welcome', role: 'ai', content: t.generate.welcomeMessage }]);
  let inputText = $state('');
  let isTyping = $state(false);
  let showViewButton = $state(false);
  let chatContainer = $state(null);
  // R47-F3: ref to the textarea so we can reset its height after sending
  // (setting inputText='' via bind:value doesn't fire oninput, so the
  // auto-resize handler never runs and the textarea stays expanded).
  let textareaEl = $state(null);
  // R52-F5: track mount state so sendMessage's finally block can skip
  // writing to $state after the component is unmounted (AbortError path).
  let mounted = false;

  // R43-F1: Three-state loading — null=fetching /api/health, true=locked,
  // false=available. Previously defaulted to true, causing a 100-500ms
  // "locked" flash for Pro users before /api/health resolved.
  let generateLocked = $state(null);
  let parseError = $state(false);
  const generateAdSlot = env.PUBLIC_AD_SLOT_GENERATE || '';

  // uuid() requires a secure context (HTTPS or localhost).
  // Provide a fallback for local dev over HTTP IP addresses.
  function uuid() {
    if (typeof crypto !== 'undefined' && crypto.randomUUID) return crypto.randomUUID();
    return 'id-' + Date.now() + '-' + Math.random().toString(36).slice(2);
  }

  // AbortController for the generate-resume request — aborted on unmount so
  // a pending 120s AI call doesn't keep running and write to unmounted state.
  let generateAbort = null;

  function scrollToBottom() {
    if (chatContainer) {
      chatContainer.scrollTop = chatContainer.scrollHeight;
    }
  }

  // R57b-F5: only auto-scroll if the user is near the bottom. If they
  // scrolled up to read history, force-scrolling would yank them back
  // down when a new AI message arrives — disruptive. 120px threshold
  // accounts for the input area height so "at the bottom" feels natural.
  function scrollToBottomIfNear() {
    if (!chatContainer) return;
    const { scrollTop, scrollHeight, clientHeight } = chatContainer;
    if (scrollHeight - scrollTop - clientHeight < 120) {
      chatContainer.scrollTop = scrollHeight;
    }
  }

  onMount(() => {
    let cancelled = false;
    mounted = true;
    // Fix 3: restore the AI conversation from sessionStorage so a refresh
    // doesn't lose the entire chat history. sessionStorage (not localStorage)
    // so it's scoped to the tab and cleared when the tab closes.
    try {
      const saved = sessionStorage.getItem('generate_messages');
      if (saved) {
        const parsed = JSON.parse(saved);
        if (Array.isArray(parsed) && parsed.length > 0) messages = parsed;
      }
    } catch (e) {}
    try {
      if (sessionStorage.getItem('generate_showViewButton') === 'true') showViewButton = true;
    } catch (e) {}
    // R50-F13: welcome message is now initialized in $state above, no
    // need to set it here. This also prevents a hydration mismatch where
    // SSR renders the welcome message but onMount overwrites it with a
    // new uuid (different id → Svelte reconciliation flicker).
    // R47-F4: removed the tick().then(scrollToBottom) call from onMount —
    // when generateLocked===null (initial state), the chat UI is hidden
    // behind the loading spinner, so chatContainer is null and scrollToBottom
    // silently does nothing. The scroll now happens via the $effect below
    // which fires after generateLocked becomes false and the chat renders.
    apiFetch('/api/health', { skipAuth: true, timeout: 5000 }).then(r => r.json()).then(d => {
      if (cancelled) return;
      generateLocked = !(d.generate_resume_enabled === true);
    }).catch(() => {
      if (cancelled) return;
      generateLocked = true;
    });
    return () => {
      cancelled = true;
      mounted = false;
      if (generateAbort) generateAbort.abort();
    };
  });

  // onMount only runs once, so the welcome message stays in the original
  // language when the user switches languages. Re-translate it when lang
  // changes, but only if the user hasn't started chatting yet (no user-role
  // messages) — we don't want to clobber an active conversation.
  // untrack() breaks the read→write dependency cycle: without it, reading
  // messages.length and writing messages in the same $effect creates an
  // infinite loop (Svelte 5 re-runs the effect because its dependency
  // changed, but the effect itself changed it).
  $effect(() => {
    lang;
    untrack(() => {
      if (messages.length === 1 && messages[0]?.role === 'ai' && messages[0].content !== t.generate.welcomeMessage) {
        messages = [{ ...messages[0], content: t.generate.welcomeMessage }];
      }
    });
  });

  // Fix 3: persist the conversation to sessionStorage whenever it changes,
  // so a refresh restores the full AI chat history instead of wiping it.
  // Only reads $state and writes to sessionStorage (not $state), so no
  // read→write dependency cycle — safe without untrack().
  $effect(() => {
    try {
      sessionStorage.setItem('generate_messages', JSON.stringify(messages));
    } catch (e) {}
  });
  $effect(() => {
    try {
      sessionStorage.setItem('generate_showViewButton', String(showViewButton));
    } catch (e) {}
  });

  // R47-F4: scroll to bottom once the chat UI actually renders. Previously
  // this was in onMount via tick().then(scrollToBottom), but generateLocked
  // starts as null (loading state) which hides the chat-container behind the
  // loading spinner — chatContainer was null and the scroll was a no-op.
  // This $effect fires after generateLocked becomes false AND the DOM
  // updates (tick) so bind:this has assigned chatContainer.
  let initialScrollDone = false;
  $effect(() => {
    if (generateLocked === false && !initialScrollDone) {
      initialScrollDone = true;
      tick().then(() => scrollToBottom());
    }
  });

  async function sendMessage() {
    if (!inputText.trim() || isTyping) return;
    const userMsg = inputText.trim();
    // R54b-F1: reset stale state from the previous AI response so the user
    // doesn't see a leftover "View Resume" button or parse-error banner
    // while the new request is in flight.
    showViewButton = false;
    parseError = false;
    messages = [...messages, { id: uuid(), role: 'user', content: userMsg }];
    inputText = '';
    // R47-F3: reset the textarea height — setting inputText='' via bind:value
    // updates the value but doesn't fire the oninput handler, so the
    // auto-resize logic never runs and the textarea stays at its expanded
    // height (e.g. after pasting a long message).
    if (textareaEl) {
      textareaEl.style.height = 'auto';
    }
    isTyping = true;
    // R54b-F2: await tick() so the new user message is in the DOM before
    // scrolling — without this, scrollHeight doesn't reflect the new message
    // and the chat doesn't scroll far enough.
    await tick();
    // R57-F5: the component could unmount during the await tick() window
    // (e.g. user navigates away). Previously generateAbort was created
    // AFTER this point, so onMount cleanup's "if (generateAbort) abort()"
    // was a no-op (null) — the subsequent apiFetch would proceed on an
    // unmounted component, wasting a 120s AI call. Bail out early.
    if (!mounted) return;
    scrollToBottom();

    try {
      // Keep only the most recent 10 messages to bound payload size and AI
      // token cost. The system prompt is server-side; only context needed.
      const recent = messages.length > 10 ? messages.slice(-10) : messages;
      generateAbort = new AbortController();
      const res = await apiFetch('/api/v1/generate-resume', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        timeout: 120000,
        signal: generateAbort.signal,
        body: JSON.stringify({
          messages: recent.map(m => ({ role: m.role === 'ai' ? 'assistant' : m.role, content: m.content })),
          lang
        })
      });
      // Guard against non-JSON responses (e.g. 502 gateway error returning HTML).
      // Without this, res.json() throws SyntaxError and the user sees a misleading
      // "network error" instead of a server error message.
      let data;
      try {
        data = await res.json();
      } catch {
        messages = [...messages, { id: uuid(), role: 'ai', content: t.generate.somethingWentWrong }];
        return;
      }
      // R57b-F4: guard against res.json() returning null (e.g. backend
      // returns literal "null"). Without this, data.success throws
      // TypeError: Cannot read properties of null, caught by the outer
      // catch and displayed as "network error" — misleading.
      if (!data || typeof data !== 'object') {
        messages = [...messages, { id: uuid(), role: 'ai', content: t.generate.somethingWentWrong }];
        return;
      }
      if (data.success && (data.data?.message || data.message)) {
        const msg = data.data?.message || data.message;
        messages = [...messages, { id: uuid(), role: 'ai', content: msg }];
        if (data.data?.resume_complete || data.resume_complete) {
          showViewButton = true;
        }
      } else {
        messages = [...messages, { id: uuid(), role: 'ai', content: data.data?.message || data.message || (t.generate.somethingWentWrong) }];
      }
    } catch (e) {
      // R46-F2: distinguish user-cancel (navigation) from timeout.
      // - User navigated away: generateAbort.signal.aborted === true (the
      //   onMount cleanup called generateAbort.abort()). Don't push a message
      //   to the unmounted component's state.
      // - apiFetch timeout: generateAbort.signal.aborted === false (only
      //   apiFetch's internal controller was aborted, not the caller's signal).
      //   Show a timeout-specific message so the user knows to retry.
      if (e?.name === 'AbortError') {
        if (generateAbort?.signal.aborted) return; // user cancelled
        messages = [...messages, { id: uuid(), role: 'ai', content: t.generate.timeoutError || t.generate.networkError }];
        return;
      }
      messages = [...messages, { id: uuid(), role: 'ai', content: t.generate.networkError }];
    } finally {
      generateAbort = null;
      // R52-F5: skip state writes if the component was unmounted while the
      // request was in flight (onMount cleanup sets mounted=false + aborts).
      if (!mounted) return;
      // Fix 9: trim message history in finally so it runs regardless of
      // success/failure, not only in the success branch.
      if (messages.length > 50) messages = messages.slice(-50);
      isTyping = false;
      await tick();
      // R57b-F5: use scrollToBottomIfNear so users who scrolled up to read
      // history aren't yanked back down when the AI response arrives.
      scrollToBottomIfNear();
    }
  }

  function handleKeydown(e) {
    // Skip when the user is composing with an IME (CJK input) — Enter
    // confirms the candidate, it shouldn't send the message.
    if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
      // R57-F6: only swallow Enter when we'll actually send. Previously
      // preventDefault ran unconditionally — when isTyping was true or
      // input was empty, sendMessage() returned early but Enter was still
      // swallowed, so the user couldn't insert a newline while the AI was
      // responding. Now we let Enter pass through (inserting a newline)
      // when sending isn't possible.
      if (!inputText.trim() || isTyping) return;
      e.preventDefault();
      sendMessage();
    }
  }

  // AI messages may wrap the resume JSON in markdown fences or surround it
  // with explanatory text. Extract the {"resume": ...} payload robustly.
  function extractResumeFromMessage(content) {
    if (!content) return null;
    // Strip markdown code fences first.
    const fenced = content.match(/```(?:json)?\s*([\s\S]*?)```/i);
    const candidates = [];
    if (fenced) candidates.push(fenced[1]);
    candidates.push(content);
    for (const text of candidates) {
      const start = text.indexOf('{');
      if (start === -1) continue;
      // Try parsing the whole tail, then progressively shrink to the last }.
      for (let end = text.lastIndexOf('}'); end > start; end = text.lastIndexOf('}', end - 1)) {
        try {
          const parsed = JSON.parse(text.slice(start, end + 1));
          if (parsed && parsed.resume) return parsed.resume;
        } catch { /* try next slice */ }
      }
    }
    return null;
  }
</script>

<svelte:head>
  <title>{t.meta.generateTitle}</title>
  <meta name="description" content={t.meta.generateDesc}>
  <link rel="canonical" href="https://resume.takee.top/{lang}/generate">
  <meta property="og:title" content={t.meta.generateTitle}>
  <meta property="og:description" content={t.meta.generateDesc}>
  <meta property="og:url" content="https://resume.takee.top/{lang}/generate">
  <meta property="og:type" content="website">
</svelte:head>

<div class="generate-page">
  <div class="generate-header">
    <div class="container" style="position:relative;display:flex;align-items:center;justify-content:space-between">
      <div>
        <h1 style="font-size:clamp(1.25rem,2.5vw,1.5rem);font-weight:700;margin-bottom:0.25rem">{t.generate.title}</h1>
        <p style="color:var(--text-secondary);font-size:0.875rem">{t.generate.subtitle}</p>
      </div>
      {#if showViewButton}
        <div style="display:flex;align-items:center;gap:0.75rem">
          {#if parseError}
            <span role="alert" style="color:var(--error-text);font-size:0.8rem">{t.generate.parseFailed}</span>
          {/if}
          <a href="/{lang}/editor" class="btn btn-primary" style="white-space:nowrap"
            onclick={(e) => {
              // R49-F4: iterate from the end to find the last AI message
              // without copying + reversing the entire array. The previous
              // [...messages].reverse() created a new array on every click,
              // which is wasteful for long chat histories.
              let lastAi = null;
              for (let i = messages.length - 1; i >= 0; i--) {
                if (messages[i].role === 'ai') { lastAi = messages[i]; break; }
              }
              const resume = extractResumeFromMessage(lastAi?.content);
              parseError = false;
              if (resume) {
                try { localStorage.setItem('generated_resume', JSON.stringify(resume)); } catch (err) {
                  e.preventDefault();
                  parseError = true;
                }
              } else {
                e.preventDefault();
                parseError = true;
              }
            }}>
            {t.generate.viewOptimizedResume} →
          </a>
        </div>
      {/if}
    </div>
  </div>

  {#if generateLocked === null}
    <div class="generate-loading" role="status" aria-live="polite">
      <span class="loading-spinner" aria-hidden="true"></span>
    </div>
  {:else if generateLocked}
    <div class="locked-panel">
      <div class="locked-card">
        <div class="locked-icon" aria-hidden="true">🔒</div>
        <h2>{t.generate.locked.title}</h2>
        <p>{t.generate.locked.desc}</p>
        <a href="/{lang}/pricing" class="btn btn-primary">{t.generate.locked.cta}</a>
      </div>
    </div>
  {:else}
  <div class="generate-ad-wrap">
    <AdSlot slot={generateAdSlot} label={t.ads.label} />
  </div>
  <div class="chat-container" bind:this={chatContainer}>
    <!-- R57b-F5: removed aria-live from the container — it announced every
         new message including the user's own typed text (redundant for SR
         users). Instead, each AI message gets role="status" so only AI
         responses are auto-announced. User messages remain in the DOM for
         SR navigation but are not live-announced. -->
    <div class="chat-messages">
      {#each messages as msg (msg.id)}
        <div class="chat-msg {msg.role === 'user' ? 'user-msg' : 'ai-msg'}" role={msg.role === 'ai' ? 'status' : undefined}>
          {#if msg.role === 'ai'}
            <div class="ai-avatar" aria-hidden="true">R</div>
          {/if}
          <div class="msg-bubble">
            {msg.content}
          </div>
          {#if msg.role === 'user'}
            <div class="user-avatar" aria-hidden="true">👤</div>
          {/if}
        </div>
      {/each}

      {#if isTyping}
        <div class="chat-msg ai-msg" role="status" aria-label={t.generate.aiTyping}>
          <div class="ai-avatar" aria-hidden="true">R</div>
          <div class="msg-bubble typing-bubble" aria-hidden="true">
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
          </div>
        </div>
      {/if}
    </div>
  </div>

  <div class="chat-input-area">
    <div class="chat-input-wrap">
      <textarea
        class="chat-input"
        placeholder={t.generate.placeholder}
        aria-label={t.generate.chatAria}
        bind:value={inputText}
        bind:this={textareaEl}
        onkeydown={handleKeydown}
        oninput={(e) => { e.currentTarget.style.height = 'auto'; e.currentTarget.style.height = Math.min(e.currentTarget.scrollHeight, 128) + 'px'; }}
        rows="1"
        maxlength="5000"
      ></textarea>
      <button class="send-btn" onclick={sendMessage} disabled={!inputText.trim() || isTyping} aria-label={t.generate.send}>
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </button>
    </div>
  </div>
  {/if}
</div>

<style>
  .generate-page {
    display: flex; flex-direction: column;
    height: calc(100vh - 4rem); /* fallback */
    height: calc(100dvh - 4rem); /* dynamic viewport height — mobile toolbar collapse no longer causes jank */
  }
  .generate-header {
    padding: 1rem 0; border-bottom: 1px solid var(--border);
    background: var(--bg-glass); backdrop-filter: blur(12px);
  }
  .generate-loading {
    flex: 1; display: grid; place-items: center; padding: 3rem;
  }
  .loading-spinner {
    width: 32px; height: 32px;
    border: 3px solid var(--border);
    border-top-color: var(--primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
  .locked-panel {
    flex: 1; display: grid; place-items: center; padding: 2rem 1.5rem;
    background: var(--gradient-hero); background-size: 200% 200%;
  }
  .locked-card {
    width: min(100%, 34rem); padding: 2rem; border-radius: var(--radius-lg);
    background: var(--bg-glass); border: 1px solid var(--border); text-align: center;
    box-shadow: var(--shadow-lg); backdrop-filter: blur(16px);
  }
  .locked-icon { font-size: 2.5rem; margin-bottom: 1rem; }
  .locked-card h2 { font-size: 1.35rem; margin-bottom: 0.75rem; color: var(--text); }
  .locked-card p { color: var(--text-secondary); line-height: 1.7; margin-bottom: 1.5rem; }
  .generate-ad-wrap {
    max-width: 48rem; width: 100%; margin: 0 auto; padding: 0 1.5rem;
  }
  .chat-container {
    flex: 1; overflow-y: auto; padding: 1.5rem;
  }
  .chat-messages {
    max-width: 48rem; margin: 0 auto;
    display: flex; flex-direction: column; gap: 1.25rem;
  }
  .chat-msg {
    display: flex; align-items: flex-end; gap: 0.75rem;
    animation: fadeInUp 0.3s ease;
  }
  .user-msg { justify-content: flex-end; }
  .ai-msg { justify-content: flex-start; }
  .ai-avatar {
    width: 2rem; height: 2rem; border-radius: 50%;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    display: flex; align-items: center; justify-content: center;
    color: white; font-weight: 700; font-size: 0.75rem;
    flex-shrink: 0;
  }
  .user-avatar {
    width: 2rem; height: 2rem; border-radius: 50%;
    background: var(--bg-surface); border: 1px solid var(--border);
    display: flex; align-items: center; justify-content: center;
    font-size: 0.875rem; flex-shrink: 0;
  }
  .msg-bubble {
    padding: 0.75rem 1rem; border-radius: var(--radius-lg);
    font-size: 0.9375rem; line-height: 1.6; max-width: 75%;
  }
  .ai-msg .msg-bubble {
    background: var(--bg-glass); border: 1px solid var(--border);
    color: var(--text); border-bottom-left-radius: 0.25rem;
  }
  .user-msg .msg-bubble {
    background: linear-gradient(135deg, var(--primary), var(--accent));
    color: white; border-bottom-right-radius: 0.25rem;
  }
  .typing-bubble {
    display: flex; align-items: center; gap: 0.25rem;
    padding: 0.75rem 1.25rem;
  }
  .typing-dot {
    width: 8px; height: 8px; border-radius: 50%;
    background: var(--text-secondary); opacity: 0.4;
    animation: dotPulse 1.4s infinite ease-in-out;
  }
  .typing-dot:nth-child(2) { animation-delay: 0.2s; }
  .typing-dot:nth-child(3) { animation-delay: 0.4s; }
  @keyframes dotPulse {
    0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; }
    40% { transform: scale(1); opacity: 1; }
  }
  @keyframes fadeInUp {
    from { opacity: 0; transform: translateY(12px); }
    to { opacity: 1; transform: translateY(0); }
  }
  .chat-input-area {
    padding: 1rem 1.5rem; border-top: 1px solid var(--border);
    background: var(--bg-glass); backdrop-filter: blur(12px);
  }
  .chat-input-wrap {
    max-width: 48rem; margin: 0 auto;
    display: flex; align-items: flex-end; gap: 0.75rem;
    background: var(--bg-surface); border: 1px solid var(--border);
    border-radius: var(--radius-lg); padding: 0.5rem;
    transition: border-color 0.2s;
  }
  .chat-input-wrap:focus-within { border-color: var(--primary); }
  .chat-input {
    flex: 1; border: none; background: none; outline: none;
    font-size: 0.9375rem; color: var(--text); resize: none;
    padding: 0.5rem; font-family: inherit; line-height: 1.5;
    min-height: 2.5rem; max-height: 8rem;
  }
  .chat-input::placeholder { color: var(--text-secondary); }
  .send-btn {
    width: 2.5rem; height: 2.5rem; border-radius: var(--radius);
    background: linear-gradient(135deg, var(--primary), var(--accent));
    color: white; border: none; cursor: pointer;
    display: flex; align-items: center; justify-content: center;
    flex-shrink: 0; transition: all 0.2s;
  }
  .send-btn:hover:not(:disabled) { transform: scale(1.05); }
  .send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
