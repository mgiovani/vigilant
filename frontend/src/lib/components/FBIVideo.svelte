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
    <div class="flex items-center justify-center h-full">
      <div class="text-center bg-black/40 backdrop-blur-xl rounded-2xl p-8 border border-white/10">
        <div class="p-4 rounded-full bg-red-500/20 inline-block mb-4">
          <AlertTriangle size={48} class="text-red-400" />
        </div>
        <p class="text-xl font-bold text-white">FBI ALERT</p>
        <p class="text-sm text-white/40 mt-4">{error}</p>
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

    <!-- Mute Button - Positioned top-right, glass style -->
    <button
      on:click={toggleMute}
      class="absolute top-4 right-4 z-10 p-3 rounded-xl bg-black/40 backdrop-blur-xl border border-white/10 text-white/70 hover:text-white hover:bg-black/60 transition-all"
      title={effectiveMuted ? 'Unmute' : 'Mute'}
    >
      {#if effectiveMuted}
        <VolumeX size={20} />
      {:else}
        <Volume2 size={20} />
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
