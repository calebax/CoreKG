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
// 根据环境判断登录路径
// 本地开发时使用内部路由，生产环境使用外部地址
const isDevelopment = typeof window !== 'undefined' && (
  process.env.NODE_ENV === 'development' ||
  window.location.hostname === 'localhost' ||
  window.location.hostname === '127.0.0.1' ||
  window.location.hostname.startsWith('172.16.')
);

// export const signPath = isDevelopment ? '/sign' : 'https://app.corekg.com/';
const loginUrl = window.__CONFIG?.corekg_url || '';
export const signPath = isDevelopment ? '/sign' : loginUrl;
export const signRedirectKey = 'redirect';
