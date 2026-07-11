/**
 * Centralized API fetch helper.
 *
 * Why this exists: before, every page reimplemented token injection + error
 * handling inline, and 401/403 responses left stale tokens in localStorage
 * with no redirect to /auth. This helper standardizes that behavior.
 *
 * Behavior:
 *  - Automatically attaches `Authorization: Bearer <token>` when a token is
 *    present in localStorage (unless `skipAuth` is set).
 *  - On 401 (and only 401 — 403 is intentionally excluded as it indicates
 *    a permission issue, not an auth failure; network errors are surfaced
 *    via the thrown exception): clears the stale token, dispatches a
 *    `auth:unauthorized`
 *    CustomEvent so layouts can update their login button, and redirects to
 *    `/<lang>/auth` (lang derived from the current URL).
 *  - Returns the raw Response so callers can inspect status / parse JSON as
 *    needed. Throws TypeError on network failure (same as native fetch).
 *
 * Usage:
 *   import { apiFetch, getToken } from '$lib/api.js';
 *   const res = await apiFetch('/api/v1/optimize', { method: 'POST', body: ... });
 *   const data = await res.json();
 *   if (!res.ok) { ... }
 */

import { goto } from '$app/navigation';
import { LANGUAGES } from '$lib/i18n/index.js';

// Guards against multiple concurrent 401s all firing clearAuth + goto at once
// (common when a page loads several authenticated endpoints in parallel).
// The first 401 triggers the redirect; subsequent ones become no-ops.
let authRedirectInProgress = false;

export function getToken() {
  if (typeof window === 'undefined') return null;
  try {
    return localStorage.getItem('token');
  } catch {
    // localStorage access can throw in private-browsing mode or when storage
    // is disabled by the user. Treat as "no token" rather than crashing.
    return null;
  }
}

export function clearAuth() {
  if (typeof window === 'undefined') return;
  try {
    localStorage.removeItem('token');
  } catch {
    // Ignore — see getToken() for when this can throw.
  }
  // Notify layouts (e.g. [lang]/+layout.svelte) to refresh isLoggedIn state.
  window.dispatchEvent(new CustomEvent('auth:unauthorized'));
}

function currentLang() {
  if (typeof window === 'undefined') return 'en';
  const m = window.location.pathname.match(/^\/([a-z]{2})(?:\/|$)/);
  // Only accept codes we actually support; default to en for anything else.
  if (m && Object.prototype.hasOwnProperty.call(LANGUAGES, m[1])) return m[1];
  return 'en';
}

/**
 * Fetch wrapper with automatic token injection + 401/403 handling.
 *
 * `skipAuth` controls ONLY the 401 redirect behavior:
 *  - skipAuth: false (default) → 401 clears token + redirects to /auth
 *  - skipAuth: true → 401 is surfaced to the caller (no redirect)
 *
 * Token is ALWAYS attached when present, regardless of skipAuth. This lets
 * public pages (editor, templates) probe /auth/me with a stale token without
 * being force-redirected, while still sending the token if the user has one.
 *
 * @param {string} url
 * @param {RequestInit & { skipAuth?: boolean }} [options]
 * @returns {Promise<Response>}
 */
export async function apiFetch(url, options = {}) {
  const { skipAuth, timeout = 30000, headers: rawHeaders, ...rest } = options;
  const headers = new Headers(rawHeaders || {});

  // Always attach token when present — even for skipAuth calls. Public
  // endpoints ignore it; authenticated endpoints need it. skipAuth only
  // affects whether a 401 triggers a redirect to /auth.
  // R51b-F3: only attach token to same-origin URLs. If a caller ever passes
  // an absolute cross-origin URL, the Bearer token would leak to a third
  // party. Relative URLs are always same-origin; for absolute URLs, check
  // the origin matches.
  const token = getToken();
  if (token && !headers.has('Authorization')) {
    let sameOrigin = true;
    if (/^https?:\/\//i.test(url)) {
      try {
        sameOrigin = new URL(url, location.origin).origin === location.origin;
      } catch {
        sameOrigin = false;
      }
    }
    if (sameOrigin) {
      headers.set('Authorization', 'Bearer ' + token);
    }
  }

  // AbortController ensures requests don't hang forever if the server stalls
  // or a network intermediary drops the connection silently. Default 30s;
  // callers can override (e.g. AI endpoints may need 120s).
  // If the caller provided their own signal, link it so abort propagates
  // both ways (caller abort → request abort; timeout → request abort).
  // Previously the caller's signal was silently overwritten by rest spread +
  // signal: controller.signal, so callers couldn't cancel requests.
  const callerSignal = rest.signal;
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);
  // R37b-F3: name the handler so we can removeEventListener in finally.
  // Without this, every request that supplies a long-lived callerSignal
  // (e.g. a page-level AbortController) permanently attaches a listener
  // to that signal — even on successful completion. { once: true } only
  // auto-removes when the event fires, not when the request succeeds.
  const onCallerAbort = () => controller.abort();
  if (callerSignal) {
    if (callerSignal.aborted) {
      controller.abort();
    } else {
      callerSignal.addEventListener('abort', onCallerAbort, { once: true });
    }
  }

  let res;
  try {
    res = await fetch(url, { ...rest, headers, signal: controller.signal });
  } finally {
    clearTimeout(timeoutId);
    if (callerSignal && !callerSignal.aborted) {
      callerSignal.removeEventListener('abort', onCallerAbort);
    }
  }

  // 401 means the token is invalid/expired. Clean up and bounce to /auth
  // so the user isn't stuck looking at "please login" errors on every call.
  // We do NOT redirect for public endpoints (skipAuth) — those opted out
  // so stale-token 401s don't interrupt browsing public pages.
  //
  // 403 is intentionally NOT treated as an auth failure: it means the token
  // is valid but the user lacks permission for this specific action (e.g.
  // free-tier usage limit exceeded, or IDOR protection on a resume they
  // don't own). Clearing the token + redirecting to /auth would log out a
  // perfectly authenticated user. The caller handles 403 in-page.
  if (res.status === 401 && !skipAuth) {
    // Only the first concurrent 401 should trigger clearAuth + goto.
    // Subsequent 401s (from parallel requests on the same page) are no-ops.
    if (!authRedirectInProgress) {
      authRedirectInProgress = true;
      clearAuth();
      if (typeof window !== 'undefined') {
        const lang = currentLang();
        // Avoid double-redirecting if we're already on the auth page.
        if (!window.location.pathname.startsWith(`/${lang}/auth`)) {
          // R52b-F5: pass the current path as ?redirect= so the auth page
          // can return the user to where they were after login, instead
          // of always dumping them on /editor.
          const currentPath = window.location.pathname + window.location.search;
          const redirectParam = encodeURIComponent(currentPath);
          // R53-F3: timeout fallback — if goto()'s Promise never settles
          // (e.g. route guard interception), the flag would stay true
          // forever, silently swallowing all future 401 redirects.
          const redirectTimeout = setTimeout(() => { authRedirectInProgress = false; }, 5000);
          // R56-F1: .catch() before .finally() — if goto rejects (e.g. route
          // guard), the unhandled rejection would log to console. The cleanup
          // in .finally() still runs regardless.
          goto(`/${lang}/auth?redirect=${redirectParam}`).catch(() => {}).finally(() => { clearTimeout(redirectTimeout); authRedirectInProgress = false; });
        } else {
          authRedirectInProgress = false;
        }
      } else {
        authRedirectInProgress = false;
      }
    }
  }

  return res;
}
