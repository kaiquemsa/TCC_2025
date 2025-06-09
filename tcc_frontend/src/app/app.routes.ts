import { Routes } from '@angular/router';
import { ChatComponent } from './pages/chat/chat.component';
import { AppComponent } from './app.component';
import { HomeComponent } from './pages/home/home.component';
import { ChatDefaultComponent } from './pages/chat-default/chat-default.component';
import { ChatAccessGuard } from './guards/chat-access.guard';

export const routes: Routes = [
  { path: '', component: ChatDefaultComponent },
  { path: 'c/:id', component: ChatComponent, canActivate: [ChatAccessGuard] },
];
