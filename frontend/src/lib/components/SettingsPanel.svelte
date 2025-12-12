<script>
  import { onMount } from 'svelte';
  import { X, Settings, Music } from 'lucide-svelte';
  import BlocklistEditor from './BlocklistEditor.svelte';

  export let isOpen = false;

  let lofiUrl = '';
  let saving = false;

  onMount(async () => {
    try {
      lofiUrl = await window.go.main.VanillaApp.GetLofiPlaylist();
    } catch (e) {
      console.warn('Failed to get lofi playlist:', e);
    }
  });

  async function saveLofiUrl() {
    if (saving) return;
    saving = true;
    try {
      await window.go.main.VanillaApp.SetLofiPlaylist(lofiUrl);
    } catch (e) {
      console.error('Failed to save lofi playlist:', e);
    }
    saving = false;
  }

  function closePanel() {
    isOpen = false;
  }

  function handleBackdropClick(e) {
    if (e.target === e.currentTarget) {
      closePanel();
    }
  }
</script>

{#if isOpen}
  <div
    class="fixed inset-0 bg-black/60 backdrop-blur-md z-50 flex justify-end"
    on:click={handleBackdropClick}
  >
    <!-- Settings Panel -->
    <div class="w-full max-w-2xl h-full bg-black/80 backdrop-blur-xl border-l border-white/10 flex flex-col animate-in slide-in-from-right duration-300">
      <!-- Header -->
      <div class="border-b border-white/10 p-6 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <Settings size={24} class="text-white/70" />
          <h2 class="text-xl font-semibold text-white">Settings</h2>
        </div>
        <button
          on:click={closePanel}
          class="p-2 hover:bg-white/10 rounded-lg transition-colors"
          title="Close settings"
        >
          <X size={20} class="text-white/50 hover:text-white" />
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-6 space-y-8">
        <!-- Focus Music Section -->
        <div>
          <div class="flex items-center gap-2 mb-3">
            <Music size={18} class="text-white/70" />
            <h3 class="text-base font-medium text-white/90">Focus Music</h3>
          </div>
          <div class="flex gap-2">
            <input
              type="text"
              bind:value={lofiUrl}
              placeholder="YouTube URL or Video ID"
              class="flex-1 bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm text-white placeholder-white/30 focus:outline-none focus:border-white/30"
            />
            <button
              on:click={saveLofiUrl}
              disabled={saving}
              class="px-4 py-2 bg-white/10 hover:bg-white/20 disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
            >
              {saving ? '...' : 'Save'}
            </button>
          </div>
          <p class="mt-2 text-xs text-white/40">Supports YouTube URLs, video IDs, or live streams</p>
        </div>

        <BlocklistEditor />
      </div>
    </div>
  </div>
{/if}

<style>
  :global(.animate-in.slide-in-from-right) {
    animation: slide-in 0.3s ease-out;
  }

  @keyframes slide-in {
    from {
      transform: translateX(100%);
    }
    to {
      transform: translateX(0);
    }
  }
</style>
