<svelte:head>
  <title>AI简历优化工具 - ResumeTake | 智能简历制作平台</title>
  <meta name="description" content="AI智能简历优化工具，一键生成专业简历，ATS关键词匹配，提升求职成功率。">
  <link rel="canonical" href="https://resume.takee.top/">
</svelte:head>

<section style="text-align:center;padding:5rem 0">
  <div class="container">
    <div style="display:inline-flex;align-items:center;gap:0.5rem;padding:0.375rem 1rem;border-radius:9999px;background:rgba(37,99,235,0.1);color:var(--primary);font-size:0.875rem;font-weight:500;margin-bottom:1.5rem">
      ✨ AI驱动的智能简历优化
    </div>
    <h1 style="font-size:clamp(2.25rem,5vw,3.75rem);font-weight:700;line-height:1.1;margin-bottom:1.5rem">
      让AI帮你打造<br><span style="background:linear-gradient(135deg,#2563eb,#8b5cf6);-webkit-background-clip:text;-webkit-text-fill-color:transparent">完美简历</span>
    </h1>
    <p style="font-size:1.125rem;color:var(--text-secondary);max-width:36rem;margin:0 auto 2.5rem;line-height:1.7">
      上传简历，AI自动优化内容、匹配ATS关键词、提升通过率。一键生成专业简历。
    </p>
    <div style="display:flex;gap:1rem;justify-content:center;flex-wrap:wrap">
      <a href="/editor" class="btn btn-primary" style="padding:0.875rem 2rem;font-size:1rem">立即开始 →</a>
      <a href="#features" class="btn btn-secondary" style="padding:0.875rem 2rem;font-size:1rem">了解更多</a>
    </div>
  </div>
</section>

<section id="features" style="padding:5rem 0;background:var(--bg-surface)">
  <div class="container">
    <h2 style="font-size:2rem;font-weight:700;text-align:center;margin-bottom:1rem">为什么选择 ResumeTake</h2>
    <p style="text-align:center;color:var(--text-secondary);margin-bottom:3rem">基于先进AI技术，让简历优化变得简单高效</p>
    <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:1.5rem">
      {#each [
        { icon: '⚡', title: 'AI智能优化', desc: '自动优化简历内容、用词和格式，突出核心竞争力' },
        { icon: '🔍', title: 'ATS关键词匹配', desc: '智能分析目标职位，自动匹配ATS关键词，提升通过率' },
        { icon: '📄', title: '专业模板', desc: '多种模板覆盖各行各业，一键切换风格，导出PDF' }
      ] as f}
        <div class="card" style="text-align:center;padding:2rem">
          <div style="font-size:2.5rem;margin-bottom:1rem">{f.icon}</div>
          <h3 style="font-weight:600;margin-bottom:0.5rem">{f.title}</h3>
          <p style="color:var(--text-secondary);font-size:0.9375rem">{f.desc}</p>
        </div>
      {/each}
    </div>
  </div>
</section>

<section style="padding:5rem 0">
  <div class="container" style="text-align:center">
    <h2 style="font-size:2rem;font-weight:700;margin-bottom:1rem">开始优化你的简历</h2>
    <p style="color:var(--text-secondary);margin-bottom:2.5rem">免费使用AI简历优化工具，提升求职成功率</p>
    <a href="/editor" class="btn btn-primary" style="padding:0.875rem 2.5rem;font-size:1rem">免费开始 →</a>
  </div>
</section>
