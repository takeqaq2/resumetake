<script>
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
        body: JSON.stringify({ resume_content: resume, target_job: targetJob, job_description: jobDescription })
      });
      const data = await res.json();
      result = data.data;
      tab = 'result';
    } catch (e) {
      alert('优化失败，请重试');
    } finally {
      isOptimizing = false;
    }
  }
</script>

<svelte:head>
  <title>简历编辑器 - ResumeTake</title>
  <meta name="description" content="使用AI智能优化你的简历，匹配ATS关键词。">
</svelte:head>

<div class="container" style="padding:2.5rem 1.5rem">
  <h1 style="font-size:1.875rem;font-weight:700;margin-bottom:0.5rem">简历编辑器</h1>
  <p style="color:var(--text-secondary);margin-bottom:2rem">填写信息，AI将自动优化内容</p>

  <div style="display:grid;grid-template-columns:1fr 1fr;gap:2rem">
    <div style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="card">
        <h2 style="font-weight:600;margin-bottom:1.25rem">基本信息</h2>
        <div style="display:flex;flex-direction:column;gap:1rem">
          <div>
            <label class="label">姓名</label>
            <input class="input" placeholder="请输入姓名" bind:value={resume.name}>
          </div>
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:1rem">
            <div>
              <label class="label">邮箱</label>
              <input type="email" class="input" placeholder="email@example.com" bind:value={resume.email}>
            </div>
            <div>
              <label class="label">电话</label>
              <input type="tel" class="input" placeholder="13800138000" bind:value={resume.phone}>
            </div>
          </div>
          <div>
            <label class="label">个人简介</label>
            <textarea class="input" rows="4" placeholder="简要介绍你的专业背景..." bind:value={resume.summary}></textarea>
          </div>
        </div>
      </div>

      <div class="card">
        <h2 style="font-weight:600;margin-bottom:1.25rem">目标职位</h2>
        <div style="display:flex;flex-direction:column;gap:1rem">
          <div>
            <label class="label">目标职位</label>
            <input class="input" placeholder="例如：产品经理、前端工程师" bind:value={targetJob}>
          </div>
          <div>
            <label class="label">职位描述（可选）</label>
            <textarea class="input" rows="4" placeholder="粘贴职位描述..." bind:value={jobDescription}></textarea>
          </div>
        </div>
      </div>

      <button class="btn btn-primary" style="width:100%;padding:0.875rem;font-size:1rem" onclick={optimize} disabled={isOptimizing}>
        {isOptimizing ? 'AI优化中...' : '✨ AI智能优化'}
      </button>
    </div>

    <div>
      <div class="card" style="min-height:500px">
        <div style="display:flex;gap:1rem;border-bottom:1px solid var(--border);margin-bottom:1.25rem">
          <button class="btn" style="padding:0.5rem 1rem;border-bottom:2px solid {tab === 'edit' ? 'var(--primary)' : 'transparent'};color:{tab === 'edit' ? 'var(--primary)' : 'var(--text-secondary)'};border-radius:0" onclick={() => tab = 'edit'}>预览简历</button>
          <button class="btn" style="padding:0.5rem 1rem;border-bottom:2px solid {tab === 'result' ? 'var(--primary)' : 'transparent'};color:{tab === 'result' ? 'var(--primary)' : 'var(--text-secondary)'};border-radius:0" onclick={() => tab = 'result'}>优化结果</button>
        </div>

        {#if tab === 'edit'}
          <div style="color:var(--text-secondary)">
            <h3 style="font-size:1.25rem;font-weight:600;color:var(--text)">{resume.name || '你的姓名'}</h3>
            <p style="font-size:0.875rem;margin-top:0.25rem">{resume.email} | {resume.phone}</p>
            <p style="margin-top:1rem;line-height:1.7">{resume.summary || '个人简介将显示在这里...'}</p>
          </div>
        {:else if result}
          <div style="display:flex;flex-direction:column;gap:1.25rem">
            <div style="display:flex;justify-content:space-between;align-items:center;padding:1rem;border-radius:0.75rem;background:rgba(16,185,129,0.1);border:1px solid rgba(16,185,129,0.2)">
              <span style="font-weight:500;color:#059669">ATS匹配度</span>
              <span style="font-size:1.5rem;font-weight:700;color:#059669">{result.ats_score}%</span>
            </div>
            <div>
              <h4 style="font-weight:500;margin-bottom:0.5rem">推荐关键词</h4>
              <div style="display:flex;flex-wrap:wrap;gap:0.5rem">
                {#each result.keywords || [] as kw}
                  <span style="padding:0.25rem 0.75rem;border-radius:9999px;background:rgba(37,99,235,0.1);color:var(--primary);font-size:0.8125rem">{kw}</span>
                {/each}
              </div>
            </div>
            <div>
              <h4 style="font-weight:500;margin-bottom:0.5rem">优化建议</h4>
              <ul style="list-style:none;display:flex;flex-direction:column;gap:0.5rem">
                {#each result.suggestions || [] as s}
                  <li style="font-size:0.875rem;color:var(--text-secondary);display:flex;gap:0.5rem"><span style="color:var(--primary)">•</span>{s}</li>
                {/each}
              </ul>
            </div>
          </div>
        {:else}
          <div style="text-align:center;padding:4rem 0;color:var(--text-secondary)">
            <div style="font-size:2.5rem;margin-bottom:0.75rem;opacity:0.5">✨</div>
            <p style="font-size:0.875rem">点击"AI智能优化"查看结果</p>
          </div>
        {/if}
      </div>

      <button class="btn btn-primary" style="width:100%;padding:0.875rem;margin-top:1rem;font-size:1rem">📥 导出PDF</button>
    </div>
  </div>
</div>
