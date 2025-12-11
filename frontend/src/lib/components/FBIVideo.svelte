<script>
  import { onMount } from 'svelte';
  import { AlertTriangle, Volume2, VolumeX } from 'lucide-svelte';

  export let muted = false;

  let videoElement;
  let error = '';
  let userMuted = false;

  // Effective mute state: muted by parent OR muted by user
  $: effectiveMuted = muted || userMuted;

  // Update video element when mute state changes
  $: if (videoElement) {
    videoElement.muted = effectiveMuted;
    // Pause when hidden (muted by parent), play when visible
    if (muted) {
      videoElement.pause();
    } else {
      videoElement.play().catch(() => {});
    }
  }

  onMount(() => {
    if (videoElement) {
      videoElement.volume = 1.0;
      videoElement.muted = effectiveMuted;
    }
  });

  function handleError(e) {
    error = 'Failed to load FBI video';
    console.error('Video error:', e);
  }

  function toggleMute() {
    userMuted = !userMuted;
  }
</script>

<div class="player-container bg-black">
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
      class="player-video animate-fbi-enter"
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
      {#if userMuted}
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
  .player-container {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
  }

  .player-video {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

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
