import { redirect } from '@sveltejs/kit';
import { LANGUAGES } from '$lib/i18n/index.js';

const supported = new Set(Object.keys(LANGUAGES));

function detectLanguage(acceptLanguage = '') {
  const parsed = acceptLanguage
    .split(',')
    .map((item) => {
      const [tag, qPart] = item.trim().split(';');
      const q = qPart?.startsWith('q=') ? Number(qPart.slice(2)) : 1;
      const code = tag?.toLowerCase().split('-')[0];
      return { code, q: Number.isFinite(q) ? q : 0 };
    })
    .filter((item) => item.code && supported.has(item.code) && item.q > 0)
    .sort((a, b) => b.q - a.q);

  return parsed[0]?.code || 'en';
}

export function load({ request, url }) {
  const lang = detectLanguage(request.headers.get('accept-language') || '');
  // NOTE: url.hash is NOT available on the server (hash is client-side only,
  // never sent in the HTTP request). Using it here throws in SvelteKit 2:
  // "Cannot access event.url.hash". The browser preserves the hash across a
  // same-origin 302 redirect automatically, so we only forward search here.
  throw redirect(302, `/${lang}${url.search}`);
}
