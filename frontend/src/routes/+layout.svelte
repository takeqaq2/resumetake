<script>
  import '../app.css';
</script>

<svelte:head>
  <meta name="theme-color" content="#2563eb">
  <link rel="icon" type="image/svg+xml" href="/favicon.svg">
  <link rel="apple-touch-icon" href="/apple-touch-icon.png">
  <link rel="canonical" href="https://resume.takee.top{$page.url.pathname}">
  <meta property="og:site_name" content="ResumeTake">
  <meta property="og:locale" content="zh_CN">
</svelte:head>

<script>
  import { page } from '$app/stores';
</script>

<header style="position:sticky;top:0;z-index:50;backdrop-filter:blur(16px);-webkit-backdrop-filter:blur(16px);background:color-mix(in srgb,var(--bg) 85%,transparent);border-bottom:1px solid var(--border);transition:background 0.2s">
  <nav class="container" style="display:flex;align-items:center;justify-content:space-between;height:4rem">
    <a href="/" style="display:flex;align-items:center;gap:0.625rem;text-decoration:none">
      <div style="width:2.25rem;height:2.25rem;background:linear-gradient(135deg,#3b82f6,#8b5cf6);border-radius:0.625rem;display:flex;align-items:center;justify-content:center;box-shadow:0 2px 8px rgba(59,130,246,0.3)">
        <span style="color:white;font-weight:700;font-size:1.125rem">R</span>
      </div>
      <span style="font-weight:600;font-size:1.125rem;color:var(--text)">ResumeTake</span>
    </a>
    <nav style="display:flex;align-items:center;gap:0.5rem">
      <a href="/" class="btn btn-secondary" style="padding:0.5rem 1rem">首页</a>
      <a href="/editor" class="btn btn-primary" style="padding:0.5rem 1rem">免费开始</a>
    </nav>
  </nav>
</header>

<main style="flex:1;min-height:calc(100vh - 4rem)">
  <slot />
</main>

<footer style="background:var(--bg-surface);border-top:1px solid var(--border);padding:2.5rem 0;margin-top:4rem">
  <div class="container" style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:1rem">
    <div>
      <span style="color:var(--text-secondary);font-size:0.875rem">&copy; {new Date().getFullYear()} ResumeTake. All rights reserved.</span>
    </div>
    <div style="display:flex;gap:1.5rem;color:var(--text-secondary);font-size:0.875rem">
      <a href="/editor" style="color:var(--text-secondary);transition:color 0.2s">创建简历</a>
      <a href="/templates" style="color:var(--text-secondary);transition:color 0.2s">模板</a>
    </div>
  </div>
</footer>
