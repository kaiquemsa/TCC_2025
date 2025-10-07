import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { RouterLink, RouterOutlet } from '@angular/router';
import { BsDropdownConfig, BsDropdownModule } from 'ngx-bootstrap/dropdown';
import { MessagerService } from './services/messager.service';
import { ToasterComponent } from './shared/toast/toaster.component';

interface Chat {
  id_chat: string;
  last_message_at: string;
}

@Component({
  selector: 'app-root',
  imports: [CommonModule, RouterOutlet, BsDropdownModule, ToasterComponent],
  providers: [{ provide: BsDropdownConfig, useValue: { isAnimated: true, autoClose: true } }],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent {
/*
  constructor(private messager: MessagerService) { }

  chats: Chat[] = [];
  isSidebarOpen = true;
  profileOpen = false;

  ngOnInit() {
    this.messager.getChats()
      .subscribe({
        next: (res) => this.chats = res,
        error: (err) => console.error('Erro ao carregar chats', err)
      });
  }

  toggleSidebar() {
    this.isSidebarOpen = !this.isSidebarOpen;
  }

  allowChat() {
    console.log("sendo clicado");
    sessionStorage.setItem('allowChat', 'true');
  }

  reloadChatHistories() {
    this.messager.getChats()
    .subscribe({
      next: (res) => this.chats = res,
      error: (err) => console.error('Erro ao carregar chats', err)
    });
  }

*/
}
