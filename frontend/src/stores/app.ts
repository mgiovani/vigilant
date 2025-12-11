import { writable, get } from 'svelte/store';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import type { Stats, PlayerState, FocusState } from '../types/index';

// Initial stats
const initialStats: Stats = {
  focusedTime: 0,
  distractedTime: 0,
  triggerCount: 0,
  sessionStart: new Date(),
  lastUpdate: new Date(),
};

export const stats = writable<Stats | null>(initialStats);
export const playerState = writable<PlayerState>('lofi');
export const focusState = writable<FocusState>({ state: 'working' });

// Flag to track initialization
let eventsInitialized = false;

/**
 * Initialize Wails event listeners for backend communication.
 * This should be called once when the app starts.
 */
export function initializeEventListeners() {
  if (eventsInitialized) {
    console.log('[stores] Events already initialized');
    return;
  }

  console.log('[stores] Initializing Wails event listeners...');

  // Listen for player state changes (lofi/fbi)
  EventsOn('player:state-change', (data: { state: string; timestamp: number }) => {
    console.log('[stores] player:state-change received:', data);
    playerState.set(data.state as PlayerState);
  });

  // Listen for focus state changes (working/distracted)
  EventsOn('focus:state-change', (data: { state: string; window: string; process: string; timestamp: number }) => {
    console.log('[stores] focus:state-change received:', data);
    focusState.set({ state: data.state as 'working' | 'distracted' });
  });

  // Listen for stats updates
  EventsOn('stats:update', (data: {
    focusedTime: string;
    distractedTime: string;
    triggerCount: number;
    focusRate: number;
    sessionStart: number;
    lastUpdate: number;
  }) => {
    console.log('[stores] stats:update received:', data);

    // Parse duration strings (e.g., "1h2m3s" or "5m30s" or "45s" or "500ms") to milliseconds
    const parseDuration = (duration: string): number => {
      // Handle Go duration format like "1h2m3.456s" or "5m30s" or "45s" or "500ms" or "1µs"
      let ms = 0;

      const hourMatch = duration.match(/(\d+)h/);
      // Use negative lookahead to avoid matching 'm' in 'ms' (milliseconds)
      const minMatch = duration.match(/(\d+)m(?![sµn])/);
      const secMatch = duration.match(/([\d.]+)s$/);
      const msMatch = duration.match(/([\d.]+)ms/);
      const usMatch = duration.match(/([\d.]+)µs/);
      const nsMatch = duration.match(/([\d.]+)ns/);

      if (hourMatch) ms += parseInt(hourMatch[1]) * 3600000;
      if (minMatch) ms += parseInt(minMatch[1]) * 60000;
      if (secMatch) ms += parseFloat(secMatch[1]) * 1000;
      if (msMatch) ms += parseFloat(msMatch[1]);
      if (usMatch) ms += parseFloat(usMatch[1]) / 1000;
      if (nsMatch) ms += parseFloat(nsMatch[1]) / 1000000;

      return ms;
    };

    stats.set({
      focusedTime: parseDuration(data.focusedTime),
      distractedTime: parseDuration(data.distractedTime),
      triggerCount: data.triggerCount,
      sessionStart: new Date(data.sessionStart * 1000),
      lastUpdate: new Date(data.lastUpdate * 1000),
    });
  });

  // Listen for window focus changes (for debugging/display)
  EventsOn('window:focus-change', (data: {
    title: string;
    process: string;
    pid: number;
    blocked: boolean;
    timestamp: number;
  }) => {
    console.log('[stores] window:focus-change received:', data);
  });

  // Listen for trigger events
  EventsOn('trigger:confirmed', (data: { window: string; process: string; timestamp: number }) => {
    console.log('[stores] trigger:confirmed received:', data);
  });

  // Listen for app errors
  EventsOn('app:error', (data: { message: string; timestamp: number }) => {
    console.error('[stores] app:error received:', data);
  });

  eventsInitialized = true;
  console.log('[stores] Wails event listeners initialized successfully');
}
