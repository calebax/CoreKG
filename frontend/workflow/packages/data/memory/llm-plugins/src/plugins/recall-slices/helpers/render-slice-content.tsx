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

import React from 'react';

/**
 * 知识库切片中图片可能存在两种写法：
 * 1. 标准 Markdown：![alt](https://url)
 * 2. 非标准写法：![alt] https://url （方括号后跟空格 + 裸 URL，无圆括号）
 *
 * 这里统一匹配上述两种形式，把切片拆分为「文本片段」与「图片片段」，
 * 文本按原样展示，图片用 <img> 渲染，避免召回区只看到 alt 文案而看不到图片。
 */
const IMAGE_PATTERN =
  /!\[([^\]]*)\]\s*(?:\((https?:\/\/[^)\s]+)\)|(https?:\/\/[^\s)]+))/g;

export const renderSliceContent = (slice: string): React.ReactNode[] => {
  const nodes: React.ReactNode[] = [];
  let lastIndex = 0;
  let match: RegExpExecArray | null;
  let key = 0;

  IMAGE_PATTERN.lastIndex = 0;
  // eslint-disable-next-line no-cond-assign
  while ((match = IMAGE_PATTERN.exec(slice)) !== null) {
    const [matched, alt, urlInParen, bareUrl] = match;
    const url = urlInParen ?? bareUrl;

    if (match.index > lastIndex) {
      nodes.push(
        <span key={`text-${key++}`}>{slice.slice(lastIndex, match.index)}</span>,
      );
    }

    if (url) {
      nodes.push(
        <img
          key={`img-${key++}`}
          src={url}
          alt={alt}
          loading="lazy"
          referrerPolicy="no-referrer"
        />,
      );
    }

    lastIndex = match.index + matched.length;
  }

  if (lastIndex < slice.length) {
    nodes.push(<span key={`text-${key++}`}>{slice.slice(lastIndex)}</span>);
  }

  return nodes;
};
