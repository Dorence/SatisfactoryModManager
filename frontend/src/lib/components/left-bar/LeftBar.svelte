<script lang="ts">
  import { mdiAlert, mdiCheckCircle, mdiCloseCircle, mdiDownload, mdiFolderOpen, mdiHelpCircle, mdiLoading, mdiPencil, mdiPlusCircle, mdiServerNetwork, mdiTrashCan, mdiUpload, mdiWeb } from '@mdi/js';
  import { type PopupSettings, popup } from '@skeletonlabs/skeleton';
  import { getTranslate } from '@tolgee/svelte';
  import _ from 'lodash';
  import { siDiscord, siGithub } from 'simple-icons/icons';

  import Tooltip from '../Tooltip.svelte';
  import DeleteProfile from '../modals/profiles/DeleteProfile.svelte';
  import RenameProfile from '../modals/profiles/RenameProfile.svelte';

  import LaunchButton from './LaunchButton.svelte';
  import Settings from './Settings.svelte';
  import Updates from './Updates.svelte';

  import SvgIcon from '$lib/components/SVGIcon.svelte';
  import Select from '$lib/components/Select.svelte';
  import { getModalStore } from '$lib/skeletonExtensions';
  import { canChangeInstall, canModify, installs, installsMetadata, modsEnabled, profiles, selectedInstall, selectedProfile } from '$lib/store/ficsitCLIStore';
  import { error, siteURL } from '$lib/store/generalStore';
  import { OpenExternal } from '$wailsjs/go/app/app';
  import { ExportCurrentProfile } from '$wailsjs/go/ficsitcli/ficsitCLI';
  import { common, ficsitcli } from '$wailsjs/go/models';
  import { BrowserOpenURL } from '$wailsjs/runtime/runtime';

  const modalStore = getModalStore();
  const { t } = getTranslate();

  const selectedInstallPathInit = selectedInstall.isInit;
  const selectedProfileInit = selectedProfile.isInit;

  async function installSelectChanged({ detail: value }: CustomEvent<string>) {
    if (!value) {
      return;
    }
    if (!$selectedInstallPathInit) {
      return;
    }
    try {
      await selectedInstall.asyncSet(value);
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

  async function profileSelectChanged({ detail: value }: CustomEvent<string>) {
    if (!value) {
      return;
    }
    if (!$selectedProfileInit) {
      return;
    }
    try {
      await selectedProfile.asyncSet(value);
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

  async function setModsEnabled(enabled: boolean) {
    if (!$selectedInstallPathInit) {
      return;
    }
    try {
      await modsEnabled.asyncSet(enabled);
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

  async function exportCurrentProfile() {
    try {
      await ExportCurrentProfile();
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

  function installOptionPopupId(install: string) {
    return `install-path-${install.replace(/[^a-zA-Z0-9]/g, '-')}`;
  }

  $: installOptionPopups = $installs.map((i) => [i, {
    event: 'hover',
    target: installOptionPopupId(i),
    middleware: {
      offset: 4,
    },
    placement: 'right',
  } as PopupSettings]).reduce((acc, [k, v]) => ({ ...acc, [k as string]: v as PopupSettings }), {} as Record<string, PopupSettings>);
</script>

<div class="flex flex-col h-full p-4 space-y-4 h-md:space-y-8 left-bar w-[22rem] min-w-[22rem] ">
  <div class="flex flex-col flex-auto h-full w-full space-y-4 h-md:space-y-8 overflow-y-auto">
    <div class="flex flex-col gap-2">
      <span class="pl-4 sticky top-0 z-[1] bg-surface-50-900-token">{$t('left-bar.game-version')}</span>
      <Select
        name="installsCombobox"
        class="w-full h-8"
        buttonClass="bg-surface-200-700-token px-4 text-sm"
        disabled={!$canChangeInstall}
        itemActiveClass="!bg-surface-300/20"
        itemClass="bg-surface-50-900-token"
        items={_.orderBy($installs)}
        value={$selectedInstall ?? ''}
        on:change={installSelectChanged}
      >
        <svelte:fragment slot="item" let:item>
          <span>
            {#if $installsMetadata[item]?.state === ficsitcli.InstallState.VALID}
              {$installsMetadata[item].info?.branch}{$installsMetadata[item].info?.type !== common.InstallType.WINDOWS ? ' - DS' : ''}
              ({$installsMetadata[item]?.info?.launcher})
            {:else if $installsMetadata[item]?.state === ficsitcli.InstallState.LOADING}
              {$t('loading')}
            {:else if $installsMetadata[item]?.state === ficsitcli.InstallState.INVALID}
              {$t('invalid')}
            {:else}
              {$t('unknown')}
            {/if}
          </span>
        </svelte:fragment>
        <svelte:fragment slot="itemTrail" let:item>
          <Tooltip popupId={installOptionPopupId(item)}>
            <div class="flex flex-col">
              <span>{item}</span>
              {#if $installsMetadata[item]?.state === ficsitcli.InstallState.VALID}
                <!-- nothing extra -->
              {:else if $installsMetadata[item]?.state === ficsitcli.InstallState.LOADING}
                <span>{$t('left-bar.install-status.loading')}</span>
              {:else if $installsMetadata[item]?.state === ficsitcli.InstallState.INVALID}
                <span>{$t('left-bar.install-status.invalid')}</span>
              {:else}
                <span>{$t('left-bar.install-status.unknown')}</span>
              {/if}
            </div>
          </Tooltip>
          <button
            class="!w-5 !h-5"
            on:click|stopPropagation={() => OpenExternal(item)}
            use:popup={installOptionPopups[item]} >
            {#if $installsMetadata[item]?.state === ficsitcli.InstallState.VALID}
              <SvgIcon class="!w-full !h-full" icon={mdiFolderOpen}/>
            {:else if $installsMetadata[item]?.state === ficsitcli.InstallState.LOADING}
              <SvgIcon class="!w-full !h-full animate-spin text-primary-600" icon={mdiLoading}/>
            {:else if $installsMetadata[item]?.state === ficsitcli.InstallState.INVALID}
              <SvgIcon class="!w-full !h-full" icon={mdiAlert}/>
            {:else}
              <SvgIcon class="!w-full !h-full" icon={mdiAlert}/>
            {/if}
          </button>
        </svelte:fragment>
        <svelte:fragment slot="selected" let:item>
          {#if $installsMetadata[item]?.state === ficsitcli.InstallState.VALID}
            {$installsMetadata[item].info?.branch}{$installsMetadata[item].info?.type !== common.InstallType.WINDOWS ? ' - DS' : ''}
            ({$installsMetadata[item]?.info?.launcher})
          {:else if $installsMetadata[item]?.state === ficsitcli.InstallState.LOADING}
            {$t('loading')}
          {:else if $installsMetadata[item]?.state === ficsitcli.InstallState.INVALID}
            {$t('invalid')}
          {:else}
            {$t('unknown')}
          {/if}
        </svelte:fragment>
      </Select>

      <div class="flex w-full">
        <div class="btn-group bg-surface-200-700-token w-full text-xl">
          <button
            class="w-1/2 !btn-sm !px-4"
            class:!bg-error-900={!$modsEnabled}
            disabled={!$canModify}
            on:click={() => setModsEnabled(false)}
          >
            <span>
              {$t('left-bar.mods-off')}
            </span>
            <div class="grow"/>
            <SvgIcon
              class="h-5 w-5"
              icon={mdiCloseCircle} />
          </button>
          <button
            class="w-1/2 !btn-sm !px-4"
            class:!bg-primary-900={$modsEnabled}
            disabled={!$canModify}
            on:click={() => setModsEnabled(true)}>
            <span>
              {$t('left-bar.mods-on')}
            </span>
            <div class="grow"/>
            <SvgIcon
              class="h-5 w-5"
              icon={mdiCheckCircle} />
          </button>
        </div>
      </div>
    </div>
    <div class="flex flex-col gap-2">
      <span class="pl-4 sticky top-0 z-[1] bg-surface-50-900-token">{$t('left-bar.profile')}</span>

      <Select
        name="profileCombobox"
        class="w-full h-8"
        buttonClass="bg-surface-200-700-token px-4 text-sm"
        disabled={!$canModify}
        itemActiveClass="!bg-surface-300/20"
        itemClass="bg-surface-50-900-token"
        items={_.orderBy($profiles)}
        value={$selectedProfile ?? ''}
        on:change={profileSelectChanged}
      >
        <svelte:fragment slot="itemTrail" let:item>
          <button
            disabled={!$canModify}
            on:click|stopPropagation={() => modalStore.trigger({ type:'component', component: { ref: RenameProfile, props: { profile: item } } })}
          >
            <SvgIcon class="!w-5 !h-5 text-warning-500" icon={mdiPencil}/>
          </button>
          <button
            disabled={!$canModify || $profiles.length === 1}
            on:click|stopPropagation={() => modalStore.trigger({ type:'component', component: { ref: DeleteProfile, props: { profile: item } } })}
          >
            <SvgIcon class="!w-5 !h-5 text-error-700" icon={mdiTrashCan}/>
          </button>
        </svelte:fragment>
      </Select>

      <div class="flex w-full gap-1">
        <button
          class="btn w-1/3 bg-surface-200-700-token px-4 h-8 text-sm"
          disabled={!$canModify}
          on:click={() => modalStore.trigger({ type:'component', component: 'addProfile' })}>
          <span>
            {$t('add')}
          </span>
          <div class="grow"/>
          <SvgIcon
            class="h-5 w-5 text-primary-600"
            icon={mdiPlusCircle} />
        </button>
        <button
          class="btn w-1/3 bg-surface-200-700-token px-2 h-8 text-sm"
          disabled={!$canModify}
          on:click={() => modalStore.trigger({ type:'component', component: { ref: RenameProfile, props: { profile: $selectedProfile } } })}>
          <span>
            {$t('rename')}
          </span>
          <div class="grow"/>
          <SvgIcon
            class="h-5 w-5 text-warning-500"
            icon={mdiPencil} />
        </button>
        <button
          class="btn w-1/3 bg-surface-200-700-token px-3 h-8 text-sm"
          disabled={!$canModify || $profiles.length === 1}
          on:click={() => modalStore.trigger({ type:'component', component: { ref: DeleteProfile, props: { profile: $selectedProfile } } })}>
          <span>
            {$t('delete')}
          </span>
          <div class="grow"/>
          <SvgIcon
            class="h-5 w-5 text-error-700"
            icon={mdiTrashCan} />
        </button>
      </div>
      <div class="flex w-full gap-1">
        <button
          class="btn w-1/2 bg-surface-200-700-token px-4 h-8 text-sm"
          disabled={!$canModify}
          on:click={() => modalStore.trigger({ type: 'component', component: 'importProfile' })}
        >
          <span>
            {$t('import')}
          </span>
          <div class="grow"/>
          <SvgIcon
            class="h-5 w-5"
            icon={mdiDownload} />
        </button>
        <button
          class="btn w-1/2 bg-surface-200-700-token px-4 h-8 text-sm"
          disabled={!$canModify}
          on:click={() => exportCurrentProfile()}
        >
          <span>
            {$t('export')}
          </span>
          <div class="grow"/>
          <SvgIcon
            class="h-5 w-5"
            icon={mdiUpload} />
        </button>
      </div>
    </div>
    <div class="flex flex-col gap-2">
      <span class="pl-4 sticky top-0 z-[1] bg-surface-50-900-token">{$t('left-bar.updates')}</span>
      <Updates />
    </div>
    <div class="flex flex-col gap-2">
      <span class="pl-4 sticky top-0 z-[1] bg-surface-50-900-token">{$t('left-bar.other')}</span>
      <button
        class="btn px-4 h-8 w-full text-sm bg-surface-200-700-token"
        on:click={() => modalStore.trigger({ type: 'component', component: 'serverManager' })}>
        <span>{$t('left-bar.manage-servers')}</span>
        <div class="grow" />
        <SvgIcon
          class="h-5 w-5"
          icon={mdiServerNetwork} />
      </button>
      <Settings />
      <button
        class="btn w-full bg-surface-200-700-token px-4 h-8 text-sm"
        on:click={() => BrowserOpenURL('https://docs.ficsit.app/satisfactory-modding/latest/ForUsers/SatisfactoryModManager.html')}
      >
        <span>
          {$t('left-bar.help')}
        </span>
        <div class="grow"/>
        <SvgIcon
          class="h-5 w-5"
          icon={mdiHelpCircle} />
      </button>
    </div>
    <div class="flex flex-col gap-2">
      <span class="pl-4 sticky top-0 z-[1] bg-surface-50-900-token">{$t('left-bar.links.title')}</span>
      <button
        class="btn w-full bg-surface-200-700-token px-4 h-8 text-sm"
        on:click={() => BrowserOpenURL($siteURL)}>
        <span>
          {$t('left-bar.links.ficsit')}
        </span>
        <div class="grow" />
        <SvgIcon
          class="h-5 w-5"
          icon={mdiWeb} />
      </button>
      <button
        class="btn w-full bg-surface-200-700-token px-4 h-8 text-sm"
        on:click={() => BrowserOpenURL('https://discord.gg/xkVJ73E')}>
        <span>
          {$t('left-bar.links.discord')}
        </span>
        <div class="grow" />
        <SvgIcon
          class="h-5 w-5"
          icon={siDiscord.path} />
      </button>
      <button
        class="btn w-full bg-surface-200-700-token px-4 h-8 text-sm"
        on:click={() => BrowserOpenURL('https://github.com/satisfactorymodding/SatisfactoryModManager')} >
        <span>
          {$t('left-bar.links.github')}
        </span>
        <div class="grow" />
        <SvgIcon
          class="h-5 w-5"
          icon={siGithub.path} />
      </button>
    </div>
  </div>
  <LaunchButton />
</div>

<style>
  :global(.update-check) {
    animation: spin 4s linear infinite;
  }
  @keyframes spin {
    100% {
      transform: rotate(-360deg);
    }
  }
</style>
