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

import axios from 'axios';

// 文件节点接口定义
export interface FileNode {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  uin: number;
  company_id: number;
  forest_id: number;
  file_id: number;
  priview_file_id: number;
  priview_ext: string;
  is_dir: boolean;
  parent_id: number;
  name: string;
  size: number;
  ext: string;
  parent_ids: string;
  depth: number;
  parse_status: string;
  mindmap_status: string;
  analysis_status: string;
  knowledge_status: string;
  graph_status: string;
  desc_status: string;
  preview_able: string;
  status: string;
  file_config: {
    split_config: {
      split_mode: string;
      chunk_size: number;
      split_mark: string[];
      split_overlap: number;
      preprocessing_rules: {
        remove_empty_line: boolean;
        remove_url: boolean;
        remove_email: boolean;
      };
    };
  };
  file_status: string;
  file_progress: string;
  forest_type: string;
  data_source_type: string;
  data_source_subtype: string;
  // 添加type字段用于区分文件和文件夹
  type?: 'file' | 'folder';
  // 添加预览URL字段
  preview_tos_url?: string;
}

// 文件分块接口定义
export interface FileChunk {
  _id: string;
  _score: number;
  _source: {
    forest_id: number;
    company_id: number;
    uin: number;
    type: string;
    tokens: number;
    file_id: number;
    location: number[];
    description: string;
    is_disable: boolean;
  };
}

// API 响应接口定义
export interface ApiResponse<T> {
  code: number;
  env: string;
  request_id: string;
  Response: T;
}

export interface FileListResponse {
  total: number;
  offset: number;
  limit: number;
  data: FileNode[];
}

export interface FileChunksResponse {
  chunks: FileChunk[];
}

export interface PreviewUrlResponse {
  url: string;
}

// ==================== 资源权限相关接口定义 ====================

// 资源类型枚举
export enum ResourceScopeType {
  Agent = '4', // 智能体类型
  Plugin = '5', // 插件类型
  Workflow = '6', // 工作流类型
  Prompt = '17', // 提示词类型
}

// 用户信息接口
export interface UinInfo {
  name: string;
  uin: number;
}

// 资源权限范围接口
export interface ResourceScopeItem {
  manage_scope_ids: number[];
  resource_id: number;
  resource_id_str?: string; // 字符串版本的资源ID，避免大整数精度丢失
  resource_type: string;
  view_scope_ids: number[];
  view_scope_type: 'company' | 'user';
}

// 获取资源权限响应
export interface GetResourceScopeResponse {
  resource_scope_list: ResourceScopeItem[];
  uin_list: UinInfo[];
}

// 设置资源权限请求参数
export interface SetResourceScopeRequest {
  manage_scope_ids: number[];
  resource_id_str: string; // 使用字符串类型避免大整数精度丢失
  resource_type: string;
  view_scope_ids: number[];
  view_scope_type: 'company' | 'user';
}

// 员工信息接口
export interface EmployeeInfo {
  uin: number;
  user_name: string;
  ID?: number;
  id?: number;
  Name?: string;
  ParentID?: number;
  Children?: EmployeeInfo[];
}

// 员工列表响应
export interface ListEmployeeResponse {
  Data: EmployeeInfo[];
}

// 部门信息接口
export interface DepartmentInfo {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  Name: string;
  ParentID: number;
  Sort: number;
  CompanyID: number;
}

// 员工详细信息接口（用于部门树）
export interface EmployeeDetailInfo {
  uin: number;
  created_at: string;
  user_name: string;
  name: string;
  email: string;
  phone: string;
  employee_id: number;
  role: string;
  department_ids: number[] | null;
}

// 部门树响应接口
export interface DepartmentTreeResponse {
  departments: DepartmentInfo[];
  employees: EmployeeDetailInfo[];
}

const TOKEN_STORAGE_KEY = 'coze_token';

