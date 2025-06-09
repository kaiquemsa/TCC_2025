import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, tap } from 'rxjs';
import { environment } from '../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class MessagerService {

  constructor(private http: HttpClient) { }

  public sendQuery(query: any): Observable<any> {
    const url = environment.URL + '/api/query';
    return this.http.post(url, query);
  }

  public saveChat(chat: any): Observable<any> {
    const url = environment.URL + '/api/save-chat';
    return this.http.post(url, chat);
  }

  public getHistory(chat_id: any): Observable<any> {
    const url = environment.URL + '/api/get-history?uuid=' + chat_id;
    return this.http.get(url);
  }
}
