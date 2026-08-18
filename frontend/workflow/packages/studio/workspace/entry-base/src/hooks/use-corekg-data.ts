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

import { useState, useEffect, useCallback } from 'react';
import { CoreKGApiService, type FileNode, type FileChunk } from '../services/corekg-api';

// 树节点类型定义
export interface TreeNode {
  key: string;
  title: string;
  isLeaf: boolean;
  children?: TreeNode[];
  fileData?: FileNode;
}

// 解析树节点值的工具函数
export const parseTreeNodeValue = (value: string) => {
  try {
    return JSON.parse(value);
  } catch {
    return { fileId: value };
  }
};

/**
 * 文件树数据管理 Hook
 * @param forestId 知识库ID
 * @returns 文件树相关状态和方法
 */
export const useFileTree = (forestId: number) => {
  const [treeData, setTreeData] = useState<TreeNode[]>([]);
  const [loading, setLoading] = useState(false);
  const [expandedKeys, setExpandedKeys] = useState<string[]>([]);

  // 将 FileNode 转换为 TreeNode
  const convertToTreeNode = useCallback((fileNode: FileNode): TreeNode => {
    return {
      key: fileNode.ID.toString(),
      title: fileNode.name,
      isLeaf: !fileNode.is_dir,
      fileData: fileNode,
    };
  }, []);

  // 加载根级文件列表
  const loadRootFiles = useCallback(async () => {
    if (!forestId) return;
    
    setLoading(true);
    try {
      const response = await CoreKGApiService.getFileList({
        forest_id: forestId,
        limit: 100,
        offset: 0,
        filters: [{ field: 'parent_id', value: ['0'] }],
      });

      const nodes = response.data.map(convertToTreeNode);
      setTreeData(nodes);
    } catch (error) {
      console.error('Failed to load root files:', error);
    } finally {
      setLoading(false);
    }
  }, [forestId, convertToTreeNode]);

  // 加载子文件夹内容
  const loadChildrenFiles = useCallback(async (parentId: string) => {
    if (!forestId) return [];

    try {
      const response = await CoreKGApiService.getFileList({
        forest_id: forestId,
        limit: 100,
        offset: 0,
        filters: [{ field: 'parent_id', value: [parentId] }],
      });

      return response.data.map(convertToTreeNode);
    } catch (error) {
      console.error('Failed to load children files:', error);
      return [];
    }
  }, [forestId, convertToTreeNode]);

  // 处理树节点展开
  const onLoadData = useCallback(async (treeNode: any) => {
    const { key, children, fileData } = treeNode;
    
    if (children || !fileData?.is_dir) {
      return;
    }

    const childrenNodes = await loadChildrenFiles(key);
    
    setTreeData(prevData => {
      const updateNode = (nodes: TreeNode[]): TreeNode[] => {
        return nodes.map(node => {
          if (node.key === key) {
            return { ...node, children: childrenNodes };
          }
          if (node.children) {
            return { ...node, children: updateNode(node.children) };
          }
          return node;
        });
      };
      return updateNode(prevData);
    });
  }, [loadChildrenFiles]);

  // 初始化加载
  useEffect(() => {
    loadRootFiles();
  }, [loadRootFiles]);

  return {
    treeData,
    loading,
    expandedKeys,
    setExpandedKeys,
    onLoadData,
    refreshTree: loadRootFiles,
  };
};

/**
 * 文件内容分块管理 Hook
 * @param fileId 文件ID
 * @param forestId 知识库ID
 * @returns 文件分块相关状态和方法
 */
export const useFileChunks = (fileId?: number, forestId?: number) => {
  const [chunks, setChunks] = useState<FileChunk[]>([]);
  const [loading, setLoading] = useState(false);

  const loadFileChunks = useCallback(async () => {
    if (!fileId || !forestId) {
      setChunks([]);
      return;
    }

    setLoading(true);
    try {
      const response = await CoreKGApiService.getFileChunks({
        file_id: fileId,
        forest_id: forestId,
      });

      setChunks(response.chunks || []);
    } catch (error) {
      console.error('Failed to load file chunks:', error);
      setChunks([]);
    } finally {
      setLoading(false);
    }
  }, [fileId, forestId]);

  useEffect(() => {
    loadFileChunks();
  }, [loadFileChunks]);

  return {
    chunks,
    loading,
    refreshChunks: loadFileChunks,
  };
};

/**
 * 文件预览管理 Hook
 * @param fileId 文件ID
 * @returns 文件预览相关状态和方法
 */
export const useFilePreview = (fileId?: number) => {
  const [previewUrl, setPreviewUrl] = useState<string>('');
  const [loading, setLoading] = useState(false);

  const loadPreviewUrl = useCallback(async () => {
    if (!fileId) {
      setPreviewUrl('');
      return;
    }

    setLoading(true);
    try {
      const response = await CoreKGApiService.getPreviewFileURL({
        file_id: fileId,
      });

      setPreviewUrl(response.url || '');
    } catch (error) {
      console.error('Failed to load preview URL:', error);
      setPreviewUrl('');
    } finally {
      setLoading(false);
    }
  }, [fileId]);

  useEffect(() => {
    loadPreviewUrl();
  }, [loadPreviewUrl]);

  return {
    previewUrl,
    loading,
    refreshPreviewUrl: loadPreviewUrl,
  };
};

/**
 * CoreKG 综合数据管理 Hook
 * @param forestId 知识库ID
 * @returns 所有相关状态和方法
 */
export const useCoreKGData = (forestId: number) => {
  const [selectedFileId, setSelectedFileId] = useState<number>();
  const [selectedFileData, setSelectedFileData] = useState<FileNode>();

  const fileTree = useFileTree(forestId);
  const fileChunks = useFileChunks(selectedFileId, forestId);
  const filePreview = useFilePreview(selectedFileId);

  // 处理文件选择
  const handleFileSelect = useCallback((fileId: number, fileData?: FileNode) => {
    setSelectedFileId(fileId);
    setSelectedFileData(fileData);
  }, []);

  return {
    // 文件树相关
    ...fileTree,
    
    // 文件内容相关
    selectedFileId,
    selectedFileData,
    handleFileSelect,
    
    // 文件分块相关
    chunks: fileChunks.chunks,
    chunksLoading: fileChunks.loading,
    refreshChunks: fileChunks.refreshChunks,
    
    // 文件预览相关
    previewUrl: filePreview.previewUrl,
    previewLoading: filePreview.loading,
    refreshPreviewUrl: filePreview.refreshPreviewUrl,
  };
};