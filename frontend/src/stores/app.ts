import { writable } from 'svelte/store';
import type { Stats, PlayerState, FocusState } from '../types/index';

// Mock initial stats for development
const mockStats: Stats = {
  focusedTime: 0,
  distractedTime: 0,
  triggerCount: 0,
  sessionStart: new Date(),
  lastUpdate: new Date(),
};

export const stats = writable<Stats | null>(mockStats);
export const playerState = writable<PlayerState>('lofi');
export const focusState = writable<FocusState>({ state: 'working' });
