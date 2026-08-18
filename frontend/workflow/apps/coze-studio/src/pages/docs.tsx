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

import { useEffect } from 'react';

const COZE_ORIGIN = 'https://www.coze.cn';

const DocsRedirect = () => {
  useEffect(() => {
    const basePath = (import.meta.env.BASE_URL || '/').replace(/\/$/, '');
    let pathname = location.pathname;
    if (basePath && basePath !== '/' && pathname.startsWith(basePath)) {
      pathname = pathname.slice(basePath.length) || '/';
    }
    location.href = `${COZE_ORIGIN}${pathname}${location.search}${location.hash}`;
  }, []);
  return null;
};

export default DocsRedirect;
