
import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ChatSidebarComponent } from '../../components/chat-sidebar/chat-sidebar.component';
import { ChatInterfaceComponent } from '../../components/chat-interface/chat-interface.component';
import { ChatService } from '../../services/chat.service';
import { InitialChatInterfaceComponent } from '../../components/initial-chat-interface/initial-chat-interface.component';

@Component({
  selector: 'app-index',
  standalone: true,
  imports: [CommonModule, ChatSidebarComponent, ChatInterfaceComponent, InitialChatInterfaceComponent],
  templateUrl: './index.page.html'
})
export class IndexPage implements OnInit {
  constructor(private chat: ChatService) {}
  chatStarted = signal(false); initialMessage = signal('');
  ngOnInit(){ this.chat.initOnce(); }
  onStart(message:string){ this.initialMessage.set(message); this.chatStarted.set(true); }

  onNewChat() {
    this.initialMessage.set('');
    this.chatStarted.set(false); // volta para a tela Initial
  }
  
}
