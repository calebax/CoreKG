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

import { type MutableRefObject, useEffect, useState, Fragment } from 'react';

import { useBotInfoStore } from '@coze-studio/bot-detail-store/bot-info'; // Keep if botId is needed directly
import {
  PictureUpload,
  type UploadValue,
} from '@coze-common/biz-components/picture-upload';
import { I18n } from '@coze-arch/i18n';
import { Modal, TabBar, TabPane, Button, Space } from '@coze-arch/coze-design';
import { useSpaceStore } from '@coze-arch/bot-studio-store';
import {
  FileBizType,
  IconType,
  type DraftBot,
} from '@coze-arch/bot-api/developer_api';
import { PlaygroundApi } from '@coze-arch/bot-api';
import {
  useAgentPersistence,
  useAgentFormManagement,
  AgentInfoForm,
  AICreateForm,
  type AICreateBotResponse,
} from '@coze-agent-ide/space-bot/hook';

import './index.module.less';

export interface CreateAgentEntityProps {
  onBefore?: () => void;
  onError?: () => void;
  /** Return Promise only if you need the onSuccess callback to block the pop-up window from closing automatically. */
  onSuccess?: (
    botId?: string,
    spaceId?: string,
    extra?: {
      botName?: string;
      botAvatar?: string;
      botDesc?: string;
    },
  ) => void | Promise<void>;
  botInfoRef?: MutableRefObject<DraftBot | undefined>;
  mode: 'update' | 'add';
  showSpace?: boolean;
  /**
   * Pass this parameter when you need to control externally which space to create the bot in
   * Only suitable for creating
   */
  spaceId?: string;
  /**
   * Navigation bar
   * Button in the upper right corner of the space workspace
   * */
  bizCreateFrom?: 'navi' | 'space';
  existingNames?: string[];
  excludeName?: string;
}

const getPictureUploadInitValue = (
  botInfo?: Partial<DraftBot>,
): UploadValue | undefined => {
  if (!botInfo?.icon_url) {
    return;
  }
  return [
    {
      url: botInfo.icon_url || '',
      uid: botInfo.icon_uri || '',
    },
  ];
};

