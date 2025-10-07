
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ToastService } from './toast.service';

@Component({
  selector: 'app-toaster',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="fixed bottom-4 right-4 space-y-2 z-50">
      <div *ngFor="let t of svc.toasts()" class="px-4 py-2 rounded-xl shadow border border-border bg-white dark:bg-gray-900">
        {{ t.message }}
      </div>
    </div>
  `
})
export class ToasterComponent {
  constructor(public svc: ToastService) {}
}
