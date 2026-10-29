<script>
  import { onDestroy } from 'svelte';
  import { Browser, Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/Henry-Tsukino/wbg-albiondata-client/internal/dashboard/index.js';
  import { setMarketDisabled, setUploadDisabled } from './api.js';
  import CountersPanel from './CountersPanel.svelte';

  const WIDTH_KEY = 'sidebar-width';
  const MIN_WIDTH = 168;
  const MAX_WIDTH = 360;
  const DEFAULT_WIDTH = 190;

  function loadWidth() {
    const stored = Number(localStorage.getItem(WIDTH_KEY));
    if (Number.isFinite(stored) && stored >= MIN_WIDTH && stored <= MAX_WIDTH) return stored;
    return DEFAULT_WIDTH;
  }

  let width = $state(loadWidth());
  let resizing = $state(false);

  function startResize(e) {
    e.preventDefault();
    resizing = true;
    window.addEventListener('pointermove', onResize);
    window.addEventListener('pointerup', stopResize);
  }

  function onResize(e) {
    width = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, e.clientX));
  }

  function stopResize() {
    resizing = false;
    localStorage.setItem(WIDTH_KEY, String(width));
    window.removeEventListener('pointermove', onResize);
    window.removeEventListener('pointerup', stopResize);
  }

  onDestroy(() => {
    window.removeEventListener('pointermove', onResize);
    window.removeEventListener('pointerup', stopResize);
  });

  let status = $state({
    Version: '',
    UpdateAvailable: '',
    CaptureRunning: false,
    CaptureError: false,
    ServerID: 0,
    IngestBaseURL: '',
    CustomPublicIngest: false,
    DriverWarning: '',
    DriverHelpURL: '',
    EncryptionStatus: '',
    MarketDisabled: false,
    UploadDisabled: false,
  });

  DashboardService.GetStatus().then((s) => (status = s));

  const unlisten = Events.On('status:changed', (evt) => {
    status = evt.data;
  });

  onDestroy(unlisten);

  const serverNames = { 0: 'Unknown', 1: 'West', 2: 'East', 3: 'Europe' };
  let serverLabel = $derived(
    status.CustomPublicIngest ? 'Private' : (serverNames[status.ServerID] ?? status.ServerID)
  );
  let badgeLabel = $derived(
    status.CaptureError ? 'Error' : status.CaptureRunning ? 'Capturing' : 'Ready'
  );
  let encryptionLabel = $derived(
    status.EncryptionStatus === 'encrypted'
      ? 'Encrypted'
      : status.EncryptionStatus === 'clear'
        ? 'Not Encrypted'
        : 'Encrypted?'
  );

  function openDriverHelp(e) {
    e.preventDefault();
    Browser.OpenURL(status.DriverHelpURL);
  }

  async function toggleUpload() {
    const previous = status.UploadDisabled;
    status = { ...status, UploadDisabled: !previous };
    try {
      await setUploadDisabled(!previous);
    } catch (e) {
      console.error('SetUploadDisabled failed:', e);
      status = { ...status, UploadDisabled: previous };
    }
  }

  async function toggleMarket() {
    const previous = status.MarketDisabled;
    // Optimistic update so the checkbox reacts immediately; the backend
    // then confirms via the "status:changed" event.
    status = { ...status, MarketDisabled: !previous };
    try {
      await setMarketDisabled(!previous);
    } catch (e) {
      console.error('SetMarketDisabled failed:', e);
      status = { ...status, MarketDisabled: previous };
    }
  }
</script>

