<script>
  import Sidebar from './lib/Sidebar.svelte';
  import LogTail from './lib/LogTail.svelte';
  import Footer from './lib/Footer.svelte';
</script>

<main>
  <div class="body">
    <Sidebar />
    <div class="content">
      <fieldset class="group-box">
        <legend>Activity log</legend>
        <LogTail />
      </fieldset>
    </div>
  </div>
  <Footer />
</main>

<style>
  main {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }
  .body {
    display: flex;
    flex: 1;
    min-height: 0;
  }
  .content {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    padding: 6px 8px 4px 6px;
  }
  .group-box {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    margin: 0;
    padding: 2px 4px 4px;
    border: 1px solid var(--border);
    border-radius: 4px;
  }
  .group-box legend {
    padding: 0 4px;
    color: var(--legend);
    font-family: var(--font-head);
    text-transform: var(--head-transform);
    letter-spacing: var(--head-spacing);
  }
</style>
