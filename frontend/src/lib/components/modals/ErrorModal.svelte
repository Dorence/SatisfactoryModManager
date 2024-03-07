<script lang="ts">
  import { getTranslate } from '@tolgee/svelte';

  import { GenerateDebugInfo } from '$lib/generated/wailsjs/go/app/app';

  const { t } = getTranslate();

  export let parent: { onClose: () => void };

  export let error: string;
</script>

<div style="max-height: calc(100vh - 3rem); max-width: calc(100vw - 3rem);" class="w-[48rem] card flex flex-col gap-6">
  <header class="card-header font-bold text-2xl text-center">
    {$t('error.something-wrong')}
  </header>
  <section class="px-4 overflow-y-auto">
    <p>{error}</p>
  </section>
  <section class="px-4">
    <p>{$t('error.help-using-mods.0')}<a class="text-primary-600 underline" href="https://discord.gg/xkVJ73E">$t('error.help-using-mods.1')}</a>$t('error.help-using-mods.2')}</p>
  </section>
  <footer class="card-footer">
    <button
      class="btn text-primary-600 variant-ringed"
      on:click={GenerateDebugInfo}>
      {$t('generate-debug-info')}
    </button>
    <button
      class="btn"
      on:click={parent.onClose}>
      {$t('close')}
    </button>
  </footer>
</div>
