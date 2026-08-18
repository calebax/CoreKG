/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { useEffect, useState } from 'react';

export interface FaviconConfig {
  light: string;
  dark: string;
}

const DEFAULT_FAVICON: FaviconConfig = {
  light: '/icons/saas/favicon-light.ico',
  dark: '/icons/saas/favicon-dark.ico',
};

const isSameFavicon = (a: FaviconConfig, b: FaviconConfig) =>
  a.light === b.light && a.dark === b.dark;

const normalizeFavicon = (
  favicon?: Partial<FaviconConfig>,
  fallback: FaviconConfig = DEFAULT_FAVICON,
): FaviconConfig => ({
  light: favicon?.light || fallback.light,
  dark: favicon?.dark || fallback.dark,
});

export const resolveFaviconConfig = (): FaviconConfig => {
  const fallback = normalizeFavicon(window.__CONFIG?.favicon);

  try {
    if (window.parent !== window) {
      const parentFavicon = window.parent.__DEPLOYCONFIG?.favicon;
      if (parentFavicon?.light || parentFavicon?.dark) {
        return normalizeFavicon(parentFavicon, fallback);
      }
    }
  } catch {
    // Cross-origin iframe: fall back to local config.
  }

  return fallback;
};

export const useFaviconConfig = () => {
  const [favicon, setFavicon] = useState<FaviconConfig>(() =>
    resolveFaviconConfig(),
  );

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      const { type, payload } = event.data || {};
      if (type !== 'SYNC_CONTEXT' || !payload?.favicon) {
        return;
      }

      setFavicon(prev =>
        isSameFavicon(prev, payload.favicon) ? prev : payload.favicon,
      );
    };

    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  }, []);

  useEffect(() => {
    if (window.parent === window) {
      return;
    }

    const syncFromParent = () => {
      const next = resolveFaviconConfig();
      setFavicon(prev => (isSameFavicon(prev, next) ? prev : next));
    };

    syncFromParent();

    const intervalId = window.setInterval(syncFromParent, 1000);
    const timeoutId = window.setTimeout(() => {
      window.clearInterval(intervalId);
    }, 15000);

    return () => {
      window.clearInterval(intervalId);
      window.clearTimeout(timeoutId);
    };
  }, []);

  return favicon;
};
