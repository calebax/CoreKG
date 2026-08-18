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

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';

import classnames from 'classnames';
import { I18n } from '@coze-arch/i18n';
import {
  Typography,
  EmptyState,
  Spin,
  TreeSelect,
  TextArea,
} from '@coze-arch/coze-design';

import { IconSegmentEmpty } from '@coze-arch/bot-icons';
import {
  DocumentSource,
  FormatType,
  type DocumentInfo,
  type Dataset,
} from '@coze-arch/bot-api/knowledge';
import {
  useKnowledgeStore,
  useKnowledgeParams,
} from '@coze-data/knowledge-stores';
import { SegmentMenu } from '@coze-data/knowledge-common-components';

import { type ProgressMap } from '@/types';
import {
  CoreKGApiService,
  type FileNode,
  type FileChunk,
  type FileChunksResponse,
  type PreviewUrlResponse,
} from '../../../../../../../studio/workspace/entry-base/src/services/corekg-api';
import { TextToolbar } from '../../text-knowledge-workspace/components/text-toolbar';
import { FilePreview } from '../../text-knowledge-workspace/components/file-preview';
import { getDocumentOptions } from '../../text-knowledge-workspace/utils/document-utils';
import { buildDisplayDocumentList } from '../utils/build-display-document-list';
import {
  resolvePreviewFileType,
  resolveSourceFileType,
} from '../utils/resolve-preview-file-type';

export interface CoreKGTextKnowledgeWorkspaceProps {
  progressMap?: ProgressMap;
  reload?: () => void;
}

interface TreeNode {
  key: string;
  title: string;
  isLeaf?: boolean;
  children?: TreeNode[];
  fileData?: FileNode;
}

// 分段数据接口
interface FileSegment {
  id: string;
  content: string;
  chunk_number: number;
  charCount: number;
}

const { Text } = Typography;

export const CoreKGTextKnowledgeWorkspace: React.FC<
  CoreKGTextKnowledgeWorkspaceProps
