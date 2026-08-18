export { AUTH_MESSAGE_TYPES, TOKEN_STORAGE_KEY, USER_INFO_STORAGE_KEY } from './constants';
export type { AuthMessageType } from './constants';
export type {
  AuthMessage,
  SyncContextPayload,
  AuthErrorPayload,
  RouteStatusPayload,
  NavigateHomePayload,
} from './types';
export {
  getStoredToken,
  setStoredToken,
  removeStoredToken,
  getStoredUserInfo,
  setStoredUserInfo,
  removeStoredUserInfo,
  isInIframe,
} from './token';
export {
  postAuthMessage,
  notifyAuthError,
  requestRelogin,
  sendReadySignal,
} from './post-message';
