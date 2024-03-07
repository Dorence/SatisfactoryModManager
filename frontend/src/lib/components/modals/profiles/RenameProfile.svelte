<script lang="ts">
  import { getTranslate } from '@tolgee/svelte';

  import { newProfileName } from './renameProfile';

  import { RenameProfile } from '$lib/generated/wailsjs/go/ficsitcli/ficsitCLI';
  import { profiles } from '$lib/store/ficsitCLIStore';
  import { error } from '$lib/store/generalStore';

  const { t } = getTranslate();

  export let parent: { onClose: () => void };

  export let profile: string;

  $: newProfileNameExists = $profiles.includes($newProfileName);

  async function finishRenameProfile() {
    try {
      await RenameProfile(profile, $newProfileName);
      $newProfileName = '';
      parent.onClose();
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
    {$t('profile-modal.rename-profile')}
  </header>
  <section class="p-4 grow space-y-2">
    <label class="label w-full">
      <span>{$t('profile-modal.old-profile-name')}</span>
      <input
        class="input px-4 py-2"
        placeholder={$t('profile-modal.placeholder-old')}
        readonly
        type="text"
        value={profile}/>
    </label>
    <label class="label w-full">
      <span>{$t('profile-modal.new-profile-name')}</span>
      <input
        class="input px-4 py-2"
        class:input-error={newProfileNameExists}
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
      disabled={!$newProfileName || newProfileNameExists}
      on:click={finishRenameProfile}>
      {$t('rename')}
    </button>
  </footer>
</div>
