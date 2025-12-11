<script>
  import { onMount } from 'svelte';
  import { AlertTriangle, Volume2, VolumeX } from 'lucide-svelte';

  let videoElement;
  let error = '';
  let isMuted = false;

  onMount(() => {
    if (videoElement) {
      videoElement.volume = 1.0;
    }
  });

  function handleError(e) {
    error = 'Failed to load FBI video';
    console.error('Video error:', e);
  }

  function toggleMute() {
    if (videoElement) {
      videoElement.muted = !videoElement.muted;
      isMuted = videoElement.muted;
    }
  }
</script>

<div class="absolute inset-0 bg-black">
  {#if error}
    <div class="flex items-center justify-center h-full bg-dark-950">
      <div class="text-center">
        <div class="p-4 rounded-full bg-rose-500/10 inline-block mb-4">
          <AlertTriangle size={48} class="text-rose-400" />
        </div>
        <p class="text-xl font-bold text-rose-400">FBI ALERT</p>
        <p class="text-sm text-gray-500 mt-4">{error}</p>
      </div>
    </div>
  {:else}
    <video
      bind:this={videoElement}
      class="w-full h-full object-cover animate-fbi-enter"
      autoplay
      loop
      on:error={handleError}
      playsinline
    >
      <source src="/assets/fbi-meme.mp4" type="video/mp4" />
      <track kind="captions" />
      Your browser does not support the video tag.
    </video>

    <!-- Mute Button - Positioned top-right, minimal style -->
    <button
      on:click={toggleMute}
      class="absolute top-4 right-4 p-2.5 bg-accent hover:bg-accent-hover rounded-xl text-white font-semibold transition-all duration-200 z-10 flex items-center gap-2"
    >
      {#if isMuted}
        <VolumeX size={18} />
        <span class="text-sm">Unmute</span>
      {:else}
        <Volume2 size={18} />
        <span class="text-sm">Mute</span>
      {/if}
    </button>
  {/if}
</div>

<style>
  @keyframes fbi-enter {
    from {
      opacity: 0;
      transform: scale(0.95);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  :global(.animate-fbi-enter) {
    animation: fbi-enter 0.25s ease-out;
  }
</style>
