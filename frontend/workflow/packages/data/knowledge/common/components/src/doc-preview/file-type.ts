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

export const IMAGE_PREVIEW_TYPES = new Set([
  'jpg',
  'jpeg',
  'png',
  'webp',
  'gif',
  'bmp',
  'svg',
  'ico',
]);

export const AUDIO_PREVIEW_TYPES = new Set(['mp3', 'wav', 'ogg', 'm4a', 'aac']);

export const VIDEO_PREVIEW_TYPES = new Set(['mp4', 'webm', 'mov']);

// Office 文档预览链接通常由后端转换为 PDF
export const OFFICE_PDF_PREVIEW_TYPES = new Set([
  'pdf',
  'doc',
  'docx',
  'ppt',
  'pptx',
]);

export const isImagePreviewType = (fileType?: string) =>
  !!fileType && IMAGE_PREVIEW_TYPES.has(fileType.toLowerCase());

export const isAudioPreviewType = (fileType?: string) =>
  !!fileType && AUDIO_PREVIEW_TYPES.has(fileType.toLowerCase());

export const isVideoPreviewType = (fileType?: string) =>
  !!fileType && VIDEO_PREVIEW_TYPES.has(fileType.toLowerCase());

export const isOfficePdfPreviewType = (fileType?: string) =>
  !!fileType && OFFICE_PDF_PREVIEW_TYPES.has(fileType.toLowerCase());
