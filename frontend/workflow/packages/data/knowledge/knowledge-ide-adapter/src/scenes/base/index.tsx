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

import { useGetKnowledgeType } from '@coze-data/knowledge-ide-base/hooks/use-case';
import { FormatType } from '@coze-arch/bot-api/knowledge';
import { KnowledgeIDEBaseLayout } from '@coze-data/knowledge-ide-base/layout/base';
import { KnowledgeDetailLoadingContent } from '@coze-data/knowledge-ide-base/components/knowledge-detail-loading';
import { BaseKnowledgeIDENavBar } from '@coze-data/knowledge-ide-base/features/nav-bar/base';
import { CoreKGTextKnowledgeWorkspace } from '@coze-data/knowledge-ide-base/features/corekg-text-knowledge-workspace';
import { KnowledgeIDERegistryContext } from '@coze-data/knowledge-ide-base/context/knowledge-ide-registry-context';

import { importKnowledgeSourceMenuContributes as multimodalImportKnowledgeSourceMenuContributes } from './img-ide/import-knowledge-source-menu-contributes';

import { type BaseKnowledgeIDEProps } from './types';
import { BaseKnowledgeTextIDE } from './text-ide';
import { BaseKnowledgeTableIDE } from './table-ide';
import { BaseKnowledgeImgIDE } from './img-ide';

export type { BaseKnowledgeIDEProps };

const renderLoadingLayout = (props: BaseKnowledgeIDEProps) => (
  <KnowledgeIDEBaseLayout
    {...props.layoutProps}
    renderNavBar={
      props.layoutProps?.renderNavBar ??
      (({ statusInfo }) => (
        <BaseKnowledgeIDENavBar
          progressMap={statusInfo.progressMap}
          {...props.navBarProps}
        />
      ))
    }
    renderContent={() => <KnowledgeDetailLoadingContent />}
  />
);

export const BaseKnowledgeIDE = (props: BaseKnowledgeIDEProps) => {
  const { dataSetDetail, isDetailLoading } = useGetKnowledgeType();
  const format_type = dataSetDetail?.format_type;

  if (isDetailLoading) {
    return renderLoadingLayout(props);
  }

  // CoreKG multimodal (format_type = 11) uses CoreKGTextKnowledgeWorkspace inside base layout
  if (format_type === 11) {
    return (
      <KnowledgeIDERegistryContext.Provider
        value={{
          importKnowledgeMenuSourceFeatureRegistry:
            multimodalImportKnowledgeSourceMenuContributes,
        }}
      >
        <KnowledgeIDEBaseLayout
          renderNavBar={({ statusInfo }) => (
            <BaseKnowledgeIDENavBar
              progressMap={statusInfo.progressMap}
              {...props.navBarProps}
            />
          )}
          renderContent={({ dataActions, statusInfo }) => (
            <CoreKGTextKnowledgeWorkspace
              progressMap={statusInfo.progressMap}
              reload={dataActions.refreshData}
              onChangeDocList={dataActions.updateDocumentList}
            />
          )}
          {...props.layoutProps}
        />
      </KnowledgeIDERegistryContext.Provider>
    );
  }
  if (format_type === FormatType.Text) {
    return <BaseKnowledgeTextIDE {...props} />;
  }

  if (format_type === FormatType.Table || format_type === 12) {
    return <BaseKnowledgeTableIDE {...props} />;
  }
  if (format_type === FormatType.Image) {
    return <BaseKnowledgeImgIDE {...props} />;
  }
  return null;
};
