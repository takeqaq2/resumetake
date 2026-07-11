import { env } from '$env/dynamic/public';

export const prerender = false;

export function GET() {
  const client = env.PUBLIC_GOOGLE_ADSENSE_CLIENT || '';
  const publisherId = client.replace(/^ca-/, '');
  const body = /^pub-\d+$/.test(publisherId)
    ? `google.com, ${publisherId}, DIRECT, f08c47fec0942fa0\n`
    : '# Configure PUBLIC_GOOGLE_ADSENSE_CLIENT=ca-pub-xxxxxxxxxxxxxxxx to enable ads.txt\n';

  return new Response(body, {
    headers: {
      'Content-Type': 'text/plain; charset=utf-8',
      'Cache-Control': 'max-age=3600'
    }
  });
}