// 权限相关接口使用的通用发送函数（使用 /apis/p/ 前缀）
const sendPermissionApi = (url: string, data: any): Promise<any> => {
  const token =
    typeof window !== 'undefined' && window.localStorage
      ? window.localStorage.getItem(TOKEN_STORAGE_KEY)
      : undefined;

  const headers: any = {
    'Content-Type': 'application/json',
    Accept: 'application/json, text/plain, */*',
    'Accept-Language': 'zh-CN',
    'Cache-Control': 'no-cache',
    Env: 'prod',
    Pragma: 'no-cache',
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const body = {
    cmd: `/v3/${url}`,
    env: 'prod',
    version: 'v1.10.9-11',
    request: data,
  };

  return axios
    .post(`/apis/p/${url}`, body, {
      timeout: 60000,
      headers,
      withCredentials: true,
    })
    .then(response => {
      const { data } = response;
      if (data.code === 0) {
        return data.Response;
      } else {
        const msg = data.message || '调用接口失败';
        console.error('Permission API Error:', msg);
        return Promise.reject(new Error(msg));
      }
    })
    .catch(error => {
      let msg = '';
      if (error.response) {
        if (error.response?.data?.message) {
          msg = error.response.data.message;
        } else {
          msg = error.response.data;
        }
      } else {
        msg =
          error.message === 'Network Error'
            ? '网络未连接，请检查后重试'
            : error.message;
      }
      console.error('Permission API Error:', msg);
      return Promise.reject(new Error(msg));
    });
};

// 创建专门用于 CoreKG API 的 axios 实例
const corekgApi = axios.create({
  baseURL: process.env.REACT_APP_COREKG_API_URL,
  timeout: 600000000,
  headers: {
    'Content-Type': 'application/json',
    'Accept-Language': 'zh-CN',
    withCredentials: true,
    Env: 'prod',
  },
});
// 请求拦截器
corekgApi.interceptors.request.use(
  config => {
    // 从本地存储获取 token
    const token =
      localStorage.getItem('coze_token') || sessionStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  error => {
    return Promise.reject(error);
  },
);

// 响应拦截器
corekgApi.interceptors.response.use(
  response => {
    const { data } = response;
    if (data.code === 0) {
      return unwrapCoreKGPayload(data);
    } else {
      const msg = data.message || '调用接口失败';
      console.error('CoreKG API Error:', msg);
      return Promise.reject(new Error(msg));
    }
  },
  error => {
    let msg = '';
    if (error.response) {
      if (error.response?.data?.message) {
        msg = error.response.data.message;
      } else {
        msg = error.response.data;
      }
    } else {
      msg =
        error.message === 'Network Error'
          ? '网络未连接，请检查后重试'
          : error.message;
    }
    console.error('CoreKG API Error:', msg);
    return Promise.reject(new Error(msg));
  },
);

/** 从 CoreKG 接口响应中解析文件列表（兼容 Response 包装与扁平结构） */
export const normalizeFileListNodes = (raw: unknown): FileNode[] => {
  if (!raw) {
    return [];
  }
  if (Array.isArray(raw)) {
    return raw as FileNode[];
  }
  const obj = raw as Record<string, unknown>;
  if (Array.isArray(obj.data)) {
    return obj.data as FileNode[];
  }
  if (Array.isArray(obj.Data)) {
    return obj.Data as FileNode[];
  }
  const nested = (obj.Response ?? obj.response) as Record<string, unknown> | undefined;
  if (nested && Array.isArray(nested.data)) {
    return nested.data as FileNode[];
  }
  return [];
};

const unwrapCoreKGPayload = (data: Record<string, unknown>): unknown => {
  if (data.Response !== undefined) {
    return data.Response;
  }
  if (data.response !== undefined) {
    return data.response;
  }
  // 扁平结构：{ code, total, data, chunks, url, ... }
  const { code: _code, msg: _msg, env: _env, request_id: _requestId, ...rest } =
    data;
  return Object.keys(rest).length > 0 ? rest : data;
};

// 通用发送函数，模拟 roc-web/ai 的 send 函数
const send = (url: string, data: any): Promise<any> => {
  // 使用 /v2/ 路径的接口
  const v2Apis = ['account.ListEmployeeNickID', 'account.GetDepartmentTree'];

  // 使用 /v3/ 路径的接口
  const fixedDomainApis = [
    'forest.ListFile',
    'kesearch.ListFileChunk',
    'forest.PreviewFileByURL',
    'forest.GetResourceScope',
    'forest.SetResourceScope',
  ];

  // 动态获取Authorization token
  const token =
    typeof window !== 'undefined' && window.localStorage
      ? window.localStorage.getItem(TOKEN_STORAGE_KEY)
      : undefined;

  const headers: any = {
    'Content-Type': 'application/json',
    Accept: 'application/json, text/plain, */*',
    'Accept-Language': 'zh-CN',
    'Cache-Control': 'no-cache',
    Env: 'prod',
    Pragma: 'no-cache',
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  // 处理使用 /v2/ 路径的接口
  if (v2Apis.includes(url)) {
    const v2Body = {
      cmd: `/v2/${url}`,
      env: 'prod',
      version: 'v1.0.0-1469',
      request: data,
    };

    return axios
      .post(`/v2/${url}`, v2Body, {
        timeout: 600000000,
        headers,
        withCredentials: true,
      })
      .then(response => {
        const { data } = response;
        if (data.code === 0) {
          return data.Response || data.response;
        } else {
          const msg = data.message || '调用接口失败';
          console.error('CoreKG API Error:', msg);
          return Promise.reject(new Error(msg));
        }
      })
      .catch(error => {
        let msg = '';
        if (error.response) {
          if (error.response?.data?.message) {
            msg = error.response.data.message;
          } else {
            msg = error.response.data;
          }
        } else {
          msg =
            error.message === 'Network Error'
              ? '网络未连接，请检查后重试'
              : error.message;
        }
        console.error('CoreKG API Error:', msg);
        return Promise.reject(new Error(msg));
      });
  }

  // 处理使用 /v3/ 路径的接口
  if (fixedDomainApis.includes(url)) {
    const body = {
      cmd: `/v3/${url}`,
      env: 'prod',
      version: 'v1.10.9-11',
      request: data,
    };

    // 使用nginx代理路径
    return axios
      .post(`/v3/${url}`, body, {
        timeout: 600000000,
        headers,
        withCredentials: true,
      })
      .then(response => {
        const { data } = response;
        if (data.code === 0) {
          return unwrapCoreKGPayload(data);
        } else {
          const msg = data.message || '调用接口失败';
          console.error('CoreKG API Error:', msg);
          return Promise.reject(new Error(msg));
        }
      })
      .catch(error => {
        let msg = '';
        if (error.response) {
          if (error.response?.data?.message) {
            msg = error.response.data.message;
          } else {
            msg = error.response.data;
          }
        } else {
          msg =
            error.message === 'Network Error'
              ? '网络未连接，请检查后重试'
              : error.message;
        }
        console.error('CoreKG API Error:', msg);
        return Promise.reject(new Error(msg));
      });
  }

  // 其他接口使用 corekgApi
  const body = {
    cmd: `/v3/${url}`,
    env: 'prod',
    version: 'v1.10.9-11',
    request: data,
  };

  return corekgApi.post(`/v3/${url}`, body);
};

// CoreKG API 服务类
export class CoreKGApiService {
  /**
   * 获取知识库下的文件列表（树状结构）
   * @param params 请求参数
   * @returns 文件列表
   */
  static async getFileList(params: {
    forest_id: number;
    parent_id?: string;
    limit?: number;
    offset?: number;
    orderBy?: string[];
    filters?: Array<{
      field: string;
      value: string[];
    }>;
  }): Promise<FileListResponse> {
    const defaultParams = {
      forest_id: params.forest_id,
      limit: params.limit || 100,
      offset: params.offset || 0,
      orderBy: params.orderBy || ['updated_at desc'],
      filters: params.filters || [
        { field: 'parent_id', value: [params.parent_id || '0'] },
      ],
    };

    const raw = await send('forest.ListFile', defaultParams);
    const files = normalizeFileListNodes(raw);
    return {
      total: (raw as FileListResponse)?.total ?? files.length,
      offset: (raw as FileListResponse)?.offset ?? 0,
      limit: (raw as FileListResponse)?.limit ?? defaultParams.limit,
      data: files,
    };
  }

  /**
   * 获取文件内容分块
   * @param params 请求参数
   * @returns 文件分块列表
   */
  static async getFileChunks(params: {
    file_id: number;
    forest_id: number;
  }): Promise<FileChunksResponse> {
    return send('kesearch.ListFileChunk', params);
  }

  /**
   * 获取文件预览下载链接
   * @param params 请求参数
   * @returns 预览链接
   */
  static async getPreviewFileURL(params: {
    file_id: number;
  }): Promise<PreviewUrlResponse> {
    return send('forest.PreviewFileByURL', params);
  }

  // ==================== 资源权限相关 API ====================

  /**
   * 获取资源权限范围
   * @param params 请求参数
   * @returns 资源权限范围信息
   */
  static async getResourceScope(params: {
    resource_id_strs: string[]; // 使用字符串数组避免大整数精度丢失
    resource_type: string;
  }): Promise<GetResourceScopeResponse> {
    return send('forest.GetResourceScope', params);
  }

  /**
   * 设置资源权限范围
   * @param params 请求参数
   * @returns 无返回值
   */
  static async setResourceScope(
    params: SetResourceScopeRequest,
  ): Promise<void> {
    return send('forest.SetResourceScope', params);
  }

  /**
   * 获取员工列表（用于权限选择弹窗）
   * @returns 员工列表
   */
  static async listEmployee(): Promise<ListEmployeeResponse> {
    return send('account.ListEmployeeNickID', { listAll: true });
  }

  /**
   * 获取部门树和员工信息（用于人员选择树）
   * @param params 请求参数
   * @returns 部门树和员工列表
   */
  static async getDepartmentTree(params: {
    include_employee: boolean;
  }): Promise<DepartmentTreeResponse> {
    return send('account.GetDepartmentTree', params);
  }
}

export default CoreKGApiService;
