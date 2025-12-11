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
    class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex justify-end"
    on:click={handleBackdropClick}
  >
    <!-- Settings Panel -->
    <div class="w-full max-w-2xl h-full bg-dark-900 border-l border-gray-800 flex flex-col animate-in slide-in-from-right duration-300">
      <!-- Header -->
      <div class="border-b border-gray-800 p-6 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <Settings size={24} class="text-accent" />
          <h2 class="text-xl font-semibold text-white">Settings</h2>
        </div>
        <button
          on:click={closePanel}
          class="p-2 hover:bg-dark-800 rounded-lg transition-colors"
          title="Close settings"
        >
          <X size={20} class="text-gray-400" />
        </button>
      </div>

      <!-- Navigation Tabs -->
      <div class="border-b border-gray-800 px-6 flex gap-0">
        <button
          on:click={() => (activeTab = 'blocklist')}
          class="px-4 py-3 text-sm font-medium transition-colors border-b-2 {activeTab === 'blocklist'
            ? 'text-white border-accent'
            : 'text-gray-500 border-transparent hover:text-gray-400'}"
        >
          Blocklist
        </button>
        <button
          on:click={() => (activeTab = 'exceptions')}
          class="px-4 py-3 text-sm font-medium transition-colors border-b-2 {activeTab === 'exceptions'
            ? 'text-white border-accent'
            : 'text-gray-500 border-transparent hover:text-gray-400'}"
        >
          Exceptions
        </button>
        <button
          on:click={() => (activeTab = 'general')}
          class="px-4 py-3 text-sm font-medium transition-colors border-b-2 {activeTab === 'general'
            ? 'text-white border-accent'
            : 'text-gray-500 border-transparent hover:text-gray-400'}"
        >
          General
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-6">
        {#if activeTab === 'blocklist'}
          <BlocklistEditor />
        {:else if activeTab === 'exceptions'}
          <div class="text-center text-gray-500 py-8">
            <p>Exceptions editor coming soon</p>
          </div>
        {:else if activeTab === 'general'}
          <div class="text-center text-gray-500 py-8">
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
