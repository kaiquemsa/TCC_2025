import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { MessagerService } from '../../services/messager.service';
import { ActivatedRoute, Router } from '@angular/router';
import { DomSanitizer } from '@angular/platform-browser';
import { v4 as uuidv4 } from 'uuid';

@Component({
  selector: 'app-chat',
  templateUrl: './chat.component.html',
  imports: [CommonModule, FormsModule],
  styleUrls: ['./chat.component.scss']
})
export class ChatComponent implements OnInit {
  constructor(private messager: MessagerService, private router: Router, private sanitizer: DomSanitizer, private route: ActivatedRoute) {
    const nav = this.router.getCurrentNavigation();
    if (nav?.extras?.state?.['initialMessage']) {
      this.initialMessage = nav.extras.state['initialMessage'];
    }
  }

  messagesList: any[] = [];
  messageText: string = '';
  isTyping: boolean = false;
  id_chat: string | null = '';
  initialMessage: string | undefined;
  isActive: boolean = false;

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

  typeWriterEffect(fullText: string, messageIndex: number, delay: number) {
    let currentText = '';
    let i = 0;
    const interval = setInterval(() => {
      currentText += fullText.charAt(i);
      this.messagesList[messageIndex].text = currentText;
      i++;
      if (i >= fullText.length) {
        this.messagesList[messageIndex].text = this.sanitizer.bypassSecurityTrustHtml(fullText);
        clearInterval(interval);
        this.isTyping = false;
      }
    }, delay);
  }

  sendMessage(text?: string) {
    if (text) {this.messageText = text};
    if (!this.messageText.trim()) return;

    let message = {
      "question": this.messageText,
      "history": "N",
      "uuid": this.id_chat
    };

    if (this.isActive) message["history"] = "Y";

    // Adiciona sua mensagem localmente
    this.messagesList.push({ from: 'me', text: this.messageText, id_chat: this.id_chat });
    this.messageText = '';
    this.isTyping = true;

    console.log(this.messagesList);
    this.messager.sendQuery(message).subscribe(
      (res) => {
        let resposta = res["data"] ?? "";
        if (!resposta) {
          setTimeout(() => {
            this.messagesList.push({
              from: 'assistant',
              text: "Não há conteúdo com esses dados.",
              id_chat: this.id_chat
            });
            this.isTyping = false;
          }, 3000);
        } else {
          resposta = resposta
            .replace(/```html/g, "")
            .replace(/```/g, "")
            .replace(/\n/g, "");
          // Adiciona a resposta do bot localmente
          this.messagesList.push({ from: 'assistant', text: '', id_chat: this.id_chat });
          const messageIndex = this.messagesList.length - 1;
          this.typeWriterEffect(resposta, messageIndex, 25);
        }
        this.isTyping = false;
      },
      (err) => {
        setTimeout(() => {
          this.messagesList.push({
            from: 'assistant',
            text: "Não há conteúdo com esses dados.",
            id_chat: this.id_chat
          });
          this.isTyping = false;
        }, 3000);
        console.log(err);
      }
    )
  }

  openHistory() {

  }

  clearMessages() {
    this.messagesList = [];
  }
}
