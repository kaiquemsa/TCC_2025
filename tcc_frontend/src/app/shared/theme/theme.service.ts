
import { Injectable, signal } from '@angular/core';

export type Theme = 'light' | 'dark';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  theme = signal<Theme>('light');

  constructor() {
    const saved = typeof localStorage !== 'undefined' ? localStorage.getItem('theme') as Theme | null : null;
    const prefersDark = typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
    const initial: Theme = saved ?? (prefersDark ? 'dark' : 'light');
    this.apply(initial);
  }

  toggle() {
    this.apply(this.theme() === 'light' ? 'dark' : 'light');
  }

  apply(t: Theme) {
    this.theme.set(t);
    if (typeof document !== 'undefined') {
      const root = document.documentElement;
      root.classList.toggle('dark', t === 'dark');
    }
    try { localStorage.setItem('theme', t); } catch {}
  }
}
