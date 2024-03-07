<script lang="ts">
  import { getTranslate } from '@tolgee/svelte';

  import { UpdateAndRestart } from '$lib/generated/wailsjs/go/autoupdate/autoUpdate';
  import { smmUpdate } from '$lib/store/smmUpdateStore';

  const { t } = getTranslate();

  export let parent: { onClose: () => void };
</script>

<div style="max-height: calc(100vh - 3rem); max-width: calc(100vw - 3rem);" class="w-[48rem] card flex flex-col gap-2">
  <header class="card-header font-bold text-2xl text-center">
    {$t('updates.smm-ready')} - {$smmUpdate?.newVersion}
  </header>
  <section class="p-4 flex-auto">
    <span class="text-base">{$t('updates.update-ready-to-install')}</span>
  </section>
  <footer class="card-footer">
    <button
      class="btn"
      on:click={parent.onClose}>
      {$t('cancel')}
    </button>
    <button
      class="btn text-primary-600"
      on:click={() => UpdateAndRestart()}>
      <span>{$t('updates.update-and-restart')}</span>
    </button>
  </footer>
</div>
