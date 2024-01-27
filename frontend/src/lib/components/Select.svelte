<script lang="ts" generics="T">
  import { mdiMenuDown } from '@mdi/js';
  import { popup, type PopupSettings, ListBox, ListBoxItem } from '@skeletonlabs/skeleton';
  import type { SizeOptions } from '@floating-ui/dom';
  import { createEventDispatcher, type EventDispatcher } from 'svelte';

  import SvgIcon from '$lib/components/SVGIcon.svelte';

  export let name: string;
  export let items: T[];
  export let value: T;
  export let disabled = false;
  let clazz = '';
  export { clazz as class };
  export let buttonClass = '';
  export let menuClass = '';
  export let itemClass = '';
  export let itemActiveClass = '';
  
  export let itemKey: ((item: T) => string) | keyof T = (item) => typeof item === 'string' ? item : JSON.stringify(item);

  function getKey(item: T) {
    return typeof itemKey === 'function' ? itemKey(item) : item[itemKey];
  }

  let comboboxOpen = false;
  $: combobox = {
    event: 'click',
    target: name,
    placement: 'bottom-start',
    middleware: {
      offset: 6,
      size: {
        apply({ availableHeight, elements }) {
          Object.assign(elements.floating.style, {
            maxHeight: `${availableHeight * 0.8}px`,
          });
        },
      } as SizeOptions,
    },
    state: ({ state }) => comboboxOpen = state,
    closeQuery: `.${name}-listbox-item`,
  } as PopupSettings;

  const dispatch: EventDispatcher<{change: T}> = createEventDispatcher();

  function onChange({ target }: Event) {
    const newValue = (target as HTMLButtonElement).value;
    value = items.find((item) => getKey(item) === newValue)!;
    dispatch('change', value);
  }
</script>

<div class="{clazz} relative">
  <div class="h-full w-full" use:popup={combobox}>
    <!--
    The button scale down on click animation changes its bounds, which the popup uses to position itself.
    Wrap button in a div so that the trigger node location does not change.    
    -->
    <button class="btn w-full h-full {buttonClass}" {disabled}>
      {#if $$slots.selected}
        <slot name="selected" item={value} />
      {:else}
        <slot name="item" item={value}>
          <span>
            {value}
          </span>
        </slot>
      {/if}
      <div class="grow" />
      <SvgIcon
        class="h-5 w-5 p-0.5 {comboboxOpen ? 'text-primary-600 -scale-y-100' : ''} transition-all shrink-0"
        icon={mdiMenuDown} />
    </button>
  </div>

  <div class="card w-full shadow-xl z-10 duration-0 overflow-y-auto {menuClass}" data-popup={name}>
    <!-- 
    Skeleton's popup close function waits for the tranistion duration...
    before actually triggering the transition...
    So we'll just not have a transition...
    -->
    <ListBox class="w-full" rounded="rounded-none" spacing="space-y-0">
      {#each items as item}
        <ListBoxItem group={getKey(value)} on:change={onChange} {name} value={getKey(item)} class="{name}-listbox-item {itemClass}" active="{itemActiveClass}">
          <slot name="item" {item}>
            {item}
          </slot>
          <slot slot="trail" name="itemTrail" {item} />
        </ListBoxItem>
      {/each}
    </ListBox>
  </div>
</div>