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

/// <reference types="@rsbuild/core/types" />

declare const IS_OVERSEA: boolean;

interface FaviconConfig {
  light: string;
  dark: string;
}

interface DeployConfig {
  /** saas 海外版 私有化 */
  version: 'saas' | 'international' | 'custom';
  coze_url: string;
  mode: string;
  logo: string;
  title: string;
  appName: string;
  favicon: FaviconConfig;
  globalGreeting?: string;
}

export {};
declare global {
  interface Window {
    __CONFIG: {
      /** saas 海外版 私有化 */
      version: 'saas' | 'international' | 'custom';
      corekg_url: string;
      corekg_api_url: string;
      mode: 'default' | '';
      /** 是否隐藏工作流变量类型中的File选项 */
      hideFileType: boolean;
      /** 拉流分片之间的超时时间，单位分钟 */
      between_chunk_timeout_minutes?: number;
      /** 标签页图标，独立访问 / 跨域新窗口时使用 */
      favicon?: FaviconConfig;
    };
    /** AI 父页面部署配置，iframe 同域时读取 */
    __DEPLOYCONFIG?: DeployConfig;
  }
}