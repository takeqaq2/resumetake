export const prerender = true;

export async function GET() {
  const robots = `User-agent: *
Allow: /
Disallow: /api/

Sitemap: https://resume.takee.top/sitemap.xml

# Multilingual SEO
# Available languages: en, zh, ja, ko, es, pt, fr, de, ar, hi
# Default language: en`;

  return new Response(robots, {
    headers: { 'Content-Type': 'text/plain', 'Cache-Control': 'max-age=86400' }
  });
}
