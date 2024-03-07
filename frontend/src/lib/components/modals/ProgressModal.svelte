<script context="module" lang="ts">
  export const supportedProgressTypes = [
    '__select_install__',
    '__select_profile__',
    '__toggle_mods__',
    '__update__',
    '__import_profile__',
  ];
</script>

<script lang="ts">
  import { ProgressBar } from '@skeletonlabs/skeleton';
  import { getTranslate } from '@tolgee/svelte';

  import { getModalStore } from '$lib/skeletonExtensions';
  import { progress, selectedInstallMetadata, selectedProfile } from '$lib/store/ficsitCLIStore';

  // Skeleton passes the parent prop to the modal component, and we would get a warning if the prop is not present here
  export let parent: { onClose: () => void };

  // Just so that it's not unused
  $: parent;

  const modalStore = getModalStore();
  const { t } = getTranslate();

  $: if(!$progress) {
    // We cannot use parent.onClose because we might not be the top modal
    // Also this can get triggered multiple times for some reason,
    // which would cause an error in skeleton, so the modal would not actually be closed
    close();
  }

  let closed = false;

  function close() {
    if (closed) {
      return;
    }
    closed = true;
    modalStore.close('progress');
  }

  let title = '';

  $: title = (() => {
    switch ($progress?.item) {
      case '__select_install__':
        return `${$t('progress.title.select-install')} ${$selectedInstallMetadata?.info?.branch} (${$selectedInstallMetadata?.info?.launcher}) - CL${$selectedInstallMetadata?.info?.version}`;
      case '__select_profile__':
        return `${$t('progress.title.select-profile')} ${$selectedProfile}`;
      case '__toggle_mods__':
        return $t('progress.title.toggle-mods');
      case '__update__':
        return $t('progress.title.update');
      case '__import_profile__':
        return `${$t('progress.title.import-profile')} ${$selectedProfile}`;
    }
    return '';
  })();
</script>

<div style="max-height: calc(100vh - 3rem); max-width: calc(100vw - 3rem);" class="w-[48rem] card flex flex-col gap-2">
  <header class="card-header font-bold text-2xl text-center">
    {title}
  </header>
  <section class="p-4">
    {#if $progress}
      <p>{$t('progress.message.' + $progress.message, $progress.message)}</p>
      <ProgressBar
        class="h-4 w-full"
        max={1}
        meter="bg-primary-600"
        value={$progress.progress === -1 ? undefined : $progress.progress}/>
    {/if}
  </section>
</div>
