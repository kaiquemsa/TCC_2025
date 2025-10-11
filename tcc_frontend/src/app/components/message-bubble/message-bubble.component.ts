import { Component, ElementRef, Input, OnChanges, SimpleChanges, ViewChild } from '@angular/core';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { ToastService } from '../../shared/toast/toast.service';
import { ChatService } from '../../services/chat.service';
import { CommonModule } from '@angular/common';
import * as XLSX from 'xlsx';

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
  @ViewChild('bubbleContent') contentRef!: ElementRef;

  @Input() set content(value: string) {
    this.rawContent = value ?? '';
    this.updateSafeContent();
  }
  rawContent = '';
  safeContent: SafeHtml = '';
  hasTable = false;

  ngOnChanges(_: SimpleChanges) {
    this.updateSafeContent();
  }

  ngAfterViewInit() {
    setTimeout(() => {
      this.hasTable = !!this.contentRef?.nativeElement.querySelector('table');
    });
  }

  private updateSafeContent() {
    const v = this.rawContent ?? '';
    this.safeContent = this.sender === 'ai'
      ? this.sanitizer.bypassSecurityTrustHtml(v)
      : v;
  }

  async copy() {
    try {
      await navigator.clipboard.writeText(this.rawContent);
      this.toast.show('Copied');
    } catch {
      this.toast.show('Copy failed');
    }
  }

  async exportTableImage() {
    const html2canvas = (await import('html2canvas')).default;
    const table = this.contentRef.nativeElement.querySelector('table');
    if (!table) return;
    const canvas = await html2canvas(table);
    const link = document.createElement('a');
    link.href = canvas.toDataURL('image/png');
    link.download = 'tabela.png';
    link.click();
  }

  exportTableXlsx() {
    const table = this.contentRef.nativeElement.querySelector('table');
    if (!table) return;
    const wb = XLSX.utils.table_to_book(table);
    XLSX.writeFile(wb, 'tabela.xlsx');
  }
}
