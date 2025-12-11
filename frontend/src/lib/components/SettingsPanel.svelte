<script>
  import { X, Settings } from 'lucide-svelte';
  import BlocklistEditor from './BlocklistEditor.svelte';

  export let isOpen = false;

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
      <div class="flex-1 overflow-y-auto p-6">
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
