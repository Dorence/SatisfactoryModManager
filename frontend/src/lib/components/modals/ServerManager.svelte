<script lang="ts">
  import { mdiAlert, mdiLoading, mdiServerNetwork, mdiTrashCan } from '@mdi/js';
  import { getTranslate } from '@tolgee/svelte';
  import _ from 'lodash';

  import RemoteServerPicker from '$lib/components/RemoteServerPicker.svelte';
  import SvgIcon from '$lib/components/SVGIcon.svelte';
  import Select from '$lib/components/Select.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';
  import { AddRemoteServer, FetchRemoteServerMetadata, RemoveRemoteServer } from '$lib/generated/wailsjs/go/ficsitcli/ficsitCLI';
  import { ficsitcli } from '$lib/generated/wailsjs/go/models';
  import { type PopupSettings, popup } from '$lib/skeletonExtensions';
  import { installsMetadata, remoteServers } from '$lib/store/ficsitCLIStore';

  const { t } = getTranslate();
  export let parent: { onClose: () => void };
  
  type RemoteType = ({ type: 'remote'; protocol: string; defaultPort: string; } | { type: 'local' }) & { name: string; };

  const remoteTypes: RemoteType[] = [
    { type: 'remote', protocol: 'ftp://', name: 'FTP', defaultPort: '21' },
    { type: 'remote', protocol: 'sftp://', name: 'SFTP', defaultPort: '22' },
    { type: 'local', name: 'Path' },
  ];

  let newRemoteType = remoteTypes[0];

  async function removeServer(server: string) {
    try {
      await RemoveRemoteServer(server);
    } catch (e) {
      if(e instanceof Error) {
        err = e.message;
      } else if (typeof e === 'string') {
        err = e;
      } else {
        err = $t('error.unknown-error');
      }
    }
  }

  let newServerUsername = '';
  let newServerPassword = '';
  let newServerHost = '';
  let newServerPort = '';
  let newServerPath = '';
  let err = '';

  let advancedMode = false;

  let addInProgress = false;

  $: authString = encodeURIComponent(newServerUsername) + (newServerPassword ? ':' + encodeURIComponent(newServerPassword) : '');
  $: actualPort = newRemoteType.type === 'remote' ? (newServerPort.length > 0 ? newServerPort : newRemoteType.defaultPort) : '';

  $: trimmedPath = _.trimStart(newServerPath, '/');

  $: fullInstallPath = (() => {
    if (newRemoteType.type === 'local') {
      return newServerPath;
    }
    if (advancedMode) {
      return newRemoteType.protocol + newServerPath;
    }
    return newRemoteType.protocol + authString + '@' + newServerHost + ':' + actualPort + '/' + trimmedPath;
  })();

  $: baseServerPath = (() => {
    if (newRemoteType.type === 'local') {
      return newServerPath;
    }
    if (advancedMode) {
      return newRemoteType.protocol + newServerPath;
    }
    return newRemoteType.protocol + authString + '@' + newServerHost + ':' + actualPort;
  })();

  $: isBaseValid = (() => {
    if (newRemoteType.type === 'local') {
      return newServerPath.length > 0;
    }
    if (advancedMode) {
      return newServerPath.length > 0;
    }
    return newServerUsername.length > 0 && newServerHost.length > 0;
  })();

  let isPathValid = false;

  $: isValid = (() => {
    if (newRemoteType.type === 'local') {
      return newServerPath.length > 0;
    }
    if (advancedMode) {
      return newServerPath.length > 0;
    }
    return newServerUsername.length > 0 && newServerHost.length > 0 && isPathValid;
  })();

  async function addNewRemoteServer() {
    if (!isValid) {
      return;
    }
    try {
      err = '';
      addInProgress = true;
      await AddRemoteServer(fullInstallPath);
      newServerUsername = '';
      newServerPassword = '';
      newServerHost = '';
      newServerPort = '';
      newServerPath = '';
    } catch (e) {
      if(e instanceof Error) {
        err = e.message;
      } else if (typeof e === 'string') {
        err = e;
      } else {
        err = $t('error.unknown-error');
      }
    } finally {
      addInProgress = false;
    }
  }

  async function retryConnect(server: string) {
    try {
      await FetchRemoteServerMetadata(server);
    } catch (e) {
      if(e instanceof Error) {
        err = e.message;
      } else if (typeof e === 'string') {
        err = e;
      } else {
        err = $t('error.unknown-error');
      }
    }
  }

  function installWarningPopupId(install: string) {
    return `remote-server-warning-${install}`;
  }

  $: installWarningPopups = $remoteServers.map((i) => [i, {
    event: 'hover',
    target: installWarningPopupId(i),
    middleware: {
      offset: 4,
    },
    placement: 'bottom',
  } as PopupSettings]).reduce((acc, [k, v]) => ({ ...acc, [k as string]: v as PopupSettings }), {} as Record<string, PopupSettings>);
</script>


