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

import {
  DocumentSource,
  FormatType,
  type Dataset,
  type DocumentInfo,
} from '@coze-arch/bot-api/knowledge';

import { type FileNode } from '../../../../../../studio/workspace/entry-base/src/services/corekg-api';

const fileNodeToDocumentInfo = (file: FileNode): DocumentInfo =>
  ({
    document_id: String(file.ID),
    name: file.name,
    size: file.size,
    format_type: FormatType.Text,
    source_type: DocumentSource.Document,
  }) as DocumentInfo;

/**
 * 合并多来源文档列表：
 * - detail 接口的 dataset_details 是「多个知识库」，不是当前库的全部文件
 * - 当前库文件以 forest.ListFile / list 为准，取更完整的一份
 */
export const buildDisplayDocumentList = (
  documentList: DocumentInfo[] | undefined,
  forestFiles: FileNode[],
  dataSetDetail?: Dataset,
): DocumentInfo[] => {
  const forestDocs = forestFiles
    .filter(file => !file.is_dir)
    .map(fileNodeToDocumentInfo);

  const cozeDocs = documentList ?? [];
  const expectedCount =
    dataSetDetail?.doc_count ?? dataSetDetail?.file_list?.length ?? 0;

  const mergeByName = (
    primary: DocumentInfo[],
    secondary: DocumentInfo[],
  ): DocumentInfo[] => {
    const secondaryByName = new Map(
      secondary.filter(d => d.name).map(d => [d.name!, d]),
    );
    return primary.map(doc => {
      const matched = doc.name ? secondaryByName.get(doc.name) : undefined;
      return matched
        ? {
            ...matched,
            document_id: doc.document_id ?? matched.document_id,
            name: doc.name ?? matched.name,
          }
        : doc;
    });
  };

  if (forestDocs.length > 0 && cozeDocs.length > 0) {
    if (forestDocs.length >= cozeDocs.length) {
      return mergeByName(forestDocs, cozeDocs);
    }
    return mergeByName(cozeDocs, forestDocs);
  }

  if (forestDocs.length > 0) {
    return forestDocs;
  }

  if (
    cozeDocs.length > 0 &&
    (expectedCount === 0 || cozeDocs.length >= expectedCount)
  ) {
    return cozeDocs;
  }

  return cozeDocs;
};
