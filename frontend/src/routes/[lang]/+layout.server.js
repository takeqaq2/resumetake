import { redirect } from '@sveltejs/kit';
import { LANGUAGES } from '$lib/i18n/index.js';

const VALID_LANGS = Object.keys(LANGUAGES);

export function load({ params, url }) {
  if (!VALID_LANGS.includes(params.lang)) {
    // Preserve the trailing path + query so /xx/editor?x=1 -> /en/editor?x=1
    throw redirect(302, '/en' + url.pathname.replace(/^\/[^\/]+/, '') + url.search);
  }
}