> = ({
  progressMap = {},
  reload,
}) => {
  const [searchParams] = useSearchParams();
  const forestIdParam = searchParams.get('forestId') || '';
  const [leftWidth, setLeftWidth] = useState(300);
  const [isDragging, setIsDragging] = useState(false);

  // CoreKG 相关状态
  const [forestFiles, setForestFiles] = useState<FileNode[]>([]);
  const [treeData, setTreeData] = useState<any[]>([]);
  const [selectedNode, setSelectedNode] = useState<string | undefined>();
  const [selectedFileData, setSelectedFileData] = useState<FileNode | null>(
    null,
  );
  const [chunks, setChunks] = useState<FileChunk[]>([]);
  const [previewUrl, setPreviewUrl] = useState<string>('');
  const [chunksLoading, setChunksLoading] = useState(false);
  const [treeLoading, setTreeLoading] = useState(false);

  // TextToolbar 相关状态 - 修改为预览模式状态
  const [showOriginalFile, setShowOriginalFile] = useState(false);

  // 新增：分段数据状态
  const [segments, setSegments] = useState<FileSegment[]>([]);

  const { dataSetDetail, documentList, curDocId, setCurDocId } = useKnowledgeStore(
    state => ({
      dataSetDetail: state.dataSetDetail,
      documentList: state.documentList,
      curDocId: state.curDocId,
      setCurDocId: state.setCurDocId,
    }),
  );
  const { biz, datasetID = '' } = useKnowledgeParams();

  const displayDocumentList = useMemo(
    () =>
      buildDisplayDocumentList(
        documentList,
        forestFiles,
        dataSetDetail,
      ),
    [documentList, forestFiles, dataSetDetail],
  );

  const docOptions = useMemo(
    () => getDocumentOptions(displayDocumentList, progressMap),
    [displayDocumentList, progressMap],
  );
  const useDocumentListMode = displayDocumentList.length > 0;
  const fromProject = biz === 'project';

  // 基于 URL、corekg_detail_id 或 dataset_id 计算 forest_id（与知识库列表跳转逻辑一致）
  const forestIdNum = useMemo(() => {
    const fromURL = parseInt(forestIdParam, 10);
    if (!isNaN(fromURL) && fromURL > 0) {
      return fromURL;
    }
    const coreId = Number(
      (dataSetDetail as Dataset & { corekg_detail_id?: number })
        ?.corekg_detail_id,
    );
    if (!isNaN(coreId) && coreId > 0) {
      return coreId;
    }
    const fromRouteDatasetId = parseInt(datasetID, 10);
    if (!isNaN(fromRouteDatasetId) && fromRouteDatasetId > 0) {
      return fromRouteDatasetId;
    }
    const fromDetailDatasetId = parseInt(
      String(dataSetDetail?.dataset_id ?? ''),
      10,
    );
    if (!isNaN(fromDetailDatasetId) && fromDetailDatasetId > 0) {
      return fromDetailDatasetId;
    }
    return NaN;
  }, [forestIdParam, dataSetDetail, datasetID]);

  const setChunkData = useCallback((chunkList: FileChunk[]) => {
    setChunks(chunkList);
    setSegments(
      chunkList.map((chunk: FileChunk) => ({
        id: chunk._id,
        chunk_number: chunk._source.sequence || 0,
        content: chunk._source.description || '',
        charCount: chunk._source.chunk_size || chunk._source.tokens,
      })),
    );
  }, []);

  const clearFileContent = useCallback(() => {
    setChunkData([]);
    setPreviewUrl('');
  }, [setChunkData]);

  const loadSelectedFileContent = useCallback(
    async (fileData: FileNode) => {
      setSelectedFileData(fileData);
      setPreviewUrl(fileData.preview_tos_url || '');

      setChunksLoading(true);
      try {
        const res = await CoreKGApiService.getFileChunks({
          file_id: fileData.ID,
          forest_id: forestIdNum,
        });
        const chunkList =
          (res as FileChunksResponse)?.chunks ??
          (Array.isArray(res) ? res : []);
        setChunkData(chunkList);
      } catch (error) {
        console.error('Failed to load file chunks:', error);
        setChunkData([]);
      } finally {
        setChunksLoading(false);
      }

      try {
        const previewRes = await CoreKGApiService.getPreviewFileURL({
          file_id: fileData.ID,
        });
        const url =
          (previewRes as PreviewUrlResponse)?.url ??
          (previewRes as { url?: string })?.url ??
          '';
        setPreviewUrl(url);
        setSelectedFileData(prev =>
          prev && prev.ID === fileData.ID ? { ...prev, preview_tos_url: url } : prev,
        );
      } catch (error) {
        console.error('Failed to get preview URL:', error);
        setPreviewUrl('');
      }
    },
    [forestIdNum, setChunkData],
  );

  // 根据文档 ID 加载分块与预览（document_id 与 CoreKG file_id 对应）
  const loadDocumentContent = useCallback(
    async (docId: string, doc?: DocumentInfo) => {
      const fileId = Number(docId);
      if (!docId || isNaN(fileId) || fileId <= 0 || isNaN(forestIdNum)) {
        return;
      }

      const fileName = doc?.name ?? '';
      const sourceFileType = resolveSourceFileType({ doc, forestFiles });
      const previewFileTypeForNode = resolvePreviewFileType({ doc, forestFiles });
      const fileData: FileNode = {
        ID: fileId,
        name: fileName,
        is_dir: false,
        size: Number(doc?.size ?? 0),
        ext: sourceFileType ? `.${sourceFileType}` : '',
        priview_ext: previewFileTypeForNode
          ? `.${previewFileTypeForNode}`
          : '',
      } as FileNode;
      await loadSelectedFileContent(fileData);
    },
    [forestFiles, forestIdNum, loadSelectedFileContent],
  );

  // 加载文件树（无 documentList 时的兜底）
  const loadFileTree = useCallback(
    async (parentId?: string) => {
      if (isNaN(forestIdNum) || forestIdNum <= 0) {
        console.error('Invalid forestId:', forestIdParam);
        return [];
      }

      try {
        const res = await CoreKGApiService.getFileList({
          forest_id: forestIdNum,
          parent_id: parentId || '0',
        });
        const files = res?.data ?? [];
        if (!parentId || parentId === '0') {
          setForestFiles(files);
        }
        return files.map((file: FileNode) => ({
          key: `${parentId || '0'}-${file.ID}`,
          label: file.name,
          value: file.name,
          title: file.name,
          isLeaf: !file.is_dir,
          fileData: file,
        }));
      } catch (error) {
        console.error('Failed to load file tree:', error);
        return [];
      }
    },
    [forestIdNum, forestIdParam],
  );

  // 始终拉取 CoreKG 文件列表，并与 list 接口结果合并展示
  useEffect(() => {
    const loadRootFiles = async () => {
      if (isNaN(forestIdNum) || forestIdNum <= 0) {
        return;
      }
      setTreeLoading(true);
      try {
        const rootFiles = await loadFileTree();
        setTreeData(rootFiles);
      } catch (error) {
        console.error('Failed to load root files:', error);
        setTreeData([]);
        setForestFiles([]);
      } finally {
        setTreeLoading(false);
      }
    };
    loadRootFiles();
  }, [forestIdNum, loadFileTree]);

  useEffect(() => {
    if (!useDocumentListMode || isNaN(forestIdNum)) {
      return;
    }
    const targetDocId =
      curDocId || displayDocumentList[0]?.document_id;
    if (!targetDocId) {
      return;
    }
    if (!curDocId) {
      setCurDocId(targetDocId);
    }
    const doc = displayDocumentList.find(
      item => item.document_id === targetDocId,
    );
    loadDocumentContent(targetDocId, doc);
  }, [
    useDocumentListMode,
    displayDocumentList,
    curDocId,
    forestIdNum,
    loadDocumentContent,
    setCurDocId,
  ]);

  const handleSelectFromDocumentList = useCallback(
    (docId: string) => {
      setCurDocId(docId);
    },
    [setCurDocId],
  );

  // 处理 TreeSelect 文档选择（兜底模式）
  const handleDocumentSelect = useCallback(
    async (
      value: string | number | (string | number | undefined)[] | undefined,
    ) => {
      if (!value || typeof value !== 'string') return;

      setSelectedNode(value);

      const findFileData = (
        nodes: any[],
        targetValue: string,
      ): FileNode | null => {
        for (const node of nodes) {
          if (node.value === targetValue) {
            return node.fileData;
          }
          if (node.children) {
            const found = findFileData(node.children, targetValue);
            if (found) return found;
          }
        }
        return null;
      };

      const fileData = findFileData(treeData, value);

      // 只有当选择的是文件（非文件夹）时才加载分块数据和预览URL
      if (fileData && !fileData.is_dir) {
        await loadSelectedFileContent(fileData);
      } else {
        // 如果选择的是文件夹，清空分块数据和预览URL
        setSelectedFileData(fileData);
        clearFileContent();
      }
    },
    [clearFileContent, loadSelectedFileContent, treeData],
  );

  // 动态加载子节点
  const onLoadData = useCallback(
    async (node: any) => {
      if (node.isLeaf || node.children) {
        return;
      }

      // 从node.fileData中获取实际的ID来加载子节点
      const parentId = node.fileData ? node.fileData.ID.toString() : node.key;
      const children = await loadFileTree(parentId);

      setTreeData(prevData => {
        const updateNode = (nodes: any[]): any[] => {
          return nodes.map(item => {
            if (item.key === node.key) {
              return { ...item, children };
            }
            if (item.children) {
              return { ...item, children: updateNode(item.children) };
            }
            return item;
          });
        };
        return updateNode(prevData);
      });
    },
    [loadFileTree],
  );

  // TextToolbar 回调函数 - 修改为预览模式切换
  const handleToggleOriginalFile = useCallback((checked: boolean) => {
    setShowOriginalFile(checked);
  }, []);

  // 渲染分段列表
  const renderSegmentList = () => (
    <div className="h-full">
      <div className="space-y-3 h-full overflow-auto p-4">
        {segments.map(segment => (
          <div key={segment.id} className="space-y-2">
            {/* 文本内容框 */}
            <div className="border border-[#f3f4f8] bg-[#f9f9fc] rounded-md leading-5 text-sm w-full">
              <TextArea
                value={segment.content}
                readOnly
                autoSize={{ minRows: 1 }}
                style={{
                  padding: '6px',
                  resize: 'none',
                  cursor: 'default',
                  backgroundColor: 'transparent',
                  border: 'none',
                }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );

  const curDocFromList = useMemo(
    () => displayDocumentList.find(item => item.document_id === curDocId),
    [displayDocumentList, curDocId],
  );

  const previewFileType = useMemo(
    () =>
      resolvePreviewFileType({
        fileData: selectedFileData,
        doc: curDocFromList,
        forestFiles,
      }),
    [selectedFileData, curDocFromList, forestFiles],
  );

  // 构造 TextToolbar 所需的 props
  const documentData = useMemo(
    () => ({
      curDoc:
        curDocFromList ??
        (selectedFileData
          ? ({
              document_id: selectedFileData.ID.toString(),
              name: selectedFileData.name,
              size: selectedFileData.size,
              source_type: DocumentSource.Document,
              format_type:
                dataSetDetail?.format_type ??
                curDocFromList?.format_type ??
                FormatType.Text,
            } as DocumentInfo)
          : undefined),
      curDocId: curDocId || selectedFileData?.ID?.toString() || '',
      curFormatType:
        curDocFromList?.format_type ??
        dataSetDetail?.format_type ??
        FormatType.Text,
      docOptions: useDocumentListMode
        ? docOptions
        : selectedFileData?.name
          ? [
              {
                label: selectedFileData.name,
                value: selectedFileData.name,
                key: selectedFileData.ID.toString(),
              },
            ]
          : [],
    }),
    [
      curDocFromList,
      selectedFileData,
      curDocId,
      useDocumentListMode,
      docOptions,
      dataSetDetail?.format_type,
    ],
  );

  const filePreviewData = {
    showOriginalFile,
    fileUrl: previewUrl,
  };

  const documentActions = useMemo(
    () => ({
      onChangeDoc: handleDocumentSelect,
      onRenameDoc: () => {},
      onToggleOriginalFile: handleToggleOriginalFile,
      onResegment: () => {},
      onUpdateFrequency: () => {},
      onDelete: () => {},
      reloadDataset: reload,
    }),
    [handleDocumentSelect, handleToggleOriginalFile, reload],
  );

  const customUIElements = {
    linkOriginUrlButton: undefined,
    fetchSliceButton: undefined,
  };

  // 处理拖拽调整宽度
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    setIsDragging(true);
    e.preventDefault();
  }, []);

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isDragging) return;

      // 获取容器元素
      const container = document.querySelector(
        '.flex.grow.border-solid.coz-stroke-primary.coz-bg-max',
      );
      if (!container) return;

      // 计算相对于容器的鼠标位置
      const containerRect = container.getBoundingClientRect();
      const newWidth = e.clientX - containerRect.left;

      // 设置合理的宽度范围
      if (newWidth >= 300 && newWidth <= containerRect.width - 1000) {
        setLeftWidth(newWidth);
      }
    },
    [isDragging],
  );

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);

  // 添加全局鼠标事件监听
  useEffect(() => {
    if (isDragging) {
      // 拖拽时禁用过渡效果
      const leftPanel = document.querySelector(
        '.h-full.shrink-0.overflow-auto.p-\\[12px\\]',
      );
      if (leftPanel) {
        (leftPanel as HTMLElement).style.transition = 'none';
      }

      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);

      return () => {
        // 恢复过渡效果
        if (leftPanel) {
          (leftPanel as HTMLElement).style.transition = '';
        }
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [isDragging, handleMouseMove, handleMouseUp]);

  return (
    <>
      <div
        className={classnames(
          'flex grow border-solid coz-stroke-primary coz-bg-max',
          fromProject
            ? 'h-[calc(100%-64px)] border-0 border-t'
            : 'h-[calc(100%-112px)] border rounded-[8px] m-4',
        )}
      >
        {/* 左侧文件树 */}
        <div
          className={classnames(
            'h-full shrink-0 overflow-auto p-[12px] transition-all duration-200 ease-in-out',
            'border-0 border-r border-solid coz-stroke-primary',
          )}
          style={{ width: `${leftWidth}px` }}
        >
          <div className="mb-4">
            <Text>{I18n.t('knowledge_level_012')}</Text>
          </div>
          {useDocumentListMode ? (
            <SegmentMenu
              isSearchable
              list={displayDocumentList.map(item => ({
                id: item.document_id ?? '',
                title: item.name ?? '',
                label: docOptions.find(opt => opt.value === item.document_id)
                  ?.label,
              }))}
              selectedID={curDocId}
              onClick={id => {
                if (id && id !== curDocId) {
                  handleSelectFromDocumentList(id);
                }
              }}
              treeDisabled
              treeVisible={false}
            />
          ) : treeLoading ? (
            <div className="flex justify-center items-center py-8">
              <Spin />
            </div>
          ) : (
            <TreeSelect
              style={{ width: '100%' }}
              value={selectedNode}
              dropdownStyle={{ maxHeight: 400, overflow: 'auto' }}
              treeData={treeData}
              placeholder={I18n.t('knowledge_select_document')}
              showSearch
              treeDefaultExpandAll
              onChange={handleDocumentSelect}
              loadData={onLoadData}
            />
          )}
        </div>

        {/* 拖拽分隔条 */}
        <div
          className={classnames(
            'w-[4px] h-full cursor-col-resize bg-transparent hover:bg-blue-500/20 transition-colors',
            isDragging && 'bg-blue-500/30',
          )}
          onMouseDown={handleMouseDown}
        />

        {/* 右侧内容区域 */}
        <div className="flex-1 h-full overflow-hidden">
          {!selectedFileData && !curDocFromList ? (
            <div className="flex items-center justify-center h-full">
              <EmptyState
                size="large"
                icon={
                  !selectedFileData && (
                    <IconSegmentEmpty style={{ width: 150, height: '100%' }} />
                  )
                }
                description={I18n.t('dataset_segment_select_file')}
              />
            </div>
          ) : selectedFileData.is_dir ? (
            <div className="flex items-center justify-center h-full">
              <EmptyState
                size="large"
                icon={
                  selectedFileData && (
                    <IconSegmentEmpty style={{ width: 150, height: '100%' }} />
                  )
                }
                description={I18n.t('dataset_segment_select_file')}
              />
            </div>
          ) : (
            <div className="flex flex-col h-full">
              {/* 使用 TextToolbar 组件 */}
              <div className="flex items-center justify-between border-b border-gray-200 flex-shrink-0">
                <TextToolbar
                  documentData={documentData}
                  filePreviewData={filePreviewData}
                  documentActions={documentActions}
                  customUIElements={customUIElements}
                  forbidSwitch={segments.length === 0}
                  canRenameDoc={false}
                />
              </div>

              {/* 文件分块内容 - 支持预览模式和分段模式切换 */}
              <div className="flex-1 overflow-hidden">
                {showOriginalFile || segments.length === 0 ? (
                  <div className="flex h-full gap-4 p-4">
                    {/* 左侧预览 */}
                    <div className="flex-1 rounded-md border border-gray-200 overflow-hidden shadow-sm">
                      <FilePreview
                        fileType={previewFileType}
                        fileUrl={
                          selectedFileData?.preview_tos_url || previewUrl || ''
                        }
                        visible={true}
                      />
                    </div>

                    {/* 右侧分段列表 */}
                    <div
                      className={classnames(
                        'flex-1 rounded-lg overflow-hidden',
                        {
                          hidden: segments.length === 0,
                        },
                      )}
                    >
                      {chunksLoading ? (
                        <div className="flex justify-center items-center h-full">
                          <Spin />
                        </div>
                      ) : (
                        renderSegmentList()
                      )}
                    </div>
                  </div>
                ) : (
                  <div className="h-full overflow-hidden rounded-lg">
                    {chunksLoading ? (
                      <div className="flex justify-center items-center h-full">
                        <Spin />
                      </div>
                    ) : (
                      renderSegmentList()
                    )}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </>
  );
};
