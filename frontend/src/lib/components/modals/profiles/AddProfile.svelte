<script lang="ts">
  import { getTranslate } from '@tolgee/svelte';

  import { newProfileName } from './addProfile';

  import { AddProfile } from '$lib/generated/wailsjs/go/ficsitcli/ficsitCLI';
  import { getModalStore } from '$lib/skeletonExtensions';
  import { selectedProfile } from '$lib/store/ficsitCLIStore';
  import { error } from '$lib/store/generalStore';

  export let parent: { onClose: () => void };

  const modalStore = getModalStore();
  const { t } = getTranslate();

  async function finishAddProfile() {
    try {
      await AddProfile($newProfileName);
      await selectedProfile.asyncSet($newProfileName);

      $newProfileName = '';

      modalStore.close('addProfile');
    } catch(e) {
      if (e instanceof Error) {
        $error = e.message;
      } else if (typeof e === 'string') {
        $error = e;
      } else {
        $error = $t('error.unknown-error');
      }
    }
  }
</script>

<div style="max-height: calc(100vh - 3rem); max-width: calc(100vw - 3rem);" class="w-[40rem] card flex flex-col gap-2">
  <header class="card-header font-bold text-2xl text-center">
    {$t('profile-modal.add-profile')}
  </header>
  <section class="p-4 grow">
    <label class="label w-full">
      <span>{$t('profile-modal.profile-name')}</span>
      <input
        class="input px-4 py-2"
        placeholder={$t('profile-modal.placeholder-new')}
        type="text"
        bind:value={$newProfileName}/>
    </label>
  </section>
  <footer class="card-footer">
    <button
      class="btn"
      on:click={parent.onClose}>
      {$t('cancel')}
    </button>
    <button
      class="btn text-primary-600"
      disabled={!newProfileName}
      on:click={finishAddProfile}>
      {$t('add')}
    </button>
  </footer>
</div>
