<script>
  import { onMount } from 'svelte';
  import { X, Settings } from 'lucide-svelte';
  import BlocklistEditor from './BlocklistEditor.svelte';

  export let isOpen = false;

  let activeTab = 'blocklist';

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

      <!-- Navigation Tabs -->
      <div class="border-b border-white/10 px-6 flex gap-0">
        <button
          on:click={() => (activeTab = 'blocklist')}
          class="px-4 py-3 text-sm font-medium transition-colors border-b-2 {activeTab === 'blocklist'
            ? 'text-white border-white'
            : 'text-white/50 border-transparent hover:text-white/70'}"
        >
          Blocklist
        </button>
        <button
          on:click={() => (activeTab = 'general')}
          class="px-4 py-3 text-sm font-medium transition-colors border-b-2 {activeTab === 'general'
            ? 'text-white border-white'
            : 'text-white/50 border-transparent hover:text-white/70'}"
        >
          General
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-6">
        {#if activeTab === 'blocklist'}
          <BlocklistEditor />
        {:else if activeTab === 'general'}
          <div class="text-center text-white/40 py-8">
            <p>General settings coming soon</p>
          </div>
        {/if}
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
