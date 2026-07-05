import { redirect } from '@sveltejs/kit';
import { LANGUAGES } from '$lib/i18n/index.js';

const VALID_LANGS = Object.keys(LANGUAGES);

export function load({ params }) {
  if (!VALID_LANGS.includes(params.lang)) {
    throw redirect(302, '/en');
  }
}
