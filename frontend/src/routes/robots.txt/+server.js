export const prerender = true;

export async function GET() {
  const robots = `User-agent: *
Allow: /
Disallow: /api/

Sitemap: https://resume.takee.top/sitemap.xml`;

  return new Response(robots, {
    headers: { 'Content-Type': 'text/plain', 'Cache-Control': 'max-age=86400' }
  });
}
