<script>
  import '../app.css';
  import { page } from '$app/stores';
</script>

<header style="position:sticky;top:0;z-index:50;backdrop-filter:blur(12px);background:color-mix(in srgb,var(--bg) 80%,transparent);border-bottom:1px solid var(--border)">
  <nav class="container" style="display:flex;align-items:center;justify-content:space-between;height:4rem">
    <a href="/" style="display:flex;align-items:center;gap:0.625rem">
      <div style="width:2.25rem;height:2.25rem;background:linear-gradient(135deg,#3b82f6,#8b5cf6);border-radius:0.625rem;display:flex;align-items:center;justify-content:center">
        <span style="color:white;font-weight:700;font-size:1.125rem">R</span>
      </div>
      <span style="font-weight:600;font-size:1.125rem">ResumeTake</span>
    </a>
    <nav style="display:flex;align-items:center;gap:0.5rem">
      <a href="/" class="btn btn-secondary" style="padding:0.5rem 1rem">首页</a>
      <a href="/editor" class="btn btn-primary" style="padding:0.5rem 1rem">免费开始</a>
    </nav>
  </nav>
</header>

<main style="flex:1">
  <slot />
</main>

<footer style="background:var(--bg-surface);border-top:1px solid var(--border);padding:3rem 0;margin-top:4rem">
  <div class="container" style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:1rem">
    <span style="color:var(--text-secondary);font-size:0.875rem">&copy; 2026 ResumeTake. AI简历优化工具</span>
    <div style="display:flex;gap:1.5rem;color:var(--text-secondary);font-size:0.875rem">
      <a href="/editor">创建简历</a>
      <a href="/templates">模板</a>
    </div>
  </div>
</footer>
