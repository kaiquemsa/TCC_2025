
import { Injectable, signal } from '@angular/core';

export interface Toast { id: string; message: string; }
@Injectable({ providedIn: 'root' })
export class ToastService {
  toasts = signal<Toast[]>([]);
  show(message: string) {
    const id = crypto.randomUUID();
    this.toasts.update(ts => [...ts, { id, message }]);
    setTimeout(() => this.dismiss(id), 2500);
  }
  dismiss(id: string) {
    this.toasts.update(ts => ts.filter(t => t.id !== id));
  }
}
