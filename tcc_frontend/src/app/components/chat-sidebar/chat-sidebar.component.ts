
import { Component, EventEmitter, Output, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ChatService } from '../../services/chat.service';

@Component({
  selector: 'app-chat-sidebar',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './chat-sidebar.component.html'
})
export class ChatSidebarComponent {
  @Output() newChat = new EventEmitter<void>();
  constructor(public service: ChatService) {console.log(service.chats())}
  collapsed = signal(false);

  handleNewChat() {
    this.service.newChat();   // cria a nova conversa no serviço
    this.newChat.emit();      // sinaliza para a IndexPage voltar à tela Initial
  }

}
