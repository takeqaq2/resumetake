export const prerender = true;

import { LANGUAGES } from '$lib/i18n/index.js';

export async function GET() {
  const langs = Object.keys(LANGUAGES);
  const pages = [
    { path: '', priority: '1.0', changefreq: 'weekly' },
    { path: '/editor', priority: '0.9', changefreq: 'monthly' },
    { path: '/templates', priority: '0.8', changefreq: 'weekly' },
  ];

  let urls = '';

  for (const lang of langs) {
    for (const page of pages) {
      const lastmod = new Date().toISOString().split('T')[0];
      urls += `  <url>
    <loc>https://resume.takee.top/${lang}${page.path}</loc>
    <lastmod>${lastmod}</lastmod>
    <changefreq>${page.changefreq}</changefreq>
    <priority>${page.priority}</priority>
    <xhtml:link rel="alternate" hreflang="${lang}" href="https://resume.takee.top/${lang}${page.path}"/>
${langs.filter(l => l !== lang).map(l => `    <xhtml:link rel="alternate" hreflang="${l}" href="https://resume.takee.top/${l}${page.path}"/>`).join('\n')}
    <xhtml:link rel="alternate" hreflang="x-default" href="https://resume.takee.top/en${page.path}"/>
  </url>\n`;
    }
  }

  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
        xmlns:xhtml="http://www.w3.org/1999/xhtml">
${urls}</urlset>`;

  return new Response(xml, {
    headers: { 'Content-Type': 'application/xml', 'Cache-Control': 'max-age=3600' }
  });
}
