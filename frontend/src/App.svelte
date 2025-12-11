<script>
  import { onMount } from 'svelte';
  import { playerState, focusState, initializeEventListeners } from './stores/app';
  import LofiPlayer from './lib/components/LofiPlayer.svelte';
  import FBIVideo from './lib/components/FBIVideo.svelte';
  import StatsDisplay from './lib/components/StatsDisplay.svelte';
  import SettingsPanel from './lib/components/SettingsPanel.svelte';
  import { CheckCircle, AlertTriangle, Settings } from 'lucide-svelte';

  let currentPlayerState = 'lofi';
  let currentFocusState = 'working';
  let settingsPanelOpen = false;

  onMount(() => {
    // Initialize Wails event listeners for backend communication
    initializeEventListeners();

    // Subscribe to player state changes
    const playerUnsubscribe = playerState.subscribe((state) => {
      currentPlayerState = state;
    });

    // Subscribe to focus state changes
    const focusUnsubscribe = focusState.subscribe((state) => {
      currentFocusState = state.state;
    });

    return () => {
      playerUnsubscribe();
      focusUnsubscribe();
    };
  });

  $: statusConfig = currentFocusState === 'working'
    ? { text: 'Working', dotColor: 'bg-accent' }
    : { text: 'Distracted', dotColor: 'bg-gray-500' };
</script>

<main class="flex h-screen bg-dark-950 text-gray-100 overflow-hidden">
  <!-- Player Panel - Both players stay mounted, visibility toggled -->
  <div class="flex-1 flex flex-col relative">
    <div class="absolute inset-0" class:hidden={currentPlayerState !== 'lofi'}>
      <LofiPlayer />
    </div>
    <div class="absolute inset-0" class:hidden={currentPlayerState === 'lofi'}>
      <FBIVideo muted={currentPlayerState === 'lofi'} />
    </div>

    <!-- Status Indicator Overlay -->
    <div class="absolute top-4 left-4 z-40">
      <div class="px-3 py-1.5 rounded-lg bg-dark-900/90 backdrop-blur-sm border border-gray-800 flex items-center gap-2 transition-all duration-300">
        <div class="w-2 h-2 rounded-full {statusConfig.dotColor}"></div>
        <span class="text-xs font-medium text-white">{statusConfig.text}</span>
      </div>
    </div>

    <!-- Settings Button -->
    <button
      on:click={() => (settingsPanelOpen = true)}
      class="absolute top-4 right-4 z-40 p-2 rounded-lg bg-dark-900/90 backdrop-blur-sm border border-gray-800 text-gray-400 hover:text-white hover:border-accent transition-all duration-200"
      title="Open settings"
    >
      <Settings size={20} />
    </button>
  </div>

  <!-- Stats Sidebar -->
  <aside class="w-80 bg-dark-900 border-l border-gray-800 p-6 overflow-y-auto">
    <StatsDisplay />
  </aside>
</main>

<!-- Settings Panel -->
<SettingsPanel bind:isOpen={settingsPanelOpen} />

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }

  :global(#app) {
    margin: 0;
    padding: 0;
  }
</style>
