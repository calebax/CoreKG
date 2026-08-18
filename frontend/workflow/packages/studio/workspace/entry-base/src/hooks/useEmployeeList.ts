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

import { useState, useEffect } from 'react';

import { CoreKGApiService, type EmployeeInfo } from '@/services/corekg-api';

// 全局缓存员工列表，避免重复请求
let cachedEmployeeList: EmployeeInfo[] | null = null;
let isLoading = false;
const loadingCallbacks: Array<(data: EmployeeInfo[]) => void> = [];

/**
 * 获取员工列表的Hook
 * 用于在权限弹窗中显示用户名
 * 会在组件挂载时自动加载，并缓存结果
 */
export const useEmployeeList = () => {
  const [employeeList, setEmployeeList] = useState<EmployeeInfo[]>(
    cachedEmployeeList || [],
  );
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // 如果已经有缓存，直接使用
    if (cachedEmployeeList) {
      setEmployeeList(cachedEmployeeList);
      return;
    }

    // 如果正在加载中，等待加载完成
    if (isLoading) {
      setLoading(true);
      loadingCallbacks.push(data => {
        setEmployeeList(data);
        setLoading(false);
      });
      return;
    }

    // 开始加载
    isLoading = true;
    setLoading(true);

    CoreKGApiService.listEmployee()
      .then(res => {
        const list = res.Data || [];
        cachedEmployeeList = list;
        setEmployeeList(list);
        setLoading(false);
        isLoading = false;

        // 通知所有等待的回调
        loadingCallbacks.forEach(cb => cb(list));
        loadingCallbacks.length = 0;
      })
      .catch(error => {
        console.error('获取员工列表失败:', error);
        setLoading(false);
        isLoading = false;

        // 通知所有等待的回调
        loadingCallbacks.forEach(cb => cb([]));
        loadingCallbacks.length = 0;
      });
  }, []);

  // 根据uin获取用户名
  const getUserName = (uin: number): string => {
    const employee = employeeList.find(emp => emp.uin === uin);
    return employee?.user_name || `用户${uin}`;
  };

  // 刷新员工列表
  const refreshEmployeeList = async () => {
    setLoading(true);
    try {
      const res = await CoreKGApiService.listEmployee();
      const list = res.Data || [];
      cachedEmployeeList = list;
      setEmployeeList(list);
    } catch (error) {
      console.error('刷新员工列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  return {
    employeeList,
    loading,
    getUserName,
    refreshEmployeeList,
  };
};
