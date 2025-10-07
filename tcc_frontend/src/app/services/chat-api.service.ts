import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../environments/environment';
import { Envelope } from '../types/envelope';
import { QueryRequest, QueryResponse } from '../types/api';

export interface BackendMessage {
  id?: string;
  from: 'me' | 'assistant';
  text: string;
  id_chat: string;
  created_at?: string; // ISO
}

export interface BackendChat {
  id: string;
  title?: string;
  lastMessage?: string;
  updatedAt?: string; // ISO
}

@Injectable({ providedIn: 'root' })
export class ChatApiService {
  private http = inject(HttpClient);
  private base = environment.URL;

  sendQuery(body: QueryRequest): Observable<any> {
    return this.http.post<QueryResponse>(`${this.base}/api/query`, body);
  }
  saveChat(messages: BackendMessage[]) {
    return this.http.post(`${this.base}/api/save-chat`, messages);
  }
  getHistory(uuid: string) {
    return this.http.get<BackendMessage[]>(`${this.base}/api/get-history`, { params: { uuid } });
  }
  getChats() {
    return this.http.get<BackendChat[]>(`${this.base}/api/get-chats`);
  }
}
