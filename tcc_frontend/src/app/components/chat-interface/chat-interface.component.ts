
import { Component, ViewChild, ElementRef, effect, OnInit, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MessageBubbleComponent } from '../message-bubble/message-bubble.component';
import { TypingIndicatorComponent } from '../typing-indicator/typing-indicator.component';
import { ChatService } from '../../services/chat.service';
import { ToastService } from '../../shared/toast/toast.service';
import { ThemeService } from '../../shared/theme/theme.service';
import { ChartMessageComponent } from '../chart-message/chart-message.component';
import type { Message } from '../../services/chat.service';

@Component({
  selector: 'app-chat-interface',
  standalone: true,
  imports: [CommonModule, FormsModule, MessageBubbleComponent, TypingIndicatorComponent, ChartMessageComponent],
  templateUrl: './chat-interface.component.html'
})
export class ChatInterfaceComponent implements OnInit {
  @Input() initialMessage = '';
  isActive: boolean = false;
  isHistoryChat = false;
  draft = ''; composerHeight = 44; now = new Date();
  welcomeText = `Hello! I'm Assistant AI, your intelligent companion. I can help you with a wide range of tasks, from answering questions and providing explanations to assisting with creative projects and problem-solving. What would you like to explore today?`;
  constructor(public service: ChatService, private toast: ToastService, public theme: ThemeService) {
    effect(() => {
      this.service.messages();
      queueMicrotask(() => {
        const el = document.querySelector('.flex-1.overflow-y-auto');
        if (el) (el as HTMLElement).scrollTo({ top: (el as HTMLElement).scrollHeight });
      });
    });
  }

  isChart(m: Message): m is Extract<Message, { type: 'chart' }> {
    return m.type === 'chart';
  }
  isHtml(m: Message): m is Extract<Message, { type: 'html' }> {
    return m.type === 'html';
  }
  isText(m: Message): m is Extract<Message, { type: 'text' }> {
    return m.type === 'text';
  }

  ngOnInit(){ console.log(window.location.href); if(this.initialMessage?.trim()) this.service.sendUserMessage(this.initialMessage.trim()); }
  onSend(){ const t=this.draft.trim(); if(!t) return; this.service.sendUserMessage(t); this.draft=''; this.composerHeight=44; }
  onKey(e:KeyboardEvent){ if(e.key==='Enter' && !e.shiftKey){ console.log(e); e.preventDefault(); this.onSend(); } }
  autoResize(e:Event){ const el=e.target as HTMLTextAreaElement; el.style.height='auto'; const h=Math.min(el.scrollHeight,120); el.style.height=h+'px'; this.composerHeight=h; }
  activeHistory(){ this.isActive = !this.isActive; this.service.setHistoryActive(this.isActive); };
}
