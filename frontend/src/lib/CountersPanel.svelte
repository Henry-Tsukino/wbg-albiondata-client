<script>
  import { onDestroy } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/Henry-Tsukino/wbg-albiondata-client/internal/dashboard/index.js';

  const labels = {
    'marketorders.ingest': 'Market orders',
    'markethistories.ingest': 'Market histories',
    'goldprices.ingest': 'Gold prices',
    'mapdata.ingest': 'Map data',
    'festivities.ingest': 'Festivities',
    'banditevent.ingest': 'Bandit events',
  };

  let counts = $state({});

  DashboardService.GetUploadCounts().then((c) => (counts = c));

  const unlisten = Events.On('counters:snapshot', (evt) => {
    counts = evt.data;
  });

  onDestroy(unlisten);
</script>

<table class="counters">
  <tbody>
    {#each Object.entries(labels) as [topic, label]}
      <tr>
        <td class="label">{label}</td>
        <td class="value">{counts[topic] ?? 0}</td>
      </tr>
    {/each}
  </tbody>
</table>

<style>
  .counters {
    width: 100%;
    border-collapse: collapse;
  }
  td {
    padding: 2px 0;
  }
  .value {
    text-align: right;
    font-family: var(--font-mono);
    color: var(--accent-text);
    font-weight: bold;
    font-variant-numeric: tabular-nums;
  }
</style>
