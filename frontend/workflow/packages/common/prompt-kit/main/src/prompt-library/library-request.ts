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

import { type ResourceAction, ResType } from '@coze-arch/idl/plugin_develop';
import { PlaygroundApi, PluginDevelopApi } from '@coze-arch/bot-api';
export interface LibraryInfo {
  id: string;
  name: string;
  description: string;
  actions?: ResourceAction[];
  promptText?: string;
}
export interface LibraryListRequest {
  searchWord: string;
  cursor: string;
  category: 'Recommended' | 'Team';
  spaceId: string;
  size: number;
}
export interface LibraryListResponse {
  list: LibraryInfo[];
  hasMore: boolean;
  cursor: string;
  code: number;
  [key: string]: unknown;
}

export const getTeamLibraryRequest = async (req: LibraryListRequest) => {
  const res = await PluginDevelopApi.LibraryResourceList({
    space_id: req.spaceId,
    size: req.size,
    cursor: req.cursor,
    name: req.searchWord,
    search_keys: ['full_text'],
    res_type_filter: [ResType.Prompt],
  });
  return {
    list:
      res.resource_list?.map(item => ({
        id: item.res_id ?? '',
        name: item.name ?? '',
        description: item.desc ?? '',
        actions: item?.actions ?? [],
      })) ?? [],
    hasMore: res.has_more ?? false,
    cursor: res.cursor ?? '',
    code: Number(res.code) ?? 0,
  };
};

export const getRecommendLibraryRequest = async (req: LibraryListRequest) => {
  // 根据 cursor 判断是否为首次加载
  const isFirstLoad = req.cursor === '0'

  if (isFirstLoad) {
    // 首次加载：同时调用两个接口，合并结果
    const [officialRes, libraryRes] = await Promise.all([
      PlaygroundApi.GetOfficialPromptResourceList({
        keyword: req.searchWord,
      }),
      PluginDevelopApi.LibraryResourceList({
        space_id: req.spaceId,
        size: req.size,
        cursor: req.cursor,
        name: req.searchWord,
        search_keys: ['full_text'],
        res_type_filter: [ResType.Prompt],
      }),
    ])

    // 合并两个列表，官方列表在前，个人列表在后
    const officialList =
      officialRes.data?.map(item => ({
        id: item.id ?? '',
        name: item.name ?? '',
        description: item.description ?? '',
        promptText: item.prompt_text ?? '',
      })) ?? []

    const libraryList =
      libraryRes.resource_list?.map(item => ({
        id: item.res_id ?? '',
        name: item.name ?? '',
        description: item.desc ?? '',
        actions: item?.actions ?? [],
      })) ?? []

    return {
      list: [...officialList, ...libraryList],
      hasMore: libraryRes.has_more ?? false,
      cursor: libraryRes.cursor ?? '0',
      code: Number(officialRes.code) ?? 0,
    }
  } else {
    // 后续加载：只加载个人列表的下一页
    const libraryRes = await PluginDevelopApi.LibraryResourceList({
      space_id: req.spaceId,
      size: req.size,
      cursor: req.cursor,
      name: req.searchWord,
      search_keys: ['full_text'],
      res_type_filter: [ResType.Prompt],
    })

    const libraryList =
      libraryRes.resource_list?.map(item => ({
        id: item.res_id ?? '',
        name: item.name ?? '',
        description: item.desc ?? '',
        actions: item?.actions ?? [],
      })) ?? []

    return {
      list: libraryList,
      hasMore: libraryRes.has_more ?? false,
      cursor: libraryRes.cursor ?? '0',
      code: Number(libraryRes.code) ?? 0,
    }
  }
}

export const getLibraryListByCategory = (
  req: LibraryListRequest,
): Promise<LibraryListResponse> => {
  if (req.category === 'Team') {
    return getTeamLibraryRequest(req);
  }
  return getRecommendLibraryRequest(req);
};
