// Theme state shared by the whole UI. The active theme is stored on
// <html data-theme="..."> (see src/app.css) and persisted in localStorage.

export const THEMES = [
  { id: 'dark', label: 'Dark' },
  { id: 'light', label: 'Light' },
  { id: 'albion', label: 'Albion' },
];

const KEY = 'theme';
const DEFAULT_THEME = 'dark';

function isValid(id) {
  return THEMES.some((t) => t.id === id);
}

function load() {
  try {
    const stored = localStorage.getItem(KEY);
    if (isValid(stored)) return stored;
  } catch {
    // storage unavailable - fall through to the default
  }
  return DEFAULT_THEME;
}

export const theme = $state({ current: DEFAULT_THEME });

function apply(id) {
  document.documentElement.dataset.theme = id;
}

export function initTheme() {
  theme.current = load();
  apply(theme.current);
}

export function setTheme(id) {
  if (!isValid(id)) return;
  theme.current = id;
  apply(id);
  try {
    localStorage.setItem(KEY, id);
  } catch {
    // not fatal: the theme just won't persist
  }
}
