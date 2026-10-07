export function roomInviteUrl(code: string): string {
  const loopback = ['localhost', '127.0.0.1', '[::1]'].includes(location.hostname);
  const origin = loopback && import.meta.env.DEV && import.meta.env.VITE_INVITE_ORIGIN
    ? import.meta.env.VITE_INVITE_ORIGIN : location.origin;
  return new URL(`/rooms/${encodeURIComponent(code)}`, origin).href;
}

export async function copyText(text: string): Promise<boolean> {
  try { if (navigator.clipboard?.writeText) { await navigator.clipboard.writeText(text); return true; } } catch { /* Use the HTTP-compatible fallback. */ }
  const previous = document.activeElement;
  const input = document.createElement('textarea');
  input.value = text;
  input.readOnly = true;
  input.style.cssText = 'position:fixed;top:0;left:0;width:1px;height:1px;opacity:0;';
  // Append inside a modal when one is open, so focus is not trapped outside it.
  (document.querySelector('dialog[open]') ?? document.body).append(input);
  try { input.select(); input.setSelectionRange(0, text.length); return document.execCommand('copy'); }
  catch { return false; }
  finally { input.remove(); if (previous instanceof HTMLElement) previous.focus({ preventScroll: true }); }
}
