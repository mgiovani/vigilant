<script>
  import { onMount, onDestroy } from 'svelte';
  import { AlertTriangle } from 'lucide-svelte';
  import { focusState } from '../../stores/app';
  import lofiBackground from '../../assets/lofi-background.webp';

  let embedUrl = '';
  let error = '';
  let loading = true;
  let iframeElement;
  let unsubscribe;

  // Send postMessage command to YouTube iframe
  function sendYouTubeCommand(command) {
    if (iframeElement && iframeElement.contentWindow) {
      iframeElement.contentWindow.postMessage(
        JSON.stringify({ event: 'command', func: command, args: [] }),
        '*'
      );
    }
  }

  function pauseVideo() {
    console.log('[LofiPlayer] Pausing YouTube video');
    sendYouTubeCommand('pauseVideo');
  }

  function playVideo() {
    console.log('[LofiPlayer] Playing YouTube video');
    sendYouTubeCommand('playVideo');
  }

  onMount(async () => {
    try {
      // Get embed URL from backend (uses HTTP localhost proxy to fix Error 153)
      const url = await window.go.main.VanillaApp.GetLofiEmbedURL();
      if (url) {
        // Ensure enablejsapi=1 is in the URL for postMessage control
        const urlObj = new URL(url);
        urlObj.searchParams.set('enablejsapi', '1');
        embedUrl = urlObj.toString();
      }
      console.log('Loaded embed URL from backend:', embedUrl);
      loading = false;
    } catch (e) {
      console.warn('Failed to get embed URL from backend:', e);
      // Fallback to direct YouTube URL with enablejsapi
      embedUrl = 'https://www.youtube-nocookie.com/embed/jfKfPfyJRdk?autoplay=1&controls=1&modestbranding=1&rel=0&playsinline=1&enablejsapi=1';
      loading = false;
    }

    // Subscribe to focus state changes
    unsubscribe = focusState.subscribe((state) => {
      if (state.state === 'distracted') {
        pauseVideo();
      } else if (state.state === 'working') {
        playVideo();
      }
    });
  });

  onDestroy(() => {
    if (unsubscribe) {
      unsubscribe();
    }
  });

</script>

<div class="player-container bg-black">
  {#if error}
    <div class="flex items-center justify-center h-full">
      <div class="text-center">
        <div class="p-4 rounded-full bg-rose-500/10 inline-block mb-4">
          <AlertTriangle size={32} class="text-rose-400" />
        </div>
        <p class="text-xl font-bold text-rose-400 mb-2">Player Error</p>
        <p class="text-sm text-gray-400">{error}</p>
        <p class="text-xs text-gray-500 mt-4">YouTube API may be unavailable</p>
      </div>
    </div>
  {:else if loading}
    <div class="loading-container" style="background-image: url({lofiBackground})">
      <div class="loading-overlay">
        <div class="inline-block relative w-12 h-12 mb-4">
          <div class="animate-spin rounded-full border-2 border-gray-700 border-t-accent h-full w-full" />
        </div>
        <p class="text-sm text-gray-300">Finding lofi stream...</p>
      </div>
    </div>
  {:else}
    <iframe
      bind:this={iframeElement}
      src={embedUrl}
      title="Lofi Hip Hop Radio"
      class="player-iframe"
      allowfullscreen
      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
    />
  {/if}
</div>

<style>
  .player-container {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    flex-direction: column;
  }

  .loading-container {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    background-size: cover;
    background-position: center;
    background-repeat: no-repeat;
  }

  .loading-overlay {
    text-align: center;
    padding: 2rem;
    background: rgba(0, 0, 0, 0.6);
    border-radius: 1rem;
    backdrop-filter: blur(4px);
  }

  .player-iframe {
    width: 100%;
    height: 100%;
    flex: 1;
    border: none;
  }

  :global(#yt-player) {
    width: 100% !important;
    height: 100% !important;
  }

  :global(iframe[srcdoc]) {
    width: 100% !important;
    height: 100% !important;
  }
</style>
