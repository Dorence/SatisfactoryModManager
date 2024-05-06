<script lang="ts">
  import { getTranslate } from '@tolgee/svelte';

  import { CloseAndRestart } from '$lib/generated/wailsjs/go/app/app';
  import { progress } from '$lib/store/ficsitCLIStore';
  import { proxy } from '$lib/store/settingsStore';

  const { t } = getTranslate();

  export let parent: { onClose: () => void };
  
  let proxyURL = $proxy;
  let saving = false;

  $: canChange = !$progress && !saving;
  $: canSave = proxyURL !== $proxy && canChange;

  async function setProxy() {
    saving = true;
    $proxy = proxyURL;
    setTimeout(() => {
      CloseAndRestart();
    }, 1000);
  }
</script>

<div style="max-height: calc(100vh - 3rem); max-width: calc(100vw - 3rem);" class="w-[60rem] card flex flex-col gap-2">
  <header class="card-header font-bold text-2xl text-center">
    {$t('smm-settings.proxy.title')}
  </header>
  <section class="p-4 grow">
    <label class="label">
      <span>{$t('smm-settings.proxy.proxy')}</span>
      <div class="flex items-baseline">
        <input
          class="input px-4 py-2 grow"
          type="text"
          bind:value={proxyURL}
        />
        <button
          class="btn shrink-0 text-primary-600"
          disabled={!canChange}
          on:click={() => { proxyURL = ''; setProxy(); }}>
          <span>{$t('smm-settings.proxy.remove')}</span>
          <div class="grow" />
        </button>
      </div>
    </label>
  </section>
  <footer class="card-footer">
    <button
      class="btn text-primary-600 variant-ringed"
      on:click={parent.onClose}>
      <span>{$t('close')}</span>
    </button>
    <button
      class="btn shrink-0 text-primary-600"
      disabled={!canSave}
      on:click={() => setProxy()}>
      <span>{$t('smm-settings.proxy.save-and-restart')}</span>
    </button>
  </footer>
</div>

