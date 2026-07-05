<script>
  import { onMount } from 'svelte';
  let mounted = $state(false);
  onMount(() => { mounted = true; });
</script>

<svelte:head>
  <title>AI简历优化工具 - ResumeTake | 智能简历制作平台</title>
  <meta name="description" content="AI智能简历优化工具，一键生成专业简历，ATS关键词匹配，提升求职成功率。免费在线简历制作，支持PDF导出。">
  <meta name="keywords" content="AI简历优化,简历生成器,智能简历制作,ATS简历优化,免费简历工具,在线简历编辑,一键生成简历,求职简历">
  <link rel="canonical" href="https://resume.takee.top/">
  <meta property="og:title" content="AI简历优化工具 - ResumeTake">
  <meta property="og:description" content="AI智能简历优化，一键生成专业简历，ATS关键词匹配。">
  <meta property="og:url" content="https://resume.takee.top/">
  <meta property="og:type" content="website">
  <meta property="og:image" content="https://resume.takee.top/og-image.png">
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content="AI简历优化工具 - ResumeTake">
  <meta name="twitter:description" content="AI智能简历优化，一键生成专业简历。">
  <script type="application/ld+json">
  {
    "@context":"https://schema.org",
    "@type":"WebApplication",
    "name":"ResumeTake AI简历优化工具",
    "url":"https://resume.takee.top",
    "description":"智能AI简历优化工具，一键生成专业简历，ATS关键词匹配，提升求职成功率",
    "applicationCategory":"BusinessApplication",
    "operatingSystem":"Web",
    "offers":{"@type":"Offer","price":"0","priceCurrency":"CNY"},
    "featureList":["AI简历优化","ATS关键词匹配","简历模板","PDF导出"]
  }
  </script>
  <script type="application/ld+json">
  {
    "@context":"https://schema.org",
    "@type":"FAQPage",
    "mainEntity":[
      {"@type":"Question","name":"ResumeTake是什么？","acceptedAnswer":{"@type":"Answer","text":"ResumeTake是一款AI智能简历优化工具，帮助用户一键生成专业简历，自动匹配ATS关键词，提升求职成功率。"}},
      {"@type":"Question","name":"ResumeTake是免费的吗？","acceptedAnswer":{"@type":"Answer","text":"是的，ResumeTake提供免费的基础AI简历优化功能。"}},
      {"@type":"Question","name":"如何使用AI优化简历？","acceptedAnswer":{"@type":"Answer","text":"上传简历或填写信息，输入目标职位，点击AI优化按钮即可获得优化建议。"}}
    ]
  }
  </script>
</svelte:head>

<section style="position:relative;overflow:hidden">
  <div style="position:absolute;inset:0;background:linear-gradient(135deg,rgba(59,130,246,0.05) 0%,rgba(139,92,246,0.05) 100%)"></div>
  <div class="container" style="position:relative;text-align:center;padding:6rem 1.5rem 7rem">
    <div style="display:inline-flex;align-items:center;gap:0.5rem;padding:0.375rem 1rem;border-radius:9999px;background:rgba(37,99,235,0.08);color:var(--primary);font-size:0.8125rem;font-weight:500;margin-bottom:2rem;border:1px solid rgba(37,99,235,0.15)">
      <span style="font-size:0.875rem">✨</span> AI驱动的智能简历优化
    </div>
    <h1 style="font-size:clamp(2.5rem,5.5vw,4rem);font-weight:800;line-height:1.08;margin-bottom:1.5rem;letter-spacing:-0.02em;color:var(--text)">
      让AI帮你打造<br>
      <span style="background:linear-gradient(135deg,#2563eb,#7c3aed);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text">完美简历</span>
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
    <div style="text-align:center;margin-bottom:3rem">
      <h2 style="font-size:2rem;font-weight:700;margin-bottom:0.75rem;color:var(--text)">为什么选择 ResumeTake</h2>
      <p style="color:var(--text-secondary);max-width:32rem;margin:0 auto">基于先进AI技术，让简历优化变得简单高效</p>
    </div>
    <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:1.5rem">
      {#each [
        { icon: '⚡', title: 'AI智能优化', desc: '自动优化简历内容、用词和格式，突出核心竞争力与成就' },
        { icon: '🔍', title: 'ATS关键词匹配', desc: '智能分析目标职位，自动匹配ATS关键词，提升通过率' },
        { icon: '📄', title: '专业模板', desc: '多种模板覆盖各行各业，一键切换风格，导出PDF' }
      ] as f}
        <div class="card" style="text-align:center;padding:2rem;transition:transform 0.2s,box-shadow 0.2s;cursor:default">
          <div style="font-size:2.5rem;margin-bottom:1rem">{f.icon}</div>
          <h3 style="font-weight:600;margin-bottom:0.5rem;color:var(--text)">{f.title}</h3>
          <p style="color:var(--text-secondary);font-size:0.9375rem;line-height:1.6">{f.desc}</p>
        </div>
      {/each}
    </div>
  </div>
</section>

<section style="padding:5rem 0">
  <div class="container" style="text-align:center">
    <h2 style="font-size:2rem;font-weight:700;margin-bottom:0.75rem;color:var(--text)">开始优化你的简历</h2>
    <p style="color:var(--text-secondary);margin-bottom:2.5rem">免费使用AI简历优化工具，提升求职成功率</p>
    <a href="/editor" class="btn btn-primary" style="padding:0.875rem 2.5rem;font-size:1rem">免费开始 →</a>
  </div>
</section>
