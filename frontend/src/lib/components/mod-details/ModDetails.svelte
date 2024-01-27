<script lang="ts">
  import { getContextClient, queryStore } from '@urql/svelte';
  import Checkbox from '@smui/checkbox';
  import { mdiCheck, mdiChevronDown, mdiImport, mdiRocketLaunch, mdiTestTube, mdiWeb } from '@mdi/js';
  import Menu from '@smui/menu';
  import List, { Item, PrimaryText, SecondaryText, Separator, Text } from '@smui/list';
  import Dialog, { Content, Title } from '@smui/dialog';
  import { minVersion, valid, validRange, sort, coerce, SemVer, parse } from 'semver';
  import { popup, type PopupSettings } from '@skeletonlabs/skeleton';

  import Tooltip from '../Tooltip.svelte';

  import { CompatibilityState, GetModDetailsDocument, GetModReferenceDocument, type Version } from '$lib/generated';
  import { markdown } from '$lib/utils/markdown';
  import SvgIcon from '$lib/components/SVGIcon.svelte';
  import { bytesToAppropriate } from '$lib/utils/dataFormats';
  import { canModify, lockfileMods, manifestMods, progress , selectedInstallMetadata } from '$lib/store/ficsitCLIStore';
  import { error , expandedMod, siteURL } from '$lib/store/generalStore';
  import { search } from '$lib/store/modFiltersStore';
  import { InstallModVersion, OfflineGetMod } from '$wailsjs/go/ficsitcli/FicsitCLI';
  import { BrowserOpenURL } from '$wailsjs/runtime/runtime';
  import { getAuthor } from '$lib/utils/getModAuthor';
  import { getCompatibility, getVersionCompatibility, type CompatibilityWithSource } from '$lib/utils/modCompatibility';
  import { offline } from '$lib/store/settingsStore';
  import type { ficsitcli } from '$wailsjs/go/models';

  export let focusOnEntry: HTMLElement | undefined = undefined;

  const client = getContextClient();

  $: modQuery = queryStore(
    {
      query: GetModDetailsDocument,
      client,
      pause: !$expandedMod || !!$offline,
      variables: {
        modReference: $expandedMod ?? '',
      },
    },
  );

  interface OfflineMod {
    offline: true;
    mod_reference: string;
    name: string;
    authors: {
      role: 'creator';
      user: {
        username: string;
      }
    }[];
    logo?: string;
    versions: ficsitcli.ModVersion[];
  }
  
  let offlineMod: OfflineMod = {
    offline: true,
    mod_reference: '',
    name: '',
    versions: [],
    authors: [],
    logo: undefined,
  };

  $: {
    if($offline && $expandedMod) {
      OfflineGetMod($expandedMod).then((mod) => {
        offlineMod = {
          ...mod,
          authors: mod.authors.map((author) => ({
            role: 'creator',
            user: {
              username: author,
            },
          })),
          offline: true,
        };
      });
    }
  }

  $: mod = $offline ? offlineMod : ($modQuery.fetching ? null : $modQuery.data?.mod);

  $: actualLogo = (mod && 'offline' in mod) ? (mod?.logo ? `data:image/png;base64, ${mod?.logo}` : '/images/no_image.webp') : mod?.logo;
  $: renderedLogo = actualLogo || `${$siteURL}/images/no_image.webp`;
  $: descriptionRendered = (mod && 'full_description' in mod && mod?.full_description) ? markdown(mod.full_description) : undefined;
  $: author = getAuthor(mod);

  $: isInstalled = mod && mod.mod_reference in $manifestMods;
  $: isEnabled = mod && mod.mod_reference in $lockfileMods;
  $: isDependency = !isInstalled && isEnabled;
  $: inProgress = $progress?.item === mod?.mod_reference;

  $: size = mod ? bytesToAppropriate(mod.versions[0]?.size ?? 0) : undefined;

  $: latestVersion = mod?.versions?.length ? sort(mod.versions.map((v) => parse(v.version) ?? coerce(v.version)).filter((v) => !!v) as SemVer[]).reverse()[0] : 'N/A';
  $: installedVersion = (mod && $lockfileMods[mod.mod_reference]?.version) ?? 'Not installed';

  $: ficsitAppLink = `${$siteURL}/mod/${$expandedMod}`;

  let compatibility: CompatibilityWithSource = { state: CompatibilityState.Works, source: 'reported' };
  $: {
    if(mod) {
      const gameVersion = $selectedInstallMetadata?.version;
      const branch = $selectedInstallMetadata?.branch;
      if(gameVersion && branch) {
        if(!('offline' in mod)) {
          if(mod.hidden && !isDependency) {
            compatibility = { state: CompatibilityState.Broken, note: 'This mod was hidden by the author.', source: 'reported' };
          } else {
            getCompatibility(mod.mod_reference, branch, gameVersion, client).then((result) => {
              if (result.source === 'reported') {
                compatibility = {
                  state: result.state,
                  note: result.note 
                    ? `This mod has been reported as ${result.state} on this game version.<br>${markdown(result.note)}` 
                    : `This mod has been reported as ${result.state} on this game version. (No further notes provided)`,
                  source: 'reported',
                };
              } else {
                compatibility = result;
              }
            });
          }
        } else {
          getVersionCompatibility(mod.mod_reference, gameVersion, client).then((result) => {
            compatibility = {
              ...result,
              source: 'version',
            };
          });
        }
      }
    }
  }

  function colorForCompatibilityState(state?: CompatibilityState) {
    switch(state) {
      case CompatibilityState.Broken:
        return 'error';
      case CompatibilityState.Damaged:
        return 'warning';
      case CompatibilityState.Works:
        return 'success';
    }
    return '';
  }

  let authorsMenu: Menu;

  let versionsMenu: Menu;
  
  let changelogsMenu: Menu;

  let changelogVersion: Pick<Version, 'version' | 'changelog'>;

  $: manifestVersion = mod && $manifestMods[mod.mod_reference]?.version;
  async function installVersion(version: string | null) {
    if(!mod) {
      return;
    }
    try {
      await InstallModVersion(mod.mod_reference, version ?? '>=0.0.0');
    } catch(e) {
      if (e instanceof Error) {
        $error = e.message;
      } else if (typeof e === 'string') {
        $error = e;
      } else {
        $error = 'Unknown error';
      }
    }
  }

  function close() {
    $expandedMod = null;
  }

  let imageViewSrc: string | null = null;

  let imageViewDialog = false;

  $: if(!imageViewDialog) {
    imageViewSrc = null;
  }

  $: authorClick = () => {
    $search = `author:"${author}"`;
  };

  // Does not need offline support, since descriptions are disabled in offline mode
  function handleElementClick(element: HTMLElement) {
    if(element instanceof HTMLAnchorElement) {
      const url = new URL(element.href);
      if(url.hostname === 'ficsit.app' && url.pathname.startsWith('/mod/')) {
        const modIdOrReference = url.pathname.split('/')[2];
        if(modIdOrReference) {
          client.query(GetModReferenceDocument, {
            modIdOrReference,
          }).toPromise()
            .then((result) => {
              if (result.data?.getModByIdOrReference?.mod_reference) {
                $expandedMod = result.data.getModByIdOrReference.mod_reference;
              } else {
                console.error(`Failed to GetModReferenceDocument for modIdOrReference '${modIdOrReference}', so opening the link '${element.href}' in the browser instead.`);
                BrowserOpenURL(element.href);
              }
            });
        }
        return true;
      }
      BrowserOpenURL(element.href);
      return true;
    }
    if(element instanceof HTMLImageElement) {
      imageViewSrc = element.src;
      imageViewDialog = true;
      return true;
    }
    return false;
  }

  function handleDescriptionClick(event: MouseEvent) {
    let element: HTMLElement | null = event.target as HTMLElement;
    while(element) {
      if(handleElementClick(element)) {
        event.preventDefault();
      }
      element = element.parentElement;
    }
  }

  const compatEAPopupId = 'mod-details-compat-ea';

  const compatEAPopup = {
    event: 'hover',
    target: compatEAPopupId,
    middleware: {
      offset: 4,
    },
    placement: 'bottom-end',
  } satisfies PopupSettings;

  const compatEXPPopupId = 'mod-details-compat-exp';

  const compatEXPPopup = {
    event: 'hover',
    target: compatEXPPopupId,
    middleware: {
      offset: 4,
    },
    placement: 'bottom-start',
  } satisfies PopupSettings;

  const compatUnknownPopupId = 'mod-details-compat-unknown';

  const compatUnknownPopup = {
    event: 'hover',
    target: compatUnknownPopupId,
    middleware: {
      offset: 4,
    },
    placement: 'bottom',
  } satisfies PopupSettings;
