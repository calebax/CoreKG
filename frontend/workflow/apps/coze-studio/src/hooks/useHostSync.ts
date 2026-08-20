import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { I18n } from '@coze-arch/i18n';
import {
  AUTH_MESSAGE_TYPES,
  TOKEN_STORAGE_KEY,
  USER_INFO_STORAGE_KEY,
  isInIframe,
  sendReadySignal,
} from '@coze-foundation/auth-protocol';

const AGENT_HOME_PATH_PATTERN = /^\/space\/[^/]+\/develop\/?$/;

const isAgentHomePath = (pathname: string) =>
  AGENT_HOME_PATH_PATTERN.test(pathname);

const extractSpaceIdFromPath = (pathname: string): string | null => {
  const match = pathname.match(/^\/space\/([^/]+)/);
  return match?.[1] ?? null;
};

const postRouteStatus = (pathname: string) => {
  if (!isInIframe()) {
    return;
  }

  window.parent.postMessage(
    {
      type: AUTH_MESSAGE_TYPES.COZE_ROUTE_STATUS,
      payload: {
        isHome: isAgentHomePath(pathname),
        path: pathname,
      },
    },
    '*',
  );
};

export const useIframeSync = () => {
  const [isReady, setIsReady] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    postRouteStatus(location.pathname);
  }, [location.pathname]);

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      const { type, payload } = event.data || {};

      if (type === AUTH_MESSAGE_TYPES.SYNC_CONTEXT) {
        console.log('[Iframe] 收到父页面同步数据:', payload);

        try {
          if (payload.token) {
            localStorage.setItem(TOKEN_STORAGE_KEY, payload.token);
          }

          if (payload.userInfo) {
            localStorage.setItem(
              USER_INFO_STORAGE_KEY,
              JSON.stringify(payload.userInfo),
            );

            const locale = payload.userInfo.locale;
            if (locale === 'zh-CN' || locale === 'en-US') {
              const language = locale === 'en-US' ? 'en' : locale;
              localStorage.setItem('i18next', language);
              if (I18n.language !== language) {
                I18n.setLang(language);
              }
            }
          }

          setIsReady(true);
        } catch (e) {
          console.error('[Iframe] 写入 LocalStorage 失败', e);
        }
        return;
      }

      if (type === AUTH_MESSAGE_TYPES.AI_NAVIGATE_AGENT_HOME) {
        const spaceId = extractSpaceIdFromPath(window.location.pathname);
        navigate(
          spaceId ? `/space/${spaceId}/develop` : '/space',
          { replace: true },
        );
      }
    };

    window.addEventListener('message', handleMessage);

    if (isInIframe()) {
      console.log('[Iframe] 发送 READY 信号...');
      sendReadySignal();
    }

    return () => {
      window.removeEventListener('message', handleMessage);
    };
  }, [navigate]);

  return { isReady };
};
