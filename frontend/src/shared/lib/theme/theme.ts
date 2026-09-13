import { useSyncExternalStore } from 'react';

type Theme = 'light' | 'dark';
const key = 'funtime.theme';
const eventName = 'funtime:theme';

function getTheme(): Theme { return document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light'; }
function subscribe(callback: () => void) {
  const onStorage = (event: StorageEvent) => {
    if (event.key === key || event.key === null) {
      document.documentElement.dataset.theme = event.newValue === 'dark' ? 'dark' : 'light';
      callback();
    }
  };
  window.addEventListener(eventName, callback);
  window.addEventListener('storage', onStorage);
  return () => { window.removeEventListener(eventName, callback); window.removeEventListener('storage', onStorage); };
}
export function useTheme() {
  const theme = useSyncExternalStore(subscribe, getTheme, (): Theme => 'light');
  function setTheme(value: Theme) {
    document.documentElement.dataset.theme = value;
    try { localStorage.setItem(key, value); } catch { /* Theme still works if storage is unavailable. */ }
    window.dispatchEvent(new Event(eventName));
  }
  return { theme, setTheme };
}
