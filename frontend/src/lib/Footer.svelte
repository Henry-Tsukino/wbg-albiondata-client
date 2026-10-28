<script>
  import { onDestroy } from 'svelte';
  import { Browser, Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/Henry-Tsukino/wbg-albiondata-client/internal/dashboard/index.js';
  import { THEMES, theme, setTheme } from './theme.svelte.js';

  const REPO_URL = 'https://github.com/Henry-Tsukino/wbg-albiondata-client';

  let version = $state('');

  DashboardService.GetStatus().then((s) => (version = s.Version));

  const unlisten = Events.On('status:changed', (evt) => {
    version = evt.data.Version;
  });

  onDestroy(unlisten);

  function openLink(e, url) {
    e.preventDefault();
    Browser.OpenURL(url);
  }
</script>

<footer class="statusbar">
  <span class="pane grow">
    <a href={REPO_URL} onclick={(e) => openLink(e, REPO_URL)}>github.com/Henry-Tsukino/wbg-albiondata-client</a>
  </span>

  <span class="pane themes" role="radiogroup" aria-label="Theme">
    {#each THEMES as t}
      <button
        type="button"
        class="theme-btn"
        class:active={theme.current === t.id}
        role="radio"
        aria-checked={theme.current === t.id}
        onclick={() => setTheme(t.id)}
      >{t.label}</button>
    {/each}
  </span>

  <span class="pane">Version {version || 'dev'}</span>
</footer>

<style>
  .statusbar {
    display: flex;
    gap: 4px;
    padding: 3px 4px;
    background: var(--status-bg);
    border-top: 1px solid var(--status-border);
    flex-shrink: 0;
  }
  .pane {
    display: flex;
    align-items: center;
    padding: 2px 8px;
    border: 1px solid var(--pane-border);
    border-radius: 2px;
    white-space: nowrap;
    color: var(--text);
  }
  .grow {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .pane a {
    color: var(--text);
    text-decoration: none;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .pane a:hover {
    text-decoration: underline;
    color: var(--link-hover);
  }

  .themes {
    gap: 3px;
    padding: 1px 3px;
  }
  .theme-btn {
    padding: 1px 8px;
    border: 1px solid transparent;
    border-radius: 2px;
    background: none;
    color: var(--text-muted);
    font: inherit;
    cursor: pointer;
  }
  .theme-btn:hover {
    color: var(--text);
    border-color: var(--btn-border);
  }
  .theme-btn.active {
    color: var(--btn-text);
    border-color: var(--btn-border);
    background: linear-gradient(to bottom, var(--btn-bg-from), var(--btn-bg-to));
    font-weight: bold;
  }
  .theme-btn:focus-visible {
    outline: 1px solid var(--btn-hover-ring);
  }
</style>