<aside class="sidebar" style="width: {width}px">
  <div class="brand">
    <img src="/favicon.svg" alt="" width="32" height="32" />
    <span class="name">WBG Client</span>
  </div>

  <section class="panel">
    <header>Capture status</header>
    <div class="panel-body">
      <div class="row">
        <span class="led" class:on={status.CaptureRunning && !status.CaptureError} class:err={status.CaptureError}></span>
        <span>{badgeLabel}</span>
      </div>
      <div class="row">
        <span class="led" class:warn={status.EncryptionStatus === 'encrypted'} class:on={status.EncryptionStatus === 'clear'}></span>
        <span>{encryptionLabel}</span>
      </div>
      <div class="kv"><span>Server:</span><b>{serverLabel}</b></div>

      <button
        class="xp-check"
        onclick={toggleMarket}
        role="checkbox"
        aria-checked={!status.MarketDisabled}
        aria-label="Toggle market data parsing"
      >
        <span class="box">{#if !status.MarketDisabled}<svg viewBox="0 0 9 9" width="9" height="9"><path d="M1 4.5 3.6 7 8 1.6" fill="none" stroke="var(--check-mark)" stroke-width="1.8" /></svg>{/if}</span>
        Parse market data
      </button>

      <button
        class="xp-check"
        onclick={toggleUpload}
        role="checkbox"
        aria-checked={!status.UploadDisabled}
        aria-label="Toggle sending data to servers"
      >
        <span class="box">{#if !status.UploadDisabled}<svg viewBox="0 0 9 9" width="9" height="9"><path d="M1 4.5 3.6 7 8 1.6" fill="none" stroke="var(--check-mark)" stroke-width="1.8" /></svg>{/if}</span>
        Send data to servers
      </button>
    </div>
  </section>

  {#if status.DriverWarning}
    <section class="panel">
      <header>Warning</header>
      <div class="panel-body">
        <div class="balloon">
          <span class="balloon-text">{status.DriverWarning}</span>
          {#if status.DriverHelpURL}
            <a href={status.DriverHelpURL} onclick={openDriverHelp}>Get Npcap</a>
          {/if}
        </div>
      </div>
    </section>
  {/if}

  <section class="panel">
    <header>Uploads</header>
    <div class="panel-body"><CountersPanel /></div>
  </section>

  {#if status.UpdateAvailable}
    <section class="panel">
      <header>Update</header>
      <div class="panel-body">
        <div class="balloon">
          <span class="balloon-text">Version <b>{status.UpdateAvailable}</b> is available. It will be installed automatically.</span>
        </div>
      </div>
    </section>
  {/if}

  <div
    class="resize-handle"
    class:active={resizing}
    onpointerdown={startResize}
    role="separator"
    aria-orientation="vertical"
    aria-label="Resize sidebar"
  ></div>
</aside>

<style>
  .sidebar {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 10px;
    flex-shrink: 0;
    padding: 10px 8px;
    background: linear-gradient(to bottom, var(--sidebar-top), var(--sidebar-bottom));
    border-right: 1px solid var(--sidebar-edge);
    overflow-y: auto;
    overflow-x: hidden;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--brand-text);
    min-width: 0;
  }
  .brand img {
    flex-shrink: 0;
  }
  .name {
    font-family: var(--font-head);
    font-size: 15px;
    font-weight: bold;
    text-shadow: var(--brand-shadow);
    letter-spacing: var(--head-spacing);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .panel {
    border-radius: var(--panel-radius);
    background: var(--panel-bg);
    border: 1px solid var(--panel-border);
    overflow: hidden;
  }
  .panel header {
    padding: 4px 10px;
    background: linear-gradient(to right, var(--panel-head-from), var(--panel-head-to));
    color: var(--panel-head-text);
    font-family: var(--font-head);
    font-weight: bold;
    text-transform: var(--head-transform);
    letter-spacing: var(--head-spacing);
    border-bottom: 1px solid var(--panel-border);
  }
  .panel-body {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 8px 10px 10px;
    color: var(--panel-text);
  }
  .panel-body :global(td),
  .panel-body :global(span),
  .panel-body :global(b) {
    color: inherit;
  }
  .row {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--panel-text);
  }
  .kv {
    display: flex;
    justify-content: space-between;
    color: var(--panel-text);
  }

  .led {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    border: 1px solid var(--led-border);
    background: radial-gradient(circle at 35% 35%, var(--led-off-hi), var(--led-off-lo));
    flex-shrink: 0;
  }
  .led.on {
    background: radial-gradient(circle at 35% 35%, var(--led-ok-hi), var(--ok));
    box-shadow: 0 0 5px color-mix(in srgb, var(--ok) 60%, transparent);
  }
  .led.err {
    background: radial-gradient(circle at 35% 35%, var(--led-bad-hi), var(--bad));
    box-shadow: 0 0 5px color-mix(in srgb, var(--bad) 60%, transparent);
  }
  .led.warn {
    background: radial-gradient(circle at 35% 35%, var(--led-warn-hi), var(--warn));
    box-shadow: 0 0 5px color-mix(in srgb, var(--warn) 60%, transparent);
  }

  /* checkbox */
  .xp-check {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 2px;
    padding: 0;
    border: none;
    background: none;
    font: inherit;
    color: var(--panel-text);
    cursor: pointer;
    text-align: left;
  }
  .xp-check .box {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 13px;
    height: 13px;
    box-sizing: border-box;
    border: 1px solid var(--check-border);
    border-radius: 2px;
    background: var(--check-bg);
  }
  .xp-check:hover .box {
    border-color: var(--check-hover);
  }
  .xp-check:focus-visible .box {
    outline: 1px solid var(--check-hover);
    outline-offset: 1px;
  }

  /* notice balloon */
  .balloon {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 4px 6px;
    background: var(--balloon-bg);
    border: 1px solid var(--balloon-border);
    color: var(--balloon-text);
  }
  .balloon .balloon-text,
  .balloon :global(b) {
    color: var(--balloon-text);
  }

  .resize-handle {
    position: absolute;
    top: 0;
    right: 0;
    width: 4px;
    height: 100%;
    cursor: col-resize;
    touch-action: none;
  }
  .resize-handle:hover,
  .resize-handle.active {
    background: var(--resize-hover);
  }
</style>
