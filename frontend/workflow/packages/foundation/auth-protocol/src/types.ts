import { AUTH_MESSAGE_TYPES } from './constants';

export interface SyncContextPayload {
  token: string;
  userInfo: Record<string, unknown> & { user_id_str: string };
  currentOrg?: Record<string, unknown>;
}

export interface AuthErrorPayload {
  code: number;
  message: string;
}

export interface RouteStatusPayload {
  isHome: boolean;
  path: string;
}

export interface NavigateHomePayload {
  homePath: string;
  iframeUrl: string;
}

export type AuthMessage =
  | { type: typeof AUTH_MESSAGE_TYPES.SYNC_CONTEXT; payload: SyncContextPayload }
  | { type: typeof AUTH_MESSAGE_TYPES.I_AM_READY }
  | {
      type: typeof AUTH_MESSAGE_TYPES.COZE_ROUTE_STATUS;
      payload: RouteStatusPayload;
    }
  | {
      type: typeof AUTH_MESSAGE_TYPES.AI_NAVIGATE_AGENT_HOME;
      payload: NavigateHomePayload;
    }
  | { type: typeof AUTH_MESSAGE_TYPES.AUTH_ERROR; payload: AuthErrorPayload }
  | { type: typeof AUTH_MESSAGE_TYPES.REQUEST_RELOGIN };
