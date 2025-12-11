<script>
  import { onMount } from 'svelte';
  import { Plus, Trash2, AlertCircle } from 'lucide-svelte';

  let blocklist = {
    processes: [],
    websites: [],
    patterns: []
  };

  let newEntry = {
    processes: '',
    websites: '',
    patterns: ''
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
    <div class="bg-red-500/10 border border-red-500/30 rounded-lg p-4 flex gap-3">
      <AlertCircle size={20} class="text-red-400 flex-shrink-0 mt-0.5" />
      <div>
        <p class="text-sm text-red-300">{error}</p>
      </div>
    </div>
  {/if}

  {#if loading}
    <div class="text-center py-8">
      <p class="text-gray-500">Loading blocklist...</p>
    </div>
  {:else}
    <!-- Processes Section -->
    <div>
      <h3 class="text-sm font-semibold text-white mb-4">Blocked Processes</h3>
      <div class="space-y-3">
        <!-- Input -->
        <div class="flex gap-2">
          <input
            type="text"
            bind:value={newEntry.processes}
            on:keydown={(e) => handleKeydown(e, 'processes')}
            placeholder="e.g., Discord.exe, Slack"
            class="flex-1 bg-dark-800 border border-gray-800 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-600 focus:outline-none focus:border-accent transition-colors"
          />
          <button
            on:click={() => addEntry('processes')}
            disabled={!newEntry.processes.trim()}
            class="p-2 bg-accent hover:bg-accent-hover disabled:opacity-50 disabled:cursor-not-allowed rounded-lg text-white transition-colors"
            title="Add process"
          >
            <Plus size={18} />
          </button>
        </div>

        <!-- List -->
        <div class="space-y-2">
          {#each blocklist.processes || [] as process (process)}
            <div class="flex items-center justify-between bg-dark-800 rounded-lg p-3">
              <span class="text-sm text-gray-300">{process}</span>
              <button
                on:click={() => removeEntry('processes', process)}
                class="p-1.5 hover:bg-dark-700 rounded text-gray-500 hover:text-red-400 transition-colors"
                title="Remove"
              >
                <Trash2 size={16} />
              </button>
            </div>
          {/each}
        </div>
      </div>
    </div>

    <!-- Websites Section -->
    <div>
      <h3 class="text-sm font-semibold text-white mb-4">Blocked Websites</h3>
      <div class="space-y-3">
        <!-- Input -->
        <div class="flex gap-2">
          <input
            type="text"
            bind:value={newEntry.websites}
            on:keydown={(e) => handleKeydown(e, 'websites')}
            placeholder="e.g., youtube.com, twitter.com"
            class="flex-1 bg-dark-800 border border-gray-800 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-600 focus:outline-none focus:border-accent transition-colors"
          />
          <button
            on:click={() => addEntry('websites')}
            disabled={!newEntry.websites.trim()}
            class="p-2 bg-accent hover:bg-accent-hover disabled:opacity-50 disabled:cursor-not-allowed rounded-lg text-white transition-colors"
            title="Add website"
          >
            <Plus size={18} />
          </button>
        </div>

        <!-- List -->
        <div class="space-y-2">
          {#each blocklist.websites || [] as website (website)}
            <div class="flex items-center justify-between bg-dark-800 rounded-lg p-3">
              <span class="text-sm text-gray-300">{website}</span>
              <button
                on:click={() => removeEntry('websites', website)}
                class="p-1.5 hover:bg-dark-700 rounded text-gray-500 hover:text-red-400 transition-colors"
                title="Remove"
              >
                <Trash2 size={16} />
              </button>
            </div>
          {/each}
        </div>
      </div>
    </div>

    <!-- Patterns Section -->
    <div>
      <h3 class="text-sm font-semibold text-white mb-4">Regex Patterns</h3>
      <p class="text-xs text-gray-500 mb-3">Regex patterns for advanced blocking (case-insensitive)</p>
      <div class="space-y-3">
        <!-- Input -->
        <div class="flex gap-2">
          <input
            type="text"
            bind:value={newEntry.patterns}
            on:keydown={(e) => handleKeydown(e, 'patterns')}
            placeholder="e.g., .*Netflix.*, .*Prime Video.*"
            class="flex-1 bg-dark-800 border border-gray-800 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-600 focus:outline-none focus:border-accent transition-colors"
          />
          <button
            on:click={() => addEntry('patterns')}
            disabled={!newEntry.patterns.trim()}
            class="p-2 bg-accent hover:bg-accent-hover disabled:opacity-50 disabled:cursor-not-allowed rounded-lg text-white transition-colors"
            title="Add pattern"
          >
            <Plus size={18} />
          </button>
        </div>

        <!-- List -->
        <div class="space-y-2">
          {#each blocklist.patterns || [] as pattern (pattern)}
            <div class="flex items-center justify-between bg-dark-800 rounded-lg p-3">
              <span class="text-sm text-gray-300 font-mono">{pattern}</span>
              <button
                on:click={() => removeEntry('patterns', pattern)}
                class="p-1.5 hover:bg-dark-700 rounded text-gray-500 hover:text-red-400 transition-colors"
                title="Remove"
              >
                <Trash2 size={16} />
              </button>
            </div>
          {/each}
        </div>
      </div>
    </div>
  {/if}
</div>
