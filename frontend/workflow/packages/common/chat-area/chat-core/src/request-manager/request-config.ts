/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios';

import { type DefaultRequestManagerOptions, RequestScene } from './types';
import { ApiError } from './api-error';

const useApiErrorResponseHook = (response: AxiosResponse) => {
  const { data = {} } = response;
  const { code, msg } = data;
  if (code !== 0) {
    const apiError = new ApiError(String(code), msg, response);
    return Promise.reject(apiError);
  }
  return response;
};

const injectCozePrefix = (url: string): string => {
  try {
    const apiRe = /^(\/)(api|v1|v2|v3|corekg-bucket)(\/|$)/;
    const absMatch = url.match(/^(https?:\/\/[^/]+)(\/.*)$/);
    if (absMatch) {
      const origin = absMatch[1];
      const path = absMatch[2];
      if (apiRe.test(path) && !path.startsWith('/coze/')) {
        return origin + '/coze' + path;
      }
      return url;
    }
    if (apiRe.test(url) && !url.startsWith('/coze/')) {
      return '/coze' + url;
    }
    return url;
  } catch {
    return url;
  }
};

const useCommonRequestHook = (config: InternalAxiosRequestConfig) => {
  config.headers.set('x-requested-with', 'XMLHttpRequest');

  if (
    config.method?.toLowerCase() === 'post' &&
    !config.headers.get('content-type')
  ) {
    config.headers.set('content-type', 'application/json');
    if (!config.data) {
      config.data = {};
    }
  }

  // [修改] 直接读取 coze_token 的值作为 Bearer Token
  const token = localStorage.getItem('coze_token');
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`);
  }

  // 统一为 /api、/v1、/v2、/v3、/corekg-bucket 路径补 /coze 前缀
  if (config.url) {
    config.url = injectCozePrefix(config.url);
  }

  return config;
};

export const getDefaultSceneConfig = (): DefaultRequestManagerOptions => ({
  hooks: {
    onBeforeRequest: [useCommonRequestHook],
    onAfterResponse: [useApiErrorResponseHook],
  },
  scenes: {
    [RequestScene.SendMessage]: {
      url: '/api/conversation/chat',
      method: 'POST',
      hooks: {
        onBeforeSendMessage: [sendMessageConfig => {
          sendMessageConfig.url = injectCozePrefix(sendMessageConfig.url);
          const token = localStorage.getItem('coze_token');
          if (token && !sendMessageConfig.headers.some(([key]) => key.toLowerCase() === 'authorization')) {
            sendMessageConfig.headers.push(['Authorization', `Bearer ${token}`]);
          }
          return sendMessageConfig;
        }],
      },
    },
    [RequestScene.ResumeMessage]: {
      url: '/api/conversation/resume_chat',
      method: 'POST',
      hooks: {
        onBeforeSendMessage: [sendMessageConfig => {
          sendMessageConfig.url = injectCozePrefix(sendMessageConfig.url);
          const token = localStorage.getItem('coze_token');
          if (token && !sendMessageConfig.headers.some(([key]) => key.toLowerCase() === 'authorization')) {
            sendMessageConfig.headers.push(['Authorization', `Bearer ${token}`]);
          }
          return sendMessageConfig;
        }],
      },
    },
    [RequestScene.GetMessage]: {
      url: '/api/conversation/get_message_list',
      method: 'POST',
    },
    [RequestScene.ClearHistory]: {
      url: '/api/conversation/clear_message',
      method: 'POST',
    },
    [RequestScene.ClearMessageContext]: {
      url: '/api/conversation/create_section',
      method: 'POST',
    },
    [RequestScene.DeleteMessage]: {
      url: '/api/conversation/delete_message',
      method: 'POST',
    },
    [RequestScene.BreakMessage]: {
      url: '/api/conversation/break_message',
      method: 'POST',
    },
    [RequestScene.ReportMessage]: {
      url: '/api/conversation/message/report',
      method: 'POST',
    },
    [RequestScene.ChatASR]: {
      url: '/api/audio/transcriptions',
      method: 'POST',
    },
  },
});