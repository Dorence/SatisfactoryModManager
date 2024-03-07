<script lang="ts">
  import { getTranslate } from '@tolgee/svelte';
  import { getContextClient, queryStore } from '@urql/svelte';

  import { GetModSummaryDocument } from '$lib/generated';
  import { addQueuedModAction, queuedMods } from '$lib/store/actionQueue';
  import { manifestMods } from '$lib/store/ficsitCLIStore';
  import { error } from '$lib/store/generalStore';
  import { offline } from '$lib/store/settingsStore';
  import { InstallMod, InstallModVersion } from '$wailsjs/go/ficsitcli/ficsitCLI';

  export let parent: { onClose: () => void };

  export let modReference: string;
  export let version: string | undefined;

  const { t } = getTranslate();
  const client = getContextClient();

  $: modQuery = queryStore(
    {
      query: GetModSummaryDocument,
      client,
      pause: !!$offline,
      variables: {
        modReference,
      },
    },
  );

  $: mod = $modQuery.fetching ? null : $modQuery.data?.mod;
  
  $: queued = $queuedMods.some((q) => q.mod === modReference);
  $: isInstalled = !!modReference && modReference in $manifestMods;

  function install() {
    const action = async () => (version ? InstallModVersion(modReference, version) : InstallMod(modReference)).catch((e) => $error = e);
    const actionName = 'install';
    addQueuedModAction(
      modReference,
      actionName,
      action,
    );
    parent.onClose();
  }

  $: renderedLogo = mod?.logo || 'https://ficsit.app/images/no_image.webp';
</script>

<div style="max-height: calc(100vh - 3rem); max-width: calc(100vw - 3rem);" class="w-[48rem] card flex flex-col gap-2">
  <header class="card-header font-bold text-2xl text-center">
    {$t('mod-item.install-mod')}
  </header>
  <section class="p-4 overflow-y-auto">
    {#if mod}
      <div class="flex">
        <div class="grow">
          <p>{mod.name}</p>
          {#if version}
            <p>{$t('mod-item.version.version')} {version}</p>
          {:else}
            <p>{$t('mod-item.version.latest')}</p>
          {/if}
          <p>{mod.short_description}</p>
        </div>
        <img class="logo h-24 w-24 mx-2" alt="{mod.name} Logo" src={renderedLogo} />
      </div>
    {:else if $modQuery.fetching}
      <p>{$t('loading')}</p>
    {:else if $modQuery.error}
      <p>{$t('mod-item.load-details-error')}</p>
    {/if}
  </section>
  <footer class="card-footer">
    <button
      class="btn text-primary-600 variant-ringed"
      disabled={isInstalled || queued}
      on:click={install}>
      {#if queued}
        {$t('mod-item.queued.in-queue')}
      {:else if isInstalled}
        {$t('mod-item.already-installed')}
      {:else}
        {$t('install')}
      {/if}
    </button>
    <button class="btn" on:click={parent.onClose}>{$t('cancel')}</button>
  </footer>
</div>
