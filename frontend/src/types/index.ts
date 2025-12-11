export interface Stats {
  focusedTime: number;      // milliseconds
  distractedTime: number;   // milliseconds
  triggerCount: number;
  sessionStart: Date;
  lastUpdate: Date;
}

export type PlayerState = 'lofi' | 'fbi';

export interface FocusState {
  state: 'working' | 'distracted';
  lastWindow?: string;
}