<div style="max-height: calc(100vh - 3rem); max-width: calc(100vw - 3rem);" class="card flex flex-col gap-2">
  <header class="card-header font-bold text-2xl text-center">
    {$t('server-manager.dedicated-servers')}
  </header>
  <section class="p-4 flex-auto space-y-4 overflow-y-auto flex">
    <div class="flex-auto w-full overflow-x-auto overflow-y-auto">
      <table class="table">
        <tbody>
          {#each $remoteServers as remoteServer}
            <tr>
              <td class="break-all">{remoteServer}</td>
              <td>
                {#if $installsMetadata[remoteServer]?.state === ficsitcli.InstallState.VALID}
                  {$t('arch.' + $installsMetadata[remoteServer].info?.type, $t('unknown'))}
                {:else}
                  <Tooltip popupId={installWarningPopupId(remoteServer)}>
                    <span class="text-base">
                      {#if $installsMetadata[remoteServer]?.state === ficsitcli.InstallState.LOADING}
                        {$t('loading')}
                      {:else if $installsMetadata[remoteServer]?.state === ficsitcli.InstallState.INVALID}
                        {$t('server-manager.cant-manage-server')}
                      {:else}
                        {$t('server-manager.failed-to-connect-server')}
                      {/if}
                    </span>
                  </Tooltip>
                  <div
                    class="h-6 w-full text-sm"
                    use:popup={installWarningPopups[remoteServer]}>
                    {#if $installsMetadata[remoteServer]?.state === ficsitcli.InstallState.LOADING}
                      <SvgIcon
                        class="!p-0 !m-0 !w-full !h-full animate-spin text-primary-600"
                        icon={mdiLoading} />
                    {:else if $installsMetadata[remoteServer]?.state === ficsitcli.InstallState.INVALID}
                      <SvgIcon
                        class="!p-0 !m-0 !w-full !h-full text-red-500"
                        icon={mdiAlert} />
                    {:else}
                      <button
                        class="btn-icon h-6 w-full"
                        on:click={() => retryConnect(remoteServer)}>
                        <SvgIcon
                          class="!p-0 !m-0 !w-full !h-full text-red-500"
                          icon={mdiAlert} />
                      </button>
                    {/if}
                  </div>
                {/if}
              </td>
              <td>
                {#if $installsMetadata[remoteServer]?.info?.version}
                  CL{$installsMetadata[remoteServer].info?.version}
                {/if}
              </td>
              <td>
                <button
                  class="btn-icon h-5 w-full"
                  on:click={() => removeServer(remoteServer)}>
                  <SvgIcon
                    class="!p-0 !m-0 !w-full !h-full hover:text-red-500"
                    icon={mdiTrashCan}/>
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </section>
  <section class="p-4 space-y-4 overflow-y-auto">
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 items-start auto-rows-[minmax(2.5rem,_max-content)]">
      <Select
        name="newServerProtocol"
        class="col-span-1 h-full"
        buttonClass="bg-surface-200-700-token px-4 text-sm"
        itemActiveClass="!bg-surface-300/20"
        itemClass="bg-surface-50-900-token"
        itemKey="name"
        items={remoteTypes}
        bind:value={newRemoteType}>
        <svelte:fragment slot="item" let:item>
          {item.name}
        </svelte:fragment>
      </Select>
      {#if newRemoteType.type === 'remote'}
        {#if advancedMode}
          <input
            class="input px-4 h-full sm:col-start-2 col-span-2"
            placeholder="user:pass@host:port/path"
            type="text"
            bind:value={newServerPath}/>
          <p class="sm:col-start-2 col-span-2">
            {$t('server-manager.new-server.advanced-note')}
          </p>
        {:else}
          <input
            class="input px-4 h-full sm:col-start-2"
            placeholder={$t('server-manager.new-server.username')}
            type="text"
            bind:value={newServerUsername}/>
          <input
            class="input px-4 h-full"
            placeholder={$t('server-manager.new-server.password')}
            type="text"
            bind:value={newServerPassword}/>
          <input
            class="input px-4 h-full sm:col-start-2"
            placeholder={$t('server-manager.new-server.host')}
            type="text"
            bind:value={newServerHost}/>
          <input
            class="input px-4 h-full"
            placeholder={$t('server-manager.new-server.port', { port: newRemoteType.defaultPort })}
            type="text"
            bind:value={newServerPort}/>
          <input
            class="input px-4 h-full sm:col-start-2 col-span-2"
            placeholder={$t('server-manager.new-server.path')}
            type="text"
            bind:value={newServerPath}/>
          <div class="sm:col-start-2 col-span-2">
            <RemoteServerPicker
              basePath={baseServerPath}
              disabled={!isBaseValid}
              bind:path={newServerPath}
              bind:valid={isPathValid}
            />
          </div>
        {/if}
        <button class="btn sm:col-start-1 col-span-1 row-start-2 text-sm whitespace-break-spaces bg-surface-200-700-token" on:click={() => advancedMode = !advancedMode}>
          {#if advancedMode}
            {$t('server-manager.new-server.switch-to-simple')}
          {:else}
            {$t('server-manager.new-server.switch-to-advanced')}
          {/if}
        </button>
      {:else}
        <input
          class="input px-4 h-full sm:col-start-2 col-span-2"
          placeholder={$t('server-manager.new-server.local-path')}
          type="text"
          bind:value={newServerPath}/>
        <div class="sm:col-start-2 col-span-2">
          <RemoteServerPicker
            basePath=""
            bind:path={newServerPath}
            bind:valid={isPathValid}
          />
        </div>
      {/if}
      <button
        class="btn h-full text-sm bg-primary-600 text-secondary-900 col-start-2 sm:col-start-4 row-start-1"
        disabled={addInProgress || !isValid}
        on:click={() => addNewRemoteServer()}>
        <span>
          {#if !addInProgress}
            {$t('add')}
          {:else}
            {$t('validating')}
          {/if}
        </span>
        <div class="grow" />
        <SvgIcon
          class="h-5 w-5"
          icon={mdiServerNetwork} />
      </button>
    </div>
    <p>{err}</p>
  </section>
  <footer class="card-footer">
    <button
      class="btn h-8 w-full text-sm bg-surface-200-700-token"
      on:click={parent.onClose}>
      {$t('close')}
    </button>
  </footer>
</div>
