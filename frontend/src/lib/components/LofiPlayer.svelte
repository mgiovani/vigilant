<script>
  import { onMount } from 'svelte';
  import { AlertTriangle } from 'lucide-svelte';

  let embedUrl = '';
  let error = '';
  let loading = true;

  onMount(async () => {
    try {
      // Get embed URL from backend (uses HTTP localhost proxy to fix Error 153)
      const url = await window.go.main.VanillaApp.GetLofiEmbedURL();
      if (url) {
        embedUrl = url;
      }
      console.log('Loaded embed URL from backend:', embedUrl);
      loading = false;
    } catch (e) {
      console.warn('Failed to get embed URL from backend:', e);
      // Fallback to direct YouTube URL (works in browser, not in native app)
      embedUrl = 'https://www.youtube-nocookie.com/embed/jfKfPfyJRdk?autoplay=1&controls=1&modestbranding=1&rel=0&playsinline=1';
      loading = false;
    }
  });

</script>

<div class="absolute inset-0 flex flex-col bg-dark-950">
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
    <div class="flex items-center justify-center h-full">
      <div class="text-center">
        <div class="inline-block relative w-12 h-12 mb-4">
          <div class="animate-spin rounded-full border-2 border-gray-700 border-t-accent h-full w-full" />
        </div>
        <p class="text-sm text-gray-400">Finding lofi stream...</p>
      </div>
    </div>
  {:else}
    <iframe
      src={embedUrl}
      title="Lofi Hip Hop Radio"
      class="w-full h-full"
      allowfullscreen
      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
    />
  {/if}
</div>

<style>
  :global(#yt-player) {
    width: 100% !important;
    height: 100% !important;
  }

  :global(iframe[srcdoc]) {
    width: 100% !important;
    height: 100% !important;
  }
</style>
