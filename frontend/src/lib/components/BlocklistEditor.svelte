<script>
  import { onMount } from 'svelte';
  import { Plus, Trash2 } from 'lucide-svelte';

  let blocklist = {
    patterns: [],
    exceptions: []
  };

  let newEntry = {
    patterns: '',
    exceptions: ''
  };

  let loading = true;
  let error = '';
  let saving = false;

  onMount(async () => {
    try {
      const result = await window.go.main.VanillaApp.GetBlocklist();
      if (result) {
        blocklist = result;
      }
    } catch (err) {
      error = 'Failed to load blocklist: ' + err.message;
    } finally {
      loading = false;
    }
  });

  async function addEntry(type) {
    const value = newEntry[type].trim();
    if (!value) return;

    try {
      await window.go.main.VanillaApp.AddBlocklistEntry(type, value);
      blocklist[type] = [...(blocklist[type] || []), value];
      newEntry[type] = '';
    } catch (err) {
      error = 'Failed to add entry: ' + err.message;
    }
  }

  async function removeEntry(type, value) {
    try {
      await window.go.main.VanillaApp.RemoveBlocklistEntry(type, value);
      blocklist[type] = blocklist[type].filter(item => item !== value);
    } catch (err) {
      error = 'Failed to remove entry: ' + err.message;
    }
  }

  function handleKeydown(e, type) {
    if (e.key === 'Enter') {
      addEntry(type);
    }
  }
</script>

<div class="space-y-8">
  {#if error}
    <div class="bg-red-500/20 backdrop-blur-sm border border-red-500/30 rounded-xl p-4 mb-6">
      <p class="text-red-400 text-sm">{error}</p>
    </div>
  {/if}

  {#if loading}
    <div class="text-center py-8">
      <p class="text-white/30">Loading blocklist...</p>
    </div>
  {:else}
    <!-- Block Patterns Section -->
    <div>
      <h3 class="text-sm font-semibold text-white/90 mb-2">Block Patterns</h3>
      <p class="text-xs text-white/40 mb-4">Apps and websites to block (matches window title or process name)</p>
      <div class="space-y-3">
        <!-- Input -->
        <div class="flex gap-2">
          <input
            type="text"
            bind:value={newEntry.patterns}
            on:keydown={(e) => handleKeydown(e, 'patterns')}
            placeholder="e.g., discord, youtube, netflix, slack"
            class="flex-1 bg-white/5 border border-white/10 rounded-lg px-4 py-2.5 text-sm text-white placeholder-white/30 focus:outline-none focus:border-white/30 focus:bg-white/10 transition-all"
          />
          <button
            on:click={() => addEntry('patterns')}
            disabled={!newEntry.patterns.trim()}
            class="p-2.5 bg-white/10 hover:bg-white/20 rounded-lg transition-all disabled:opacity-30 disabled:cursor-not-allowed"
            title="Add pattern"
          >
            <Plus size={18} class="text-white" />
          </button>
        </div>

        <!-- List -->
        <div class="space-y-2 max-h-64 overflow-y-auto">
          {#each blocklist.patterns || [] as pattern (pattern)}
            <div class="bg-white/5 rounded-lg p-3 flex items-center justify-between group hover:bg-white/10 transition-all">
              <span class="text-sm text-white/80 font-mono">{pattern}</span>
              <button
                on:click={() => removeEntry('patterns', pattern)}
                class="p-1.5 rounded hover:bg-white/10 text-white/40 hover:text-red-400 transition-colors"
                title="Remove"
              >
                <Trash2 size={16} />
              </button>
            </div>
          {/each}
          {#if !blocklist.patterns?.length}
            <p class="text-white/30 text-sm text-center py-4">No block patterns configured</p>
          {/if}
        </div>
      </div>
    </div>

    <!-- Exceptions Section -->
    <div>
      <h3 class="text-sm font-semibold text-white/90 mb-2">Exceptions</h3>
      <p class="text-xs text-white/40 mb-4">Patterns that bypass blocking (e.g., "youtube music" allows music while blocking YouTube)</p>
      <div class="space-y-3">
        <!-- Input -->
        <div class="flex gap-2">
          <input
            type="text"
            bind:value={newEntry.exceptions}
            on:keydown={(e) => handleKeydown(e, 'exceptions')}
            placeholder="e.g., youtube music, youtube studio"
            class="flex-1 bg-white/5 border border-white/10 rounded-lg px-4 py-2.5 text-sm text-white placeholder-white/30 focus:outline-none focus:border-white/30 focus:bg-white/10 transition-all"
          />
          <button
            on:click={() => addEntry('exceptions')}
            disabled={!newEntry.exceptions.trim()}
            class="p-2.5 bg-white/10 hover:bg-white/20 rounded-lg transition-all disabled:opacity-30 disabled:cursor-not-allowed"
            title="Add exception"
          >
            <Plus size={18} class="text-white" />
          </button>
        </div>

        <!-- List -->
        <div class="space-y-2 max-h-48 overflow-y-auto">
          {#each blocklist.exceptions || [] as exception (exception)}
            <div class="bg-white/5 rounded-lg p-3 flex items-center justify-between group hover:bg-white/10 transition-all">
              <span class="text-sm text-white/80 font-mono">{exception}</span>
              <button
                on:click={() => removeEntry('exceptions', exception)}
                class="p-1.5 rounded hover:bg-white/10 text-white/40 hover:text-red-400 transition-colors"
                title="Remove"
              >
                <Trash2 size={16} />
              </button>
            </div>
          {/each}
          {#if !blocklist.exceptions?.length}
            <p class="text-white/30 text-sm text-center py-4">No exceptions configured</p>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>
