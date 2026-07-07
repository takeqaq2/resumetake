import { redirect } from '@sveltejs/kit';

const supported = new Set(['zh', 'en', 'ja', 'ko', 'es', 'pt', 'fr', 'de', 'ar', 'hi']);

function detectLanguage(acceptLanguage = '') {
  const parsed = acceptLanguage
    .split(',')
    .map((item) => {
      const [tag, qPart] = item.trim().split(';');
      const q = qPart?.startsWith('q=') ? Number(qPart.slice(2)) : 1;
      const code = tag?.toLowerCase().split('-')[0];
      return { code, q: Number.isFinite(q) ? q : 0 };
    })
    .filter((item) => item.code && supported.has(item.code))
    .sort((a, b) => b.q - a.q);

  return parsed[0]?.code || 'en';
}

export function load({ request }) {
  const lang = detectLanguage(request.headers.get('accept-language') || '');
  throw redirect(302, `/${lang}`);
}
