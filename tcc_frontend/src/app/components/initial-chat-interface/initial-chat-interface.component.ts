
import { Component, EventEmitter, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
@Component({
  selector: 'app-initial-chat-interface',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './initial-chat-interface.component.html'
})
export class InitialChatInterfaceComponent {
  @Output() startChat = new EventEmitter<string>();
  inputValue = ''; height = 120;
  examplePrompts = ["Me mostre uma tabela com os 10 últimos usuários adicionados","Me mostre uma tabela com as 10 ordem de produção com mais números de pallets saída","Quais usuários pertencem ao grupo ADMIN?","Quantos itens cabem em um pallet de saída para as últimas 10 SKU?"];
  send(){ if(!this.inputValue.trim()) return; this.startChat.emit(this.inputValue); }
  onKey(e:KeyboardEvent){ if(e.key==='Enter' && !e.shiftKey){ e.preventDefault(); this.send(); } }
  autoResize(e:Event){ const el=e.target as HTMLTextAreaElement; el.style.height='auto'; const h=Math.min(el.scrollHeight,120); el.style.height=h+'px'; this.height=h; }
  setPrompt(p:string){ this.inputValue=p; }
}
