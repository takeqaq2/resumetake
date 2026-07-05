<svelte:head>
  <title>简历模板 - ResumeTake | 免费专业简历模板</title>
  <meta name="description" content="多种专业简历模板，覆盖互联网、金融、教育等行业，免费使用。">
  <meta name="keywords" content="简历模板,免费简历模板,专业简历模板,简历设计">
  <link rel="canonical" href="https://resume.takee.top/templates">
  <meta property="og:title" content="简历模板 - ResumeTake">
  <meta property="og:description" content="多种专业简历模板，免费使用。">
</svelte:head>

<div class="container" style="padding:3rem 1.5rem">
  <div style="text-align:center;margin-bottom:3rem">
    <h1 style="font-size:2.25rem;font-weight:700;margin-bottom:0.75rem;color:var(--text)">专业简历模板</h1>
    <p style="color:var(--text-secondary)">选择适合你行业的模板，一键套用</p>
  </div>
  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:1.5rem">
    {#each [
      {id:'professional',name:'专业商务',desc:'适合传统行业和商务岗位',gradient:'linear-gradient(135deg,#3b82f6,#2563eb)'},
      {id:'modern',name:'现代简约',desc:'适合互联网和科技行业',gradient:'linear-gradient(135deg,#8b5cf6,#7c3aed)'},
      {id:'creative',name:'创意设计',desc:'适合设计和创意岗位',gradient:'linear-gradient(135deg,#ec4899,#db2777)'},
      {id:'academic',name:'学术科研',desc:'适合教育和研究岗位',gradient:'linear-gradient(135deg,#10b981,#059669)'},
      {id:'executive',name:'高管专用',desc:'适合高级管理岗位',gradient:'linear-gradient(135deg,#374151,#111827)'},
      {id:'minimal',name:'极简风格',desc:'简洁大方，通用性强',gradient:'linear-gradient(135deg,#f59e0b,#d97706)'}
    ] as t}
      <a href="/editor?template={t.id}" class="card" style="padding:0;overflow:hidden;transition:transform 0.2s,box-shadow 0.2s;text-decoration:none">
        <div style="height:10rem;background:{t.gradient};display:flex;align-items:center;justify-content:center">
          <span style="color:rgba(255,255,255,0.2);font-size:3.5rem;font-weight:700">Aa</span>
        </div>
        <div style="padding:1.25rem">
          <h3 style="font-weight:600;margin-bottom:0.25rem;color:var(--text)">{t.name}</h3>
          <p style="font-size:0.875rem;color:var(--text-secondary)">{t.desc}</p>
        </div>
      </a>
    {/each}
  </div>
</div>
