<script>
  import { onMount } from 'svelte';
  import { playerState, focusState, initializeEventListeners } from './stores/app';
  import LofiPlayer from './lib/components/LofiPlayer.svelte';
  import FBIVideo from './lib/components/FBIVideo.svelte';
  import StatsDisplay from './lib/components/StatsDisplay.svelte';
  import SettingsPanel from './lib/components/SettingsPanel.svelte';
  
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
</script>

<main class="relative h-screen w-screen bg-black overflow-hidden">
  <!-- Full-screen video background - Both players stay mounted, visibility toggled -->
  <div class="absolute inset-0">
    <div class="absolute inset-0" class:hidden={currentPlayerState !== 'lofi'}>
      <LofiPlayer />
    </div>
    <div class="absolute inset-0" class:hidden={currentPlayerState === 'lofi'}>
      <FBIVideo muted={currentPlayerState === 'lofi'} />
    </div>
  </div>

  <!-- Stats Glass Panel - bottom overlay (raised to avoid YouTube controls) -->
  <div class="absolute bottom-20 left-6 right-6 z-30">
    <StatsDisplay openSettings={() => (settingsPanelOpen = true)} />
  </div>
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
