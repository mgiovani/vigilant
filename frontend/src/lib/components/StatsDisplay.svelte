<script>
  import { onMount } from 'svelte';
  import { stats } from '../../stores/app';
  import { Focus, AlertCircle, Crosshair, TrendingUp, Clock, Settings } from 'lucide-svelte';
  let { openSettings = () => {} } = $props();

  let focusedTime = $state(0);
  let distractedTime = $state(0);
  let triggerCount = $state(0);
  let focusRate = $state(0);
  let sessionStartTime = $state(new Date().toLocaleTimeString());

  onMount(() => {
    // Subscribe to stats updates
    const statsUnsubscribe = stats.subscribe((s) => {
      if (s) {
        focusedTime = s.focusedTime;
        distractedTime = s.distractedTime;
        triggerCount = s.triggerCount;

        const totalTime = focusedTime + distractedTime;
        focusRate = totalTime > 0 ? (focusedTime / totalTime) * 100 : 0;
      }
    });

    // Poll for updates from backend every second (will integrate with Wails later)
    const pollInterval = setInterval(() => {
      // This will be replaced with Wails.invoke('GetStats') in Phase 6
      // For now, we'll update with mock data to show the component works
    }, 1000);

    return () => {
      statsUnsubscribe();
      clearInterval(pollInterval);
    };
  });

  function formatDuration(ms) {
    const totalSeconds = Math.floor(ms / 1000);
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;

    if (hours > 0) {
      return `${hours}h ${minutes}m ${seconds}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds}s`;
    } else {
      return `${seconds}s`;
    }
  }
</script>

<!-- Glassmorphism Panel -->
<div class="relative bg-black/40 backdrop-blur-xl rounded-2xl border border-white/10 p-6">
  <!-- Settings Button - Top Right -->
  <button
    onclick={() => openSettings()}
    class="absolute top-4 right-4 p-2 text-white/50 hover:text-white transition-all rounded-lg hover:bg-white/10"
  >
    <Settings size={18} />
  </button>

  <!-- 4-Column Grid - Centered -->
  <div class="grid grid-cols-4 gap-8 max-w-6xl mx-auto">
      <!-- Stat 1: Focused -->
      <div class="flex flex-col">
        <div class="flex items-center gap-2 mb-2">
          <Focus size={18} class="text-white/60" />
          <span class="text-sm font-medium text-white/80">Focused</span>
        </div>
        <span class="text-5xl font-bold text-white tracking-tight">{formatDuration(focusedTime)}</span>
        <span class="text-sm text-white/50 mt-1">Focused</span>
      </div>

      <!-- Stat 2: Distracted -->
      <div class="flex flex-col">
        <div class="flex items-center gap-2 mb-2">
          <AlertCircle size={18} class="text-white/60" />
          <span class="text-sm font-medium text-white/80">Distracted</span>
        </div>
        <span class="text-5xl font-bold text-white tracking-tight">{formatDuration(distractedTime)}</span>
        <span class="text-sm text-white/50 mt-1">Distracted</span>
      </div>

      <!-- Stat 3: FBI Triggers -->
      <div class="flex flex-col">
        <div class="flex items-center gap-2 mb-2">
          <Crosshair size={18} class="text-white/60" />
          <span class="text-sm font-medium text-white/80">FBI Triggers</span>
        </div>
        <span class="text-5xl font-bold text-white tracking-tight">{triggerCount}x</span>
        <span class="text-sm text-white/50 mt-1">FBI Triggers</span>
      </div>

    <!-- Stat 4: Focus Rate -->
    <div class="flex flex-col">
      <div class="flex items-center gap-2 mb-2">
        <TrendingUp size={18} class="text-white/60" />
        <span class="text-sm font-medium text-white/80">Focus Rate</span>
      </div>
      <span class="text-5xl font-bold text-white tracking-tight">{focusRate.toFixed(1)}%</span>
      <span class="text-sm text-white/50 mt-1">Focus Rate</span>
    </div>
  </div>

  <!-- Footer -->
  <div class="mt-6 pt-4 border-t border-white/10">
    <!-- Session Info -->
    <div class="flex items-center justify-center gap-2">
      <Clock size={14} class="text-white/50" />
      <span class="text-sm text-white/50">Session started at </span>
      <span class="font-mono text-sm text-white/70">{sessionStartTime}</span>
    </div>
  </div>
</div>
