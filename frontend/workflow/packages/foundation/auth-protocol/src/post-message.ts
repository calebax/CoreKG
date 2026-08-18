import { AUTH_MESSAGE_TYPES } from './constants';
import type { AuthMessage, AuthErrorPayload } from './types';

export function postAuthMessage(
  target: Window,
  message: AuthMessage,
  targetOrigin: string = '*',
): void {
  target.postMessage(message, targetOrigin);
}

export function notifyAuthError(
  payload: AuthErrorPayload,
  targetOrigin: string = '*',
): void {
  if (window.parent === window) return;
  postAuthMessage(
    window.parent,
    { type: AUTH_MESSAGE_TYPES.AUTH_ERROR, payload },
    targetOrigin,
  );
}

export function requestRelogin(targetOrigin: string = '*'): void {
  if (window.parent === window) return;
  postAuthMessage(
    window.parent,
    { type: AUTH_MESSAGE_TYPES.REQUEST_RELOGIN },
    targetOrigin,
  );
}

export function sendReadySignal(targetOrigin: string = '*'): void {
  if (window.parent === window) return;
  postAuthMessage(
    window.parent,
    { type: AUTH_MESSAGE_TYPES.I_AM_READY },
    targetOrigin,
  );
}
