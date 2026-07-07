<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';
  import { onMount } from 'svelte';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  let messages = $state([]);
  let inputText = $state('');
  let isTyping = $state(false);
  let showViewButton = $state(false);
  let chatContainer = $state(null);

  let generateLocked = $state(true);
  const lockedText = {
    zh: { title: '0基础生成暂未开放', desc: '这个功能会消耗更多 AI token，之后将按次付费开放。预计中国区约 ¥9.9/次。', cta: '查看付费方案' },
    en: { title: 'Resume generation is temporarily locked', desc: 'This feature uses more AI tokens and will be available as a pay-per-use feature. US pricing is planned around $1.9/use.', cta: 'View pricing' }
  };

  let lock = $derived(lockedText[lang] || lockedText.en);

  function scrollToBottom() {
    if (chatContainer) {
      chatContainer.scrollTop = chatContainer.scrollHeight;
    }
  }

  onMount(() => {
    messages = [{
      role: 'ai',
      content: lang === 'zh'
        ? '你好！我是你的简历生成助手。让我来帮你从零开始构建一份专业的简历。首先，请告诉我你的姓名和目标职位是什么？'
        : "Hello! I'm your resume generation assistant. Let me help you build a professional resume from scratch. First, what's your name and target position?"
    }];
    fetch('/api/health').then(r => r.json()).then(d => {
      generateLocked = !(d.generate_resume_enabled === true);
    }).catch(() => {});
  });

  async function sendMessage() {
    if (!inputText.trim() || isTyping) return;
    const userMsg = inputText.trim();
    messages = [...messages, { role: 'user', content: userMsg }];
    inputText = '';
    isTyping = true;
    scrollToBottom();

    try {
      const token = localStorage.getItem('token');
      const headers = { 'Content-Type': 'application/json' };
      if (token) headers['Authorization'] = 'Bearer ' + token;
      const res = await fetch('/api/v1/generate-resume', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          messages: messages.map(m => ({ role: m.role === 'ai' ? 'assistant' : m.role, content: m.content })),
          lang
        })
      });
      const data = await res.json();
      if (data.success && (data.data?.message || data.message)) {
        const msg = data.data?.message || data.message;
        messages = [...messages, { role: 'ai', content: msg }];
        if (data.data?.resume_complete || data.resume_complete) {
          showViewButton = true;
        }
      } else {
        messages = [...messages, { role: 'ai', content: data.data?.message || data.message || (lang === 'zh' ? '抱歉，出了点问题，请重试。' : 'Sorry, something went wrong. Please try again.') }];
      }
    } catch {
      messages = [...messages, { role: 'ai', content: lang === 'zh' ? '网络错误，请重试。' : 'Network error, please try again.' }];
    } finally {
      isTyping = false;
      scrollToBottom();
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  }
</script>

<svelte:head>
  <title>{t.generate.title} - ResumeTake</title>
</svelte:head>

<div class="generate-page">
  <div class="generate-header">
    <div class="container" style="position:relative;display:flex;align-items:center;justify-content:space-between">
      <div>
        <h1 style="font-size:clamp(1.25rem,2.5vw,1.5rem);font-weight:700;margin-bottom:0.25rem">{t.generate.title}</h1>
        <p style="color:var(--text-secondary);font-size:0.875rem">{t.generate.subtitle}</p>
      </div>
      {#if showViewButton}
        <a href="/{lang}/editor" class="btn btn-primary" style="white-space:nowrap"
          onclick={() => {
            const lastAi = [...messages].reverse().find(m => m.role === 'ai');
            if (lastAi?.content) {
              try {
                const parsed = JSON.parse(lastAi.content);
                if (parsed.resume) localStorage.setItem('generated_resume', JSON.stringify(parsed.resume));
              } catch {}
            }
          }}>
          {lang === 'zh' ? '查看优化简历' : 'View Optimized Resume'} →
        </a>
      {/if}
    </div>
  </div>

  {#if generateLocked}
    <div class="locked-panel">
      <div class="locked-card">
        <div class="locked-icon">🔒</div>
        <h2>{lock.title}</h2>
        <p>{lock.desc}</p>
        <a href="/{lang}/pricing" class="btn btn-primary">{lock.cta}</a>
      </div>
    </div>
  {:else}
  <div class="chat-container" bind:this={chatContainer}>
    <div class="chat-messages">
      {#each messages as msg}
        <div class="chat-msg {msg.role === 'user' ? 'user-msg' : 'ai-msg'}">
          {#if msg.role === 'ai'}
            <div class="ai-avatar">R</div>
          {/if}
          <div class="msg-bubble">
            {msg.content}
          </div>
          {#if msg.role === 'user'}
            <div class="user-avatar">👤</div>
          {/if}
        </div>
      {/each}

      {#if isTyping}
        <div class="chat-msg ai-msg">
          <div class="ai-avatar">R</div>
          <div class="msg-bubble typing-bubble">
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
        bind:value={inputText}
        onkeydown={handleKeydown}
        rows="1"
      ></textarea>
      <button class="send-btn" onclick={sendMessage} disabled={!inputText.trim() || isTyping} aria-label={lang === 'zh' ? '发送' : 'Send'}>
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </button>
    </div>
  </div>
  {/if}
</div>

<style>
  .generate-page {
    display: flex; flex-direction: column; height: calc(100vh - 4rem);
  }
  .generate-header {
    padding: 1rem 0; border-bottom: 1px solid var(--border);
    background: var(--bg-glass); backdrop-filter: blur(12px);
  }
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
  .chat-input::placeholder { color: var(--text-secondary); opacity: 0.6; }
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
