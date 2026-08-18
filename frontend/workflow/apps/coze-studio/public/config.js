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

// 需要和src/global.d.ts中的window.__CONFIG类型保持一致
// type Config = {
//   /** saas 海外版 私有化 */
//   version: 'saas' | 'international' | 'custom'
//   corekg_url: string
//   corekg_api_url: string
//   mode: 'default' | ''
//   /** 是否隐藏工作流变量类型中的File选项 */
//   hideFileType: boolean
//   /** 拉流分片之间的超时时间，单位分钟 */
//   between_chunk_timeout_minutes?: number
//   /** 标签页图标，独立访问 / 跨域新窗口时使用，路径与 AI 项目 config.js 保持一致 */
//   favicon: {
//     light: string
//     dark: string
//   }
// }

/* saas
window.__CONFIG = {
  version: 'saas',
  corekg_url: '',
  corekg_api_url: '',
  mode: '',
  hideFileType: false,
  between_chunk_timeout_minutes: 15,
  favicon: {
    light: '/icons/saas/favicon-light.ico',
    dark: '/icons/saas/favicon-dark.ico',
  },
};
*/

/* 默认私有化
window.__CONFIG = {
  version: 'custom',
  corekg_url: '',
  corekg_api_url: '',
  mode: 'default',
  hideFileType: true,
  between_chunk_timeout_minutes: 15,
  favicon: {
    light: '/icons/saas/favicon-light.ico',
    dark: '/icons/saas/favicon-dark.ico',
  },
};
*/

window.__CONFIG = {
  version: 'saas',
  corekg_url: '',
  corekg_api_url: '',
  mode: '',
  hideFileType: false,
  between_chunk_timeout_minutes: 15,
  favicon: {
    light: '/icons/saas/favicon-light.ico',
    dark: '/icons/saas/favicon-dark.ico',
  },
};