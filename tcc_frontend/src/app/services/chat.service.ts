import { Injectable, signal, computed, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ChatApiService, BackendMessage } from './chat-api.service';
import { ChartSpec } from '../types/chart';
import { Envelope } from '../types/envelope';

export type Sender = 'user' | 'ai';

export type Message =
  | { id: string; sender: Sender; timestamp: Date; liked?: boolean | null; type: 'text'; text: string }
  | { id: string; sender: Sender; timestamp: Date; liked?: boolean | null; type: 'html'; html: string }
  | { id: string; sender: Sender; timestamp: Date; liked?: boolean | null; type: 'chart'; spec: ChartSpec };

export interface ChatSession {
  id_chat: string;
  last_message_at: Date;
  title?: string;
}

function uid() {
  try { return crypto.randomUUID(); } catch { return Math.random().toString(36).slice(2); }
}

const CHART_PREFIX = '__chart__:';

@Injectable({ providedIn: 'root' })
export class ChatService {
  private api = inject(ChatApiService);
  private http = inject(HttpClient);

  // estado
  private _messages = signal<Message[]>([]);
  private _chats = signal<ChatSession[]>([]);
  private _activeChatId = signal<string | null>(null);
  private _isTyping = signal<boolean>(false);
  private _historyActive = signal<boolean>(false);

  private _lastUserText = '';

  // leitura
  messages = this._messages.asReadonly();
  chats = this._chats.asReadonly();
  activeChatId = this._activeChatId.asReadonly();
  isTyping = this._isTyping.asReadonly();

  activeChat = computed(
    () => this._chats().find(c => c.id_chat === this._activeChatId()!) || null
  );

  // controle de histórico (Y/N)
  setHistoryActive(v: boolean) { this._historyActive.set(v); }

  initOnce() {
    if (!this._activeChatId()) this.newChat();
    this.refreshChats();
  }

  // ---------- Mapas Backend/UI ----------
  private toUiMessage(b: BackendMessage): Message {
    const sender: Sender = b.from === 'me' ? 'user' : 'ai';
    const timestamp = b.created_at ? new Date(b.created_at) : new Date();
    const id = b.id ?? uid();

    // Detecta gráfico salvo como string
    if (typeof b.text === 'string' && b.text.startsWith(CHART_PREFIX)) {
      try {
        const json = b.text.slice(CHART_PREFIX.length);
        const spec: ChartSpec = JSON.parse(json);
        return { id, sender, timestamp, type: 'chart', spec };
      } catch {
        // se falhar, cai como texto normal
      }
    }

    // Se vier com HTML (você pode definir uma convenção no backend para marcar),
    // aqui assumimos que histórico guarda só texto/HTML no mesmo campo "text".
    // Se quiser detectar HTML, pode fazer heurística simples:
    const isHtml = typeof b.text === 'string' && /<\/?[a-z][\s\S]*>/i.test(b.text);

    if (isHtml) {
      return { id, sender, timestamp, type: 'html', html: b.text || '' };
    } else {
      return { id, sender, timestamp, type: 'text', text: b.text || '' };
    }
  }

  private toBackendMessage(m: Message, chatId: string): BackendMessage {
    // Persistimos tudo no campo "text"
    if (m.type === 'chart') {
      return {
        id: m.id,
        id_chat: chatId,
        from: m.sender === 'user' ? 'me' : 'assistant',
        text: CHART_PREFIX + JSON.stringify(m.spec),
        created_at: m.timestamp.toISOString(),
      };
    }
    const text = m.type === 'html' ? m.html : m.text;
    return {
      id: m.id,
      id_chat: chatId,
      from: m.sender === 'user' ? 'me' : 'assistant',
      text,
      created_at: m.timestamp.toISOString(),
    };
  }

  private fromEnvelope(env: Envelope): Message {
    const id = uid();
    const timestamp = new Date();
    switch (env.type) {
      case 'chart':
        return { id, sender: 'ai', timestamp, type: 'chart', spec: env.spec };
      case 'html':
        return { id, sender: 'ai', timestamp, type: 'html', html: env.html };
      default:
        return { id, sender: 'ai', timestamp, type: 'text', text: (env as any).text ?? '' };
    }
  }

  // ---------- Sidebar ----------
  refreshChats() {
    this.api.getChats().subscribe({
      next: (rows) => {
        const mapped = (rows || []).map((r: any) => ({
          id_chat: r.id_chat,
          last_message_at: r.last_message_at ? new Date(r.last_message_at) : new Date(),
          title: r.title ?? undefined,
        }));
        this._chats.set(mapped);
      },
      error: (e) => console.error('getChats error', e),
    });
  }

  // ---------- Sessões ----------
  newChat() {
    const id = uid();
    const session: ChatSession = { id_chat: id, last_message_at: new Date() };
    this._chats.update(list => [session, ...list]);
    this._activeChatId.set(id);
    this._messages.set([]);
    this._isTyping.set(false);
  }

  setActiveChat(id: string) {
    this._activeChatId.set(id);
    this._isTyping.set(false);
    this._messages.set([]);

    this.api.getHistory(id).subscribe({
      next: (rows) => this._messages.set((rows || []).map(b => this.toUiMessage(b))),
      error: (e) => console.error('getHistory error', e),
    });
  }

  // ---------- Mensagens ----------
  sendUserMessage(question: string) {
    this._lastUserText = question;
    if (!this._activeChatId()) this.newChat();
    const cid = this._activeChatId()!;

    // adiciona user
    const user: Message = { id: uid(), sender: 'user', timestamp: new Date(), type: 'text', text: question };
    this._messages.update(a => [...a, user]);
    this.bumpSessionFromMessages();

    // chama backend
    this._isTyping.set(true);
    this.api.sendQuery({
      uuid: cid,
      question,
      history: this._historyActive() ? 'Y' : 'N',
    }).subscribe({
      next: (res: { sql: string; envelope: Envelope }) => {
        const ai = this.fromEnvelope(res?.envelope);
        this._messages.update(a => [...a, ai]);
        this.bumpSessionFromMessages();
        this._isTyping.set(false);
        this.saveSnapshot();
      },
      error: (e) => {
        const fallback: Message = {
          id: uid(),
          sender: 'ai',
          timestamp: new Date(),
          type: 'text',
          text: 'Erro ao consultar serviço. Tente novamente mais tarde.'
        };
        this._messages.update(a => [...a, fallback]);
        this.bumpSessionFromMessages();
        this._isTyping.set(false);
        this.saveSnapshot();
        console.error('sendQuery error', e);
      }
    });
  }

  retryLastAI() {
    if (this._lastUserText) this.sendUserMessage(this._lastUserText);
  }

  setReaction(id: string, liked: boolean | null) {
    this._messages.update(arr => arr.map(m => m.id === id ? { ...m, liked } : m));
  }

  // ---------- Persistência ----------
  private saveSnapshot() {
    const cid = this._activeChatId();
    if (!cid) return;
    const payload = this._messages().map(m => this.toBackendMessage(m, cid));
    this.api.saveChat(payload).subscribe({
      next: () => {},
      error: (e) => console.error('saveChat error', e),
    });
  }

  private bumpSessionFromMessages() {
    const id = this._activeChatId();
    if (!id) return;
    const ms = this._messages();
    const last = ms[ms.length - 1];
    const firstUser = ms.find(m => m.sender === 'user' && m.type === 'text') as Message | undefined;

    this._chats.update(list =>
      list.map(c => c.id_chat === id ? ({
        ...c,
        title: firstUser ? (firstUser as any).text?.slice(0, 30) || 'Título' : c.title,
        last_message_at: last ? last.timestamp : c.last_message_at
      }) : c)
    );
  }
}
