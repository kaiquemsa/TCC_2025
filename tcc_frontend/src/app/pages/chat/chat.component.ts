import { CommonModule } from '@angular/common';
import { ChangeDetectorRef, Component, ElementRef, NgZone, OnInit, ViewChild } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { MessagerService } from '../../services/messager.service';
import { ActivatedRoute, Router } from '@angular/router';
import { DomSanitizer } from '@angular/platform-browser';
import { v4 as uuidv4 } from 'uuid';
import { take } from 'rxjs';

@Component({
  selector: 'app-chat',
  templateUrl: './chat.component.html',
  imports: [CommonModule, FormsModule],
  styleUrls: ['./chat.component.scss']
})
export class ChatComponent implements OnInit {
  @ViewChild('messages') messagesEl!: ElementRef<HTMLDivElement>;

  constructor(private messager: MessagerService, private router: Router, private sanitizer: DomSanitizer, private route: ActivatedRoute, private ngZone: NgZone, private cdr: ChangeDetectorRef) {
    const nav = this.router.getCurrentNavigation();
    if (nav?.extras?.state?.['initialMessage']) {
      this.initialMessage = nav.extras.state['initialMessage'];
    }
  }

  private scrollToBottom() {
    this.ngZone.onStable.pipe(take(1)).subscribe(() => {
      const el = this.messagesEl?.nativeElement;
      if (el) el.scrollTop = el.scrollHeight;
    });
  }

  messagesList: any[] = [];
  messageText: string = '';
  isTyping: boolean = false;
  id_chat: string | null = '';
  initialMessage: string | undefined;
  isActive: boolean = false;
  private isSaving = false;

  ngOnInit(): void {
    console.log(this.initialMessage);
    this.route.paramMap.subscribe(params => {
      this.id_chat = params.get('id');
      if (!this.id_chat) {
        this.id_chat = uuidv4();
        this.router.navigate(['chat', 'c', this.id_chat]);
      } else {
        this.loadHistory();

        if (this.initialMessage && this.initialMessage.trim()) {
          this.messageText = this.initialMessage;
          this.messagesList.push({ from: 'me', text: this.initialMessage, id_chat: this.id_chat });
          this.messageText = '';
          setTimeout(() => {
            this.sendMessage(this.initialMessage);
          }, 300);
        }
      }
    });
  }

  loadHistory() {
    this.messager.getHistory(this.id_chat).subscribe(history => {
      this.messagesList = history || [];
    });
  }

  trackByMsg(index: number, item: any) {
    return item.id ?? index;
  }

  private saveSnapshot() {
    if (!this.messagesList?.length) return;
    if (this.isSaving) return;

    this.isSaving = true;
    this.messager.saveChat(this.messagesList).subscribe({
      next: () => { this.isSaving = false; },
      error: (e) => { console.error('erro ao salvar chat', e); this.isSaving = false; }
    });
  }

  typeWriterEffect(fullText: string, messageIndex: number, delay: number): Promise<void> {
    return new Promise((resolve) => {
      let currentText = '';
      let i = 0;
      const interval = setInterval(() => {
        currentText += fullText.charAt(i++);
        this.messagesList[messageIndex].text = currentText;
        this.scrollToBottom();

        if (i >= fullText.length) {
          clearInterval(interval);
          this.messagesList[messageIndex].text = fullText;
          this.isTyping = false;
          this.scrollToBottom();
          resolve();
        }
      }, delay);
    });
  }


  async sendMessage(text?: string) {
    if (text) this.messageText = text;
    if (!this.messageText?.trim()) return;

    const payload = {
      question: this.messageText,
      history: this.isActive ? 'Y' : 'N',
      uuid: this.id_chat
    };

    // 1) Empurra a mensagem do usuário
    this.messagesList.push({
      from: 'me',
      text: this.messageText,
      id_chat: this.id_chat,
      created_at: new Date().toISOString()
    });
    this.messageText = '';
    this.isTyping = true;
    this.scrollToBottom();

    this.messager.sendQuery(payload).subscribe({
      next: async (res) => {
        let resposta: string = res?.data ?? '';

        if (!resposta) {
          // Fallback: sem conteúdo
          this.messagesList.push({
            from: 'assistant',
            text: 'Não há conteúdo com esses dados.',
            id_chat: this.id_chat,
            created_at: new Date().toISOString()
          });
          this.isTyping = false;
          this.scrollToBottom();

          // 🔸 SALVA A INTERAÇÃO (user + fallback)
          this.saveSnapshot();
          return;
        }

        // Limpeza básica
        resposta = resposta.replace(/```html/g, '').replace(/```/g, '');

        // 2) Placeholder do assistente
        this.messagesList.push({
          from: 'assistant',
          text: '',
          id_chat: this.id_chat,
          created_at: new Date().toISOString()
        });
        const idx = this.messagesList.length - 1;
        this.scrollToBottom();

        // 3) Digitar a resposta e só então salvar o snapshot
        await this.typeWriterEffect(resposta, idx, 25);

        // 🔸 SALVA A INTERAÇÃO (user + resposta final)
        this.saveSnapshot();
      },
      error: () => {
        this.messagesList.push({
          from: 'assistant',
          text: 'Não há conteúdo com esses dados.',
          id_chat: this.id_chat,
          created_at: new Date().toISOString()
        });
        this.isTyping = false;
        this.scrollToBottom();

        // 🔸 SALVA A INTERAÇÃO (user + erro/fallback)
        this.saveSnapshot();
      }
    });
  }

  clearMessages() {
    this.messagesList = [];
  }
}
