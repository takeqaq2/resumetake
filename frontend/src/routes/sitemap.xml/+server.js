export const prerender = true;

import { LANGUAGES } from '$lib/i18n/index.js';

export async function GET() {
  const langs = Object.keys(LANGUAGES);
  const pages = [
    { path: '', priority: '1.0', changefreq: 'weekly' },
    { path: '/editor', priority: '0.9', changefreq: 'monthly' },
    { path: '/templates', priority: '0.8', changefreq: 'weekly' },
    { path: '/jobs', priority: '0.8', changefreq: 'daily' },
    { path: '/pricing', priority: '0.8', changefreq: 'monthly' },
    { path: '/generate', priority: '0.7', changefreq: 'weekly' },
    { path: '/privacy', priority: '0.3', changefreq: 'yearly' },
    { path: '/terms', priority: '0.3', changefreq: 'yearly' },
    { path: '/contact', priority: '0.3', changefreq: 'yearly' },
    // /auth excluded — page has noindex,nofollow, listing it in sitemap
    // sends a contradictory signal to search engines.
  ];

  let urls = '';

  // R31-10: omit <lastmod> — Google warns that inaccurate lastmod dates
  // reduce crawl trust more than having no lastmod at all. Since this
  // sitemap is prerendered at build time, the date would be the build date
  // for ALL pages (including /privacy that changes yearly and /jobs that
  // changes daily). Better to let Google rely on changefreq + priority.
  // R48-F1: legal pages (privacy/terms/contact) are now fully localized
  // (R45-R46 added 10-language i18n), so all language versions are included.
  for (const lang of langs) {
    for (const page of pages) {
      urls += `  <url>
    <loc>https://resume.takee.top/${lang}${page.path}</loc>
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
