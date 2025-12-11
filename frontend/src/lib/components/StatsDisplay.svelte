<script>
  import { onMount } from 'svelte';
  import { stats, focusState } from '../../stores/app';
  import { BarChart3, Focus, AlertCircle, Crosshair, TrendingUp, Clock } from 'lucide-svelte';

  let focusedTime = 0;
  let distractedTime = 0;
  let triggerCount = 0;
  let focusRate = 0;

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

  function getFocusRateColor(rate) {
    return 'text-white';
  }
</script>

<div class="h-full flex flex-col gap-6">
  <!-- Header -->
  <div class="space-y-1">
    <div class="flex items-center gap-2">
      <BarChart3 size={20} class="text-accent-400" />
      <h2 class="text-lg font-bold text-gray-100">Session Stats</h2>
    </div>
    <p class="text-xs text-gray-500">Real-time focus tracking</p>
  </div>

  <!-- Stats Cards - Monochrome -->
  <div class="space-y-2">
    <!-- Focused Time Card -->
    <div class="bg-dark-900 rounded-xl p-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <Focus size={16} class="text-gray-500" />
          <span class="text-sm text-gray-400">Focused</span>
        </div>
        <span class="font-mono text-sm font-semibold text-white">{formatDuration(focusedTime)}</span>
      </div>
    </div>

    <!-- Distracted Time Card -->
    <div class="bg-dark-900 rounded-xl p-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <AlertCircle size={16} class="text-gray-500" />
          <span class="text-sm text-gray-400">Distracted</span>
        </div>
        <span class="font-mono text-sm font-semibold text-white">{formatDuration(distractedTime)}</span>
      </div>
    </div>

    <!-- FBI Triggers Card -->
    <div class="bg-dark-900 rounded-xl p-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <Crosshair size={16} class="text-gray-500" />
          <span class="text-sm text-gray-400">FBI Triggers</span>
        </div>
        <span class="font-mono text-sm font-semibold text-white">{triggerCount}x</span>
      </div>
    </div>

    <!-- Focus Rate Card -->
    <div class="bg-dark-900 rounded-xl p-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <TrendingUp size={16} class="text-gray-500" />
          <span class="text-sm text-gray-400">Focus Rate</span>
        </div>
        <span class="font-mono text-sm font-semibold {getFocusRateColor(focusRate)}">{focusRate.toFixed(1)}%</span>
      </div>
    </div>
  </div>

  <!-- Session Info Footer -->
  <div class="mt-auto pt-4 border-t border-gray-800">
    <div class="flex items-center gap-2 text-gray-500">
      <Clock size={14} />
      <div class="text-xs">
        <span>Session started at </span>
        <span class="font-mono text-gray-400">{new Date().toLocaleTimeString()}</span>
      </div>
    </div>
  </div>
</div>
