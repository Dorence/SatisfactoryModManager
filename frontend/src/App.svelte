<script lang="ts">
  import './_global.postcss';
  import { arrow, autoUpdate, computePosition, flip, offset, shift, size } from '@floating-ui/dom';
  import { Modal, initializeStores, storePopup } from '@skeletonlabs/skeleton';
  import { DevTools, FormatSimple, Tolgee, TolgeeProvider } from '@tolgee/svelte';
  import { setContextClient } from '@urql/svelte';

  import enCommon from './i18n/common/en.json';
  import zhHansCommon from './i18n/common/zh-Hans.json';

  import InstallsErrorCard from '$lib/components/InstallsErrorCard.svelte';
  import TitleBar from '$lib/components/TitleBar.svelte';
  import LeftBar from '$lib/components/left-bar/LeftBar.svelte';
  import ModDetails from '$lib/components/mod-details/ModDetails.svelte';
  import ErrorModal from '$lib/components/modals/ErrorModal.svelte';
  import ExternalInstallMod from '$lib/components/modals/ExternalInstallMod.svelte';
  import { supportedProgressTypes } from '$lib/components/modals/ProgressModal.svelte';
  import { modalRegistry } from '$lib/components/modals/modalsRegistry';
  import ImportProfile from '$lib/components/modals/profiles/ImportProfile.svelte';
  import ModsList from '$lib/components/mods-list/ModsList.svelte';
  import { initializeGraphQLClient } from '$lib/core/graphql';
  import { getModalStore, initializeModalStore } from '$lib/skeletonExtensions';
  import { installs, invalidInstalls, progress } from '$lib/store/ficsitCLIStore';
  import { error, expandedMod, siteURL } from '$lib/store/generalStore';
  import { konami, language } from '$lib/store/settingsStore';
  import { ExpandMod, UnexpandMod } from '$wailsjs/go/app/app';
  import { Environment, EventsOn } from '$wailsjs/runtime';

  initializeStores();
  initializeModalStore();

  storePopup.set({ computePosition, autoUpdate, offset, shift, flip, arrow, size });

  let frameless = false;
  Environment().then((env) => {
    if (env.buildType !== 'dev') {
      document.addEventListener('contextmenu', (event) => event.preventDefault());
    }
    if (env.platform === 'windows') {
      frameless = true;
    }
  });

  export let apiEndpointURL!: string;
  export let siteEndpointURL!: string;
  
  $: $siteURL = siteEndpointURL;

  setContextClient(initializeGraphQLClient(apiEndpointURL));

  const tolgee = Tolgee()
    .use(DevTools())
    .use(FormatSimple())
    .init({
      language: $language,
      fallbackLanguage: 'en',

      // for development
      apiUrl: import.meta.env.VITE_TOLGEE_API_URL,
      apiKey: import.meta.env.VITE_TOLGEE_API_KEY,

      // for production
      staticData: {
        en: enCommon,
        'zh-Hans':zhHansCommon,
      },
    });

  language.subscribe((lang) => {
    if(lang !== tolgee.getLanguage()) {
      tolgee.changeLanguage(lang);
    }
  });

  let windowExpanded = false;

  $: if ($expandedMod) {
    ExpandMod().then(() => {
      setTimeout(() => {
        windowExpanded = true;
      }, 100);
    });
  } else {
    windowExpanded = false;
    setTimeout(() => {
      UnexpandMod();
    }, 100);
  }

  $: pendingExpand = $expandedMod && !windowExpanded;

  let invalidInstallsError = false;
  let focusOnEntry: HTMLSpanElement;

  const installsInit = installs.isInit;
  const invalidInstallsInit = invalidInstalls.isInit;

  $: if($installsInit && $invalidInstallsInit && $installs.length === 0) {
    invalidInstallsError = true;
  }

  const modalStore = getModalStore();

  $: if($progress && supportedProgressTypes.includes($progress.item)) {
    modalStore.triggerUnique({
      type: 'component',
      component: 'progress',
      meta: {
        persistent: true,
      },
    }, true);
  }

  $: if($error) {
    modalStore.trigger({
      type: 'component',
      component: {
        ref: ErrorModal,
        props: {
          error: $error,
        },
      },
    }, true);
    $error = null;
  }

  EventsOn('externalInstallMod', (modReference: string, version: string) => {
    if (!modReference) return;
    modalStore.trigger({
      type: 'component',
      component: {
        ref: ExternalInstallMod,
        props: {
          modReference,
          version,
        },
      },
    });
  });

  EventsOn('externalImportProfile', async (path: string) => {
    if (!path) return;
    modalStore.trigger({
      type: 'component',
      component: {
        ref: ImportProfile,
        props: {
          filepath: path,
        },
      },
    });
  });

  $: isPersistentModal = $modalStore.length > 0 && $modalStore[0].meta?.persistent;

  function modalMouseDown(event: MouseEvent) {
    if (!isPersistentModal) return;
    if (!(event.target instanceof Element)) return;
    const classList = event.target.classList;
    if (classList.contains('modal-backdrop') || classList.contains('modal-transition')) {
      event.stopImmediatePropagation();
    }
  }

  function modalKeyDown(event: KeyboardEvent) {
    if (!isPersistentModal) return;
    if (event.key === 'Escape') {
      event.stopImmediatePropagation();
    }
  }
  
  const code = [38, 38, 40, 40, 37, 39, 37, 39, 66, 65];
  const keyQueue: number[] = [];
  window.addEventListener('keydown', (event) => {
    keyQueue.push(event.keyCode);
    if (keyQueue.length > code.length) {
      keyQueue.shift();
    }
    if (keyQueue.length === code.length && keyQueue.every((val, idx) => code[idx] === val)) {
      $konami = !$konami;
    }
  });
</script>

<TolgeeProvider tolgee={tolgee}>
  <div slot="fallback">Loading translations...</div>
  <div class="flex flex-col h-screen w-screen select-none">
    {#if frameless}
      <TitleBar />
    {/if}
    <div class="flex grow h-0">
      <LeftBar />
      <div class="flex flex-auto @container/mod-list-wrapper z-[1]">
        <div class="{$expandedMod && !pendingExpand ? 'w-2/5 hidden @3xl/mod-list-wrapper:block @3xl/mod-list-wrapper:flex-auto' : 'w-full'}" class:max-w-[42.5rem]={!!$expandedMod}>
          <ModsList hideMods={invalidInstallsError} on:expandedMod={() => { focusOnEntry.focus(); }}>
            <InstallsErrorCard />
          </ModsList>
        </div>
        <div class="w-3/5" class:grow={!pendingExpand} class:hidden={!$expandedMod}>
          <ModDetails bind:focusOnEntry/>
        </div>
      </div>
    </div>
  </div>
  <Modal components={modalRegistry} />
</TolgeeProvider>

<!--
  skeleton modals don't provide a way to make them persistent (i.e. ignore mouse clicks outside and escape key)
  but we can capture the events and stop them if the modal has the persistent meta flag set, and the event would have closed the modal
-->
<svelte:window
  on:keydown|capture|nonpassive={modalKeyDown}
  on:mousedown|capture|nonpassive={modalMouseDown} />
