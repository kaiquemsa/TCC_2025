import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { ToastService } from '../../shared/toast/toast.service';
import { ChatService } from '../../services/chat.service';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-message-bubble',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './message-buble.component.html'
})
export class MessageBubbleComponent implements OnChanges {
  constructor(
    private toast: ToastService,
    private chat: ChatService,
    private sanitizer: DomSanitizer
  ) {}

  @Input() sender: 'user' | 'ai' = 'ai';
  @Input() timestamp: Date = new Date();
  @Input() id = '';
  @Input() liked: boolean | null | undefined = null;

  @Input() set content(value: string) {
    this.rawContent = value ?? '';
    this.updateSafeContent();
  }
  rawContent = '';
  safeContent: SafeHtml = '';

  ngOnChanges(_: SimpleChanges) {
    this.updateSafeContent();
  }

  private updateSafeContent() {
    const v = this.rawContent ?? '';

    // ✅ Opção A (recomendada): manter sanitização do Angular para TODO mundo
    // Basta NÃO usar bypass e, no template, usar [innerHTML]="rawContent"
    // this.safeContent = v; // (se você quiser essa opção, troque no template)

    // ✅ Opção B (quando o HTML é CONFIÁVEL — ex.: você gera o HTML do assistente)
    // Use bypass apenas nesses casos. Se preferir, aplique só para 'ai'.
    if (this.sender === 'ai') {
      this.safeContent = this.sanitizer.bypassSecurityTrustHtml(v);
    } else {
      this.safeContent = v;
    }
  }

  async copy() {
    try {
      await navigator.clipboard.writeText(this.rawContent);
      this.toast.show('Copied');
    } catch {
      this.toast.show('Copy failed');
    }
  }
}
