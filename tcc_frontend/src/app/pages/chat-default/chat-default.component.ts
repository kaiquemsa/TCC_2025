import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { v4 as uuidv4 } from 'uuid';

@Component({
  selector: 'app-chat',
  templateUrl: './chat-default.component.html',
  imports: [CommonModule, FormsModule],
  styleUrls: ['./chat-default.component.scss']
})
export class ChatDefaultComponent implements OnInit {
  messageText: string = '';

  constructor(private router: Router) {}

  ngOnInit() {

  }

  generateChat(wrapperEl: HTMLElement) {
    if (typeof window !== 'undefined' && typeof sessionStorage !== 'undefined') {
      sessionStorage.setItem('allowChat', 'true');
    }

    const rect = wrapperEl.getBoundingClientRect();
    const windowHeight = window.innerHeight;
    const targetBottom = 32;
    const distanceToBottom = windowHeight - rect.bottom - targetBottom;

    wrapperEl.style.transition = 'transform 0.6s cubic-bezier(.4,0,.2,1)';
    wrapperEl.style.transform = `translateY(${distanceToBottom}px)`;

    setTimeout(() => {
      wrapperEl.style.opacity = '0'
      const id = uuidv4();
      this.router.navigate(['/c', id], {
        state: { initialMessage: this.messageText }
      });
    }, 400);
  }
}
