import { LANGUAGES, getTranslation } from '$lib/i18n/index.js';

const RTL_LANGS = ['ar'];

// Security headers applied to every response. CSP is kept loose here because
// Nginx already sets a strict CSP at the edge; these headers cover the
// SvelteKit dev server and any direct-to-backend paths.
const SECURITY_HEADERS = {
  'X-Content-Type-Options': 'nosniff',
  'X-Frame-Options': 'SAMEORIGIN',
  'Referrer-Policy': 'strict-origin-when-cross-origin',
  'Permissions-Policy': 'camera=(), microphone=(), geolocation=()'
};

/** @type {import('@sveltejs/kit').Handle} */
export async function handle({ event, resolve }) {
  // SvelteKit's `handle` hook runs BEFORE route matching, so event.params.lang
  // is NOT populated here (it's only available inside load functions and
  // components). Extract the language from the URL pathname instead.
  const m = event.url.pathname.match(/^\/([a-z]{2})(?:\/|$)/);
  const lang = m && LANGUAGES[m[1]] ? m[1] : 'en';
  const dir = RTL_LANGS.includes(lang) ? 'rtl' : 'ltr';
  const t = getTranslation(lang);
  const skipText = (t.nav && t.nav.skipToContent) || 'Skip to content';

  const response = await resolve(event, {
    transformPageChunk: ({ html }) => {
      // Robustly replace the <html ...> tag — the previous regex only matched
      // `<html lang="en">` exactly and would silently fail if app.html added
      // a class, changed spacing, or used single quotes, leaving RTL pages
      // stuck in LTR with no error. This matches any <html ...> tag.
      let out = html
        .replace(/<html\b[^>]*>/, `<html lang="${lang}" dir="${dir}">`)
        .replace('>Skip to content<', `>${skipText}<`);
      // /admin uses ssr=false, so the initial HTML keeps app.html's default
      // "index, follow" robots meta until hydration swaps it. Replace it
      // server-side so crawlers never index admin paths.
      if (event.url.pathname.startsWith('/admin')) {
        out = out.replace(
          '<meta name="robots" content="index, follow',
          '<meta name="robots" content="noindex, nofollow'
        );
      }
      return out;
    }
  });

  // Inject security headers on all responses (pages + endpoints).
  for (const [key, value] of Object.entries(SECURITY_HEADERS)) {
    response.headers.set(key, value);
  }

  return response;
}