export const useCreateOrUpdateAgent = ({
  botInfoRef,
  onBefore,
  onSuccess,
  onError,
  mode,
  showSpace = false, // Not displayed by default
  spaceId: outerSpaceId,
  bizCreateFrom,
  existingNames,
  excludeName,
}: CreateAgentEntityProps) => {
  const [visible, setVisible] = useState(false);
  const [activeTab, setActiveTab] = useState<'standard' | 'ai'>('standard');
  const [aiCreateData, setAiCreateData] = useState<AICreateBotResponse | null>(
    null,
  );

  const botId = useBotInfoStore(state => state.botId);
  const {
    space: { id: spaceId, hide_operation },
    spaces: { bot_space_list: list },
  } = useSpaceStore();

  const {
    formRef,
    isOkButtonDisable,
    checkErr,
    errMsg,
    confirmDisabled,
    hasNameValidationError,
    setHasNameValidationError,
    setCheckErr,
    setErrMsg,
    handleFormValuesChange,
    getValues,
    resetFormState,
  } = useAgentFormManagement({
    initialBotInfo: botInfoRef?.current,
    requireNameBlurValidation: Boolean(existingNames?.length),
  });

  const {
    loading: persistenceLoading,
    handleCreateBot,
    handleUpdateBot,
  } = useAgentPersistence({
    mode,
    botId,
    currentSpaceId: spaceId,
    outerSpaceId,
    getValues,
    onSuccess,
    onError,
    onBefore,
    setVisible,
    setCheckErr,
    setErrMsg,
    bizCreateFrom,
    showSpace,
    aiCreateData, // 传递 AI 创建的数据
  });

  useEffect(() => {
    if (visible) {
      useSpaceStore
        .getState()
        .fetchSpaces()
        .then(res => {
          if (!formRef.current?.formApi?.getValues()?.spaceId) {
            formRef.current?.formApi?.setValue(
              'spaceId',
              hide_operation
                ? res?.bot_space_list?.[0].id
                : spaceId ?? res?.bot_space_list?.[0].id,
            );
          }
        });
    }
    if (visible) {
      resetFormState();
      // 重置tab和AI创建数据
      setActiveTab('standard');
      setAiCreateData(null);
    }
  }, [visible]);

  /**
   * @Param _ open source version does not support this parameter
   */
  const startEdit = (_?: boolean) => {
    setVisible(true);
  };

  const formInitialValues = botInfoRef?.current || {};

  // 处理AI创建确认
  const handleAICreateConfirm = async (data: AICreateBotResponse) => {
    // 保存 AI 创建的数据到状态中
    setAiCreateData(data);

    // 将AI创建的数据填充到表单中
    if (formRef.current?.formApi) {
      formRef.current.formApi.setValue('name', data.name);
      formRef.current.formApi.setValue('target', data.description);
      if (data.icon_url && data.icon_uri) {
        formRef.current.formApi.setValue('bot_uri', [
          {
            url: data.icon_url,
            uid: data.icon_uri,
          },
        ]);
      }
    }

    // 直接执行创建逻辑，传递 AI 数据作为参数
    // 这样可以避免 React 状态更新的异步问题
    await handleCreateBot(data);
  };

  // 处理AI创建取消
  const handleAICreateCancel = () => {
    setVisible(false);
  };

  // 只在创建模式下显示Tabs
  const showTabs = mode === 'add';

  // 处理确认按钮点击
  const handleConfirm = async () => {
    if (mode === 'add') {
      await handleCreateBot();
    } else {
      await handleUpdateBot();
    }
  };

  // 处理取消按钮点击
  const handleCancel = () => {
    setVisible(false);
  };

  // 只在非AI创建模式下显示footer（AI创建模式有自己的按钮）
  const showFooter = !showTabs || activeTab === 'standard';

  return {
    startEdit,
    modal: (
      <Fragment>
        <Modal
          data-testid="bot.ide.bot_creator.create_bot_modal"
          visible={visible}
          maskClosable={false}
          onCancel={handleCancel}
          title={
            mode === 'add'
              ? I18n.t('bot_list_create')
              : I18n.t('bot_edit_title')
          }
          footer={
            showFooter ? (
              <Space className="agent-modal-footer-buttons">
                <Button type="tertiary" onClick={handleCancel}>
                  {I18n.t('Cancel')}
                </Button>
                <Button
                  type="primary"
                  loading={persistenceLoading}
                  disabled={
                    isOkButtonDisable ||
                    confirmDisabled ||
                    hasNameValidationError
                  }
                  onClick={handleConfirm}
                >
                  {I18n.t('Confirm')}
                </Button>
              </Space>
            ) : null
          }
          keepDOM={false}
          icon={null}
        >
          {showTabs ? (
            <TabBar
              activeKey={activeTab}
              onChange={(key: string) => setActiveTab(key as 'standard' | 'ai')}
              type="button"
              tabBarClassName="w-full"
            >
              <TabPane
                itemKey="standard"
                tab={I18n.t('bot_create_standard') || '标准创建'}
              >
                <AgentInfoForm
                  ref={formRef}
                  mode={mode}
                  showSpace={showSpace}
                  initialValues={formInitialValues}
                  spacesList={list || []}
                  currentSpaceId={outerSpaceId || spaceId}
                  hideOperation={hide_operation}
                  checkErr={checkErr}
                  errMsg={errMsg}
                  existingNames={existingNames}
                  excludeName={excludeName}
                  onNameValidationChange={setHasNameValidationError}
                  onValuesChange={handleFormValuesChange}
                  slot={
                    <PictureUpload
                      accept=".jpeg,.jpg,.png,.gif"
                      label={I18n.t('bot_edit_profile_pircture')}
                      field="bot_uri"
                      initValue={getPictureUploadInitValue(formInitialValues)}
                      rules={[{ required: true }]}
                      fileBizType={FileBizType.BIZ_BOT_ICON}
                      iconType={IconType.Bot}
                    />
                  }
                />
              </TabPane>
              <TabPane itemKey="ai" tab={I18n.t('bot_create_ai') || 'AI创建'}>
                <AICreateForm
                  onCancel={handleAICreateCancel}
                  onConfirm={handleAICreateConfirm}
                />
              </TabPane>
            </TabBar>
          ) : (
            <AgentInfoForm
              ref={formRef}
              mode={mode}
              showSpace={showSpace}
              initialValues={formInitialValues}
              spacesList={list || []}
              currentSpaceId={outerSpaceId || spaceId}
              hideOperation={hide_operation}
              checkErr={checkErr}
              errMsg={errMsg}
              existingNames={existingNames}
              excludeName={excludeName ?? formInitialValues?.name}
              onNameValidationChange={setHasNameValidationError}
              onValuesChange={handleFormValuesChange}
              slot={
                <PictureUpload
                  accept=".jpeg,.jpg,.png,.gif"
                  label={I18n.t('bot_edit_profile_pircture')}
                  field="bot_uri"
                  initValue={getPictureUploadInitValue(formInitialValues)}
                  rules={[{ required: true }]}
                  fileBizType={FileBizType.BIZ_BOT_ICON}
                  iconType={IconType.Bot}
                />
              }
            />
          )}
        </Modal>
      </Fragment>
    ),
  };
};
