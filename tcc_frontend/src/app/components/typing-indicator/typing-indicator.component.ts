
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-typing-indicator',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="flex gap-3 items-start">
      <div class="w-8 h-8 rounded-full grid place-items-center bg-gray-200 text-sm select-none">AI</div>
      <div class="rounded-2xl px-4 py-3 border border-default bg-card">
        <div class="flex gap-1 items-center">
          <span class="dot"></span><span class="dot"></span><span class="dot"></span>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .dot{ width:8px; height:8px; border-radius:9999px; background:#9ca3af; display:inline-block; animation:b 1.2s infinite ease-in-out; }
    .dot:nth-child(2){ animation-delay:.15s }
    .dot:nth-child(3){ animation-delay:.3s }
    @keyframes b{ 0%,80%,100%{ transform:scale(0)} 40%{ transform:scale(1)} }
  `]
})
export class TypingIndicatorComponent {}
