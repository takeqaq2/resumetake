export const prerender = true;

export async function GET() {
  const pages = [
    { path: '/', priority: '1.0', changefreq: 'weekly' },
    { path: '/editor', priority: '0.9', changefreq: 'monthly' },
    { path: '/templates', priority: '0.8', changefreq: 'weekly' },
  ];

  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${pages.map(p => `  <url>
    <loc>https://resume.takee.top${p.path}</loc>
    <lastmod>${new Date().toISOString().split('T')[0]}</lastmod>
    <changefreq>${p.changefreq}</changefreq>
    <priority>${p.priority}</priority>
  </url>`).join('\n')}
</urlset>`;

  return new Response(xml, {
    headers: { 'Content-Type': 'application/xml', 'Cache-Control': 'max-age=3600' }
  });
}
