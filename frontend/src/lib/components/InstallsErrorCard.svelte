<script>
  import { getTranslate } from '@tolgee/svelte';

  import { installs, invalidInstalls } from '$lib/store/ficsitCLIStore';
  import { GenerateDebugInfo } from '$wailsjs/go/app/app';

  const { t } = getTranslate();
  const installsInit = installs.isInit;
  const invalidInstallsInit = invalidInstalls.isInit;
  
  let noInstallsError = false;
  $: if($installsInit && $invalidInstallsInit && $installs.length === 0) {
    if($invalidInstalls.length === 0) {
      noInstallsError = true;
    }
  }
</script>

<div class="card my-auto mr-4 select-text">
  <header class="card-header font-bold text-2xl text-center">
    {#if noInstallsError}
      {$t('error.installs.no-installs-found')}
    {:else}
      {$t('error.installs.invalid-installs-found', { n: $invalidInstalls.length })}
    {/if}
  </header>
  <section class="p-4">
    <p class="text-base text-center">
      {$t('error.help-using-mods.0')}<a class="text-primary-600 underline" href="https://discord.gg/xkVJ73E">$t('error.help-using-mods.1')}</a>$t('error.help-using-mods.2')}
    </p>
  </section>
  <footer class="card-footer">
    <button class="btn text-primary-600 w-full" on:click={GenerateDebugInfo}>{$t('generate-debug-info')}</button>
  </footer>
</div>