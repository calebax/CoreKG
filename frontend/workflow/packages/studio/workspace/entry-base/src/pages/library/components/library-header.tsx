/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from 'react';

import { I18n } from '@coze-arch/i18n';
import { IconCozArrowLeft, IconCozPlus } from '@coze-arch/coze-design/icons';
import { Button, Menu } from '@coze-arch/coze-design';
import { ResType } from '@coze-arch/bot-api/plugin_develop';
import { useNavigate } from 'react-router-dom';
import classNames from 'classnames';

import { type LibraryEntityConfig } from '../types';

export const LibraryHeader: React.FC<{
  entityConfigs: LibraryEntityConfig[];
}> = ({ entityConfigs }) => {
  const navigate = useNavigate();

  const ALLOWED_CREATE_TYPES = [
    ResType.Plugin,
    ResType.Workflow,
    ResType.Prompt,
  ];

  const handleBack = () => {
    if (window.opener && !window.opener.closed) {
      window.opener.focus();
      window.close();
      return;
    }
    navigate(-1);
  };

  return (
    <div className="flex items-center justify-between mb-[16px]">
      <div className="flex items-center gap-2">
        <div
          className={classNames(
            'flex items-center justify-center cursor-pointer hover:bg-gray-100 rounded',
            'transition-colors duration-200',
            'p-1'
          )}
          onClick={handleBack}
          data-testid="workspace.library.header.back"
        >
          <IconCozArrowLeft fontSize={20}/>
        </div>
        <div className="font-[500] text-[20px]">
          {I18n.t('navigation_workspace_library')}
        </div>
      </div>
      <Menu
        position="bottomRight"
        className="w-120px mt-4px mb-4px"
        render={
          <Menu.SubMenu mode="menu">
            {entityConfigs
              .filter(config =>
                config.target.some(t => ALLOWED_CREATE_TYPES.includes(t)),
              )
              .map(config => config.renderCreateMenu?.() ?? null)}
          </Menu.SubMenu>
        }
      >
        <Button
          theme="solid"
          type="primary"
          icon={<IconCozPlus />}
          data-testid="workspace.library.header.create"
        >
          {I18n.t('library_resource')}
        </Button>
      </Menu>
    </div>
  );
};