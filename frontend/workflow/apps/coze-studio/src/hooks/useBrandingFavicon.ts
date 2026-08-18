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
import { useFavicon } from 'ahooks';

import { useFaviconConfig } from './useFaviconConfig';

const getBodyThemeMode = () =>
  document.body.getAttribute('theme-mode') === 'dark' ? 'dark' : 'light';

export const useBrandingFavicon = () => {
  const favicon = useFaviconConfig();
  const [themeMode, setThemeMode] = useState(getBodyThemeMode);

  useEffect(() => {
    const syncThemeMode = () => {
      setThemeMode(prev => {
        const next = getBodyThemeMode();
        return prev === next ? prev : next;
      });
    };

    syncThemeMode();

    const observer = new MutationObserver(syncThemeMode);
    observer.observe(document.body, {
      attributes: true,
      attributeFilter: ['theme-mode'],
    });

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', syncThemeMode);

    return () => {
      observer.disconnect();
      mediaQuery.removeEventListener('change', syncThemeMode);
    };
  }, []);

  const faviconUrl = themeMode === 'dark' ? favicon.dark : favicon.light;

  useFavicon(faviconUrl);
};
