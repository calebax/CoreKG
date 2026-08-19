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

import { GlobalLayout, useAppInit } from '@coze-foundation/global-adapter';

import { useBrandingFavicon } from './hooks/useBrandingFavicon';
import { useIframeSync } from './hooks/useHostSync';

const InitializedLayout = () => {
  useBrandingFavicon();
  useAppInit();

  return <GlobalLayout />;
};

export const Layout = () => {
  const { isReady } = useIframeSync();

  const hasCachedToken = !!localStorage.getItem('coze_token');

  if (!isReady && !hasCachedToken) {
    return (
      <div className="w-full h-full flex items-center justify-center min-h-screen">
        <div>loading</div>
      </div>
    );
  }

  return <InitializedLayout />;
};
