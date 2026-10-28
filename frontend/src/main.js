import './app.css';
import App from './App.svelte';
import { mount } from 'svelte';
import { initTheme } from './lib/theme.svelte.js';

// Apply the saved theme (default: dark) before the first render.
initTheme();

const app = mount(App, {
  target: document.getElementById('app'),
});

export default app;
