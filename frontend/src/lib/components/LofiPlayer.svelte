<script>
  import { onMount } from 'svelte';
  import { AlertTriangle } from 'lucide-svelte';

  let videoId = 'jfKfPfyJRdk';
  let error = '';
  let loading = true;

  onMount(async () => {
    try {
      // Get lofi video ID from backend
      const id = await window.go.main.VanillaApp.GetLofiVideoID();
      if (id) {
        videoId = id;
      }
      console.log('Loaded video ID from backend:', videoId);
      loading = false;
    } catch (e) {
      console.warn('Failed to get video ID from backend, using fallback:', e);
      loading = false;
    }
  });

  // HTML proxy with proper referrer headers - fixes YouTube Error 153 in WKWebView
  $: proxyHTML = `
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="UTF-8">
      <meta name="referrer" content="strict-origin-when-cross-origin">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        html, body { width: 100%; height: 100%; overflow: hidden; }
        iframe {
          width: 100%;
          height: 100%;
          border: none;
          display: block;
        }
      </style>
    </head>
    <body>
      <iframe
        src="https://www.youtube-nocookie.com/embed/${videoId}?autoplay=1&controls=1&modestbranding=1&rel=0&playsinline=1"
        referrerpolicy="strict-origin-when-cross-origin"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
        allowfullscreen>
      </iframe>
    </body>
    </html>
  `;

</script>

<div class="flex flex-col h-full w-full bg-dark-950">
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
      srcdoc={proxyHTML}
      title="Lofi Hip Hop Radio"
      class="w-full flex-1"
      allowfullscreen
      sandbox="allow-same-origin allow-scripts allow-forms allow-top-navigation allow-top-navigation-by-user-activation"
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
