import { TestBed } from '@angular/core/testing';
import { CanActivateFn } from '@angular/router';

import { chatAccessGuard } from './chat-access.guard';

describe('chatAccessGuard', () => {
  const executeGuard: CanActivateFn = (...guardParameters) => 
      TestBed.runInInjectionContext(() => chatAccessGuard(...guardParameters));

  beforeEach(() => {
    TestBed.configureTestingModule({});
  });

  it('should be created', () => {
    expect(executeGuard).toBeTruthy();
  });
});