</script>

<div class="@container/mod-details h-full flex mods-details w-full  @3xl/mod-details:text-base text-sm">
  <div class="px-4 py-4 flex flex-col h-full @3xl/mod-details:w-64 w-52 mods-details" style="border-right-color: rgba(239, 239, 239, 0.12);">
    <img src={renderedLogo} alt="{mod?.name} Logo" class="logo w-full" />
    <span class="pt-4 font-bold @3xl/mod-details:text-lg text-base">{mod?.name ?? 'Loading...'}</span>
    <span class="pt-2 font-light">A mod by:</span>
    <span bind:this={focusOnEntry} class="font-medium color-primary cursor-pointer" role="button" tabindex="0" on:click={authorClick} on:keypress={authorClick} >{author ?? 'Loading...'}</span>
    
    <div class="pt-2" on:mouseenter={() => authorsMenu.setOpen(true)} on:mouseleave={() => authorsMenu.setOpen(false)} role="listbox" tabindex="0">
      <button class="btn px-4 h-10 text-sm w-full bg-secondary-600">
        <span class="whitespace-break-spaces">Contributors <span class="color-primary">({mod?.authors.length ?? 0})</span></span>
        <SvgIcon
          class="h-5 w-5"
          icon={mdiChevronDown}/>
      </button>
      <Menu bind:this={authorsMenu} class="w-full max-h-[32rem]" anchorCorner="BOTTOM_LEFT">
        <List>
          {#each mod?.authors ?? [] as author}
            <Item style="height: 80px" on:SMUI:action={() => $search = `author:"${author.user.username}"`}>
              {#if 'avatar' in author.user}
                <img src={author.user.avatar} alt="{author.user.username} Avatar" class="avatar" />
              {/if}
              <Text class="pl-2 h-full flex flex-col content-center -mb-4">
                <PrimaryText class="text-base">{author.user.username}</PrimaryText>
                <SecondaryText class="text-base">{author.role}</SecondaryText>
              </Text>
            </Item>
          {/each}
        </List>
      </Menu>
    </div>

    <div class="pt-4">
      <span>Mod info:</span><br>
      <span>Size: </span><span class="font-bold">{size ?? 'Loading...'}</span><br>
      {#if (!mod || !('offline' in mod)) && !$offline}
        <span>Created: </span><span class="font-bold">{mod ? new Date(mod.created_at).toLocaleDateString() : 'Loading...'}</span><br>
        <span>Updated: </span><span class="font-bold">{mod ? new Date(mod.last_version_date).toLocaleString() : 'Loading...'}</span><br>
        <span>Total downloads: </span><span class="font-bold">{mod?.downloads.toLocaleString() ?? 'Loading...'}</span><br>
        <span>Views: </span><span class="font-bold">{mod?.views.toLocaleString() ?? 'Loading...'}</span><br>
        <div class="flex h-5">
          <span>Compatibility: </span>
          {#if mod?.compatibility}
            <div class="flex pl-1">
              <div use:popup={compatEAPopup}>
                <SvgIcon icon={mdiRocketLaunch} class="{colorForCompatibilityState(mod.compatibility.EA.state)} w-5" />
              </div>
              
              <Tooltip popupId={compatEAPopupId}>
                <span class="text-lg">
                  This mod has been reported as {mod.compatibility.EA.state} on Early Access.
                </span>
                {#if mod.compatibility.EA.note}
                  <!-- eslint-disable-next-line svelte/no-at-html-tags -->
                  {@html markdown(mod.compatibility.EA.note)}
                {:else}
                  (No further notes provided)
                {/if}
              </Tooltip>
              <div use:popup={compatEXPPopup}>
                <SvgIcon icon={mdiTestTube} class="{colorForCompatibilityState(mod.compatibility.EXP.state)} w-5" />
              </div>
              <Tooltip popupId={compatEXPPopupId}>
                <span class="text-lg">
                  This mod has been reported as {mod.compatibility.EXP.state} on Experimental.
                </span>
                {#if mod.compatibility.EXP.note}
                  <!-- eslint-disable-next-line svelte/no-at-html-tags -->
                  {@html markdown(mod.compatibility.EXP.note)}
                {:else}
                  (No further notes provided)
                {/if}
              </Tooltip>
            </div>
          {:else}
            <span use:popup={compatUnknownPopup} class="font-bold">&nbsp;Unknown</span>
            <Tooltip popupId={compatUnknownPopupId}>
              <span class="text-lg">No compatibility information has been reported for this mod yet. Try it out and contact us on the Discord so it can be updated!</span>
            </Tooltip>
          {/if}
        </div>
      {/if}
    </div>

    <div class="pt-4">
      <span>Latest version: </span><span class="font-bold">{ latestVersion ?? 'Loading...' }</span><br>
      <span>Installed version: </span><span class="font-bold">{ installedVersion ?? 'Loading...' }</span><br>
      <div class="pt-2">
        <button
          class="btn px-4 h-10 text-sm w-full bg-secondary-600"
          disabled={!$canModify}
          on:click={() => $canModify && versionsMenu.setOpen(!versionsMenu.isOpen())}>
          <span>Change version</span>
          <SvgIcon
            class="h-5 w-5"
            icon={mdiChevronDown}/>
        </button>
        <Menu bind:this={versionsMenu} class="min-w-[20rem] max-h-[32rem] overflow-x-visible" anchorCorner="TOP_LEFT">
          <List>
            <Item on:SMUI:action={() => installVersion(null)} disabled={!$canModify}>
              {#if manifestVersion === '>=0.0.0'}
                <SvgIcon icon={mdiCheck} class="h-5 w-5" />
              {:else}
                <div class="w-7"/>
              {/if}
              <Text class="pl-2 h-full flex flex-col content-center mb-1.5">
                <PrimaryText class="text-base">Any</PrimaryText>
              </Text>
            </Item>
            {#each mod?.versions ?? [] as version}
              <Separator insetLeading insetTrailing />
              <Item on:SMUI:action={() => installVersion(version.version)} disabled={!$canModify}>
                {#if manifestVersion && validRange(manifestVersion) && minVersion(manifestVersion)?.format() === version.version }
                  <SvgIcon icon={mdiCheck} class="h-5 w-5" />
                {:else}
                  <div class="w-7"/>
                {/if}
                <Text class="pl-2 h-full flex flex-col content-center mb-1.5">
                  <PrimaryText class="text-base">{version.version}</PrimaryText>
                </Text>
                <div class="grow"/>
                <Text class="pl-2 h-full flex flex-col content-center mb-1.5 shrink-0">
                  <PrimaryText class="text-base">or newer</PrimaryText>
                </Text>
                <div on:click|stopPropagation={() => installVersion(`>=${version.version}`)} on:keypress|stopPropagation={() => installVersion(`>=${version.version}`)} role="button" tabindex="0">
                  <Checkbox
                    input$readonly
                    checked={!!manifestVersion && !!validRange(manifestVersion) && !valid(manifestVersion) && minVersion(manifestVersion)?.format() === version.version}
                  />
                </div>
              </Item>
              <!-- {#if validRange(manifestVersion) && minVersion(manifestVersion)?.format() === version.version}
                <Separator insetLeading insetTrailing insetPadding />
                <Item on:SMUI:action={() => installVersion(`>=${version.version}`)} disabled={!$canModify}>
                  {#if validRange(manifestVersion) && !valid(manifestVersion) && minVersion(manifestVersion)?.format() === version.version}
                    <SvgIcon icon={mdiCheck} class="h-5" />
                  {:else}
                    <div class="w-7"/>
                  {/if}
                  <Text class="pl-2 h-full flex flex-col content-center mb-1.5">
                    <PrimaryText class="text-base">{version.version} or newer</PrimaryText>
                  </Text>
                </Item>
              {/if} -->
            {/each}
          </List>
        </Menu>
      </div>
      {#if (!mod || !('offline' in mod)) && !$offline}
        <div class="pt-2">
          <button
            class="btn px-4 h-10 text-sm w-full bg-secondary-600"
            on:click={() => $canModify && changelogsMenu.setOpen(!changelogsMenu.isOpen())}>
            <span>Changelogs</span>
            <SvgIcon
              class="h-5 w-5"
              icon={mdiChevronDown}/>
          </button>
          <Menu bind:this={changelogsMenu} class="min-w-[10rem] max-h-[32rem] overflow-x-visible" anchorCorner="TOP_LEFT">
            <List>
              {#each mod?.versions ?? [] as version}
                <Item on:SMUI:action={() => { changelogVersion = version; changelogsMenu.setOpen(false); }}>
                  <Text class="pl-2 h-full flex flex-col content-center mb-1.5">
                    <PrimaryText class="text-base">{version.version}</PrimaryText>
                  </Text>
                </Item>
                <Separator insetLeading insetTrailing />
              {/each}
            </List>
          </Menu>
        </div>
      {/if}
      <div class="pt-2">
        <button
          class="btn px-4 h-10 text-sm w-full bg-primary-900"
          on:click={() => BrowserOpenURL(ficsitAppLink)}>
          <span class="whitespace-break-spaces">View on ficsit.app</span>
          <SvgIcon
            class="h-5 w-5"
            icon={mdiWeb}/>
        </button>
      </div>
    </div>

    <div class="grow"/>

    <button
      class="btn px-4 h-8 w-full bg-secondary-600 text-sm"
      on:click={close}>
      <SvgIcon
        class="h-5 w-5 -scale-x-100"
        icon={mdiImport}/>
      <span>Close</span>
    </button>
  </div>
  <div class="markdown-content break-words overflow-wrap-anywhere flex-1 px-3 my-4 overflow-y-scroll overflow-x-hidden w-0">
    {#if $offline}
      <div class="flex items-center justify-center h-full text-center font-bold">Offline mode is enabled. Changelogs and descriptions are not available.</div>
    {:else if descriptionRendered}
      <!-- Intercepting mouse clicks for the link interrupter also seems to work for pressing Enter on the keyboard without a specific key handler added -->
      <!-- svelte-ignore a11y-no-noninteractive-element-interactions a11y-click-events-have-key-events -->
      <p on:click={handleDescriptionClick} role="article">
        <!-- eslint-disable-next-line svelte/no-at-html-tags -->
        {@html descriptionRendered}
      </p>
    {:else}
      <p>Loading...</p>
    {/if}
  </div>
</div>

<Dialog
  bind:open={imageViewDialog}
  surface$style="max-height: calc(100vh - 128px); max-width: calc(100vw - 128px);"
>
  <img src={imageViewSrc} alt="Dialog" style="max-height: calc(100vh - 128px); max-width: calc(100vw - 128px);"/>
</Dialog>

<Dialog open={!!changelogVersion}>
  <Title>{mod?.name} v{changelogVersion?.version}</Title>
  <Content>
    {changelogVersion?.changelog}
  </Content>
</Dialog>

<style>
  .overflow-wrap-anywhere {
    overflow-wrap: anywhere;
  }
  .mods-details {
    background-color: #2B2B2B;
  }
  .avatar {
    border-radius: 50%;
    width: 50px;
    height: 50px;
  }
</style>