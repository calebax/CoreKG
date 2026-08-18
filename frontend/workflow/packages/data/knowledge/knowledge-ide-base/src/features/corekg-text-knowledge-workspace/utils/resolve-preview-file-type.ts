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

import { type DocumentInfo } from '@coze-arch/bot-api/knowledge';
import { OFFICE_PDF_PREVIEW_TYPES } from '@coze-data/knowledge-common-components';

import { type FileNode } from '../../../../../../../studio/workspace/entry-base/src/services/corekg-api';

const normalizeExt = (ext?: string) =>
  ext?.replace(/^\./, '').toLowerCase() ?? '';

export const getExtFromName = (name?: string) => {
  if (!name) {
    return '';
  }
  const index = name.lastIndexOf('.');
  if (index <= 0 || index === name.length - 1) {
    return '';
  }
  return name.slice(index + 1).toLowerCase();
};

export const matchForestFile = ({
  doc,
  fileData,
  forestFiles,
}: {
  doc?: DocumentInfo;
  fileData?: FileNode | null;
  forestFiles?: FileNode[];
}) =>
  forestFiles?.find(
    file =>
      (doc?.document_id && String(file.ID) === doc.document_id) ||
      (fileData?.ID && file.ID === fileData.ID) ||
      (doc?.name && file.name === doc.name) ||
      (fileData?.name && file.name === fileData.name),
  );

export const resolveSourceFileType = ({
  fileData,
  doc,
  forestFiles,
}: {
  fileData?: FileNode | null;
  doc?: DocumentInfo;
  forestFiles?: FileNode[];
}): string => {
  const matchedForestFile = matchForestFile({ doc, fileData, forestFiles });
  const fromExt = normalizeExt(fileData?.ext);
  if (fromExt) {
    return fromExt;
  }

  const fromDocType = normalizeExt(doc?.type);
  if (fromDocType) {
    return fromDocType;
  }

  const fromForestExt = normalizeExt(matchedForestFile?.ext);
  if (fromForestExt) {
    return fromForestExt;
  }

  return getExtFromName(doc?.name ?? fileData?.name);
};

export const resolvePreviewFileType = ({
  fileData,
  doc,
  forestFiles,
}: {
  fileData?: FileNode | null;
  doc?: DocumentInfo;
  forestFiles?: FileNode[];
}): string => {
  const matchedForestFile = matchForestFile({ doc, fileData, forestFiles });
  const fromForestPreviewExt = normalizeExt(matchedForestFile?.priview_ext);
  if (fromForestPreviewExt) {
    return fromForestPreviewExt;
  }

  const fromPreviewExt = normalizeExt(fileData?.priview_ext);
  if (fromPreviewExt) {
    return fromPreviewExt;
  }

  const sourceType = resolveSourceFileType({ fileData, doc, forestFiles });
  if (sourceType && OFFICE_PDF_PREVIEW_TYPES.has(sourceType)) {
    return 'pdf';
  }

  return sourceType || 'txt';
};
