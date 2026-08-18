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

/* eslint-disable @coze-arch/max-line-per-function */
import { type FC, useMemo, useState, useEffect } from 'react';

import {
  CLOUD_PLUGIN_COZE,
  type PluginInfoProps,
} from '@coze-studio/plugin-shared';
import {
  PluginForm,
  usePluginFormState,
  convertPluginMetaParams,
  registerPluginMeta,
  updatePluginMeta,
} from '@coze-studio/plugin-form-adapter';
import { withSlardarIdButton } from '@coze-studio/bot-utils';
import { I18n } from '@coze-arch/i18n';
import { useSpaceStore } from '@coze-arch/bot-studio-store';
import {
  type CreationMethod,
  type PluginType,
} from '@coze-arch/bot-api/plugin_develop';
import { ERROR_CODE } from '@coze-agent-ide/bot-plugin-tools/pluginModal/types';
import { IconCozInfoCircleFill } from '@coze-arch/coze-design/icons';
import {
  Button,
  Divider,
  Modal,
  Space,
  Toast,
  Typography,
} from '@coze-arch/coze-design';

import s from '../index.module.less';
import { PluginDocs } from '../../plugin-docs';
import { ImportModal } from './import-modal';
import { CodeModal } from './code-modal';

const isDuplicateName = (
  value: string | undefined,
  existingNames: string[],
  excludeName = '',
) => {
  const normalized = (value ?? '').trim();
  const normalizedExclude = excludeName.trim();
  if (!normalized) {
    return false;
  }
  return existingNames.some(name => {
    const existing = (name ?? '').trim();
    return (
      Boolean(existing) &&
      existing !== normalizedExclude &&
      existing === normalized
    );
  });
};

const isPluginFormSubmittable = ({
  values,
  pluginTypeCreationMethod,
  existingNames,
  excludeName,
}: {
  values: Record<string, any>;
  pluginTypeCreationMethod?: string;
  existingNames?: string[];
  excludeName?: string;
}) => {
  const name = String(values.name ?? '').trim();
  const desc = String(values.desc ?? '').trim();
  const effectivePluginTypeCreationMethod = String(
    values.creation_method ?? pluginTypeCreationMethod ?? '',
  );

  if (!name || !desc || !effectivePluginTypeCreationMethod) {
    return false;
  }

  if (name.length > 30 || desc.length > 600) {
    return false;
  }

  const namePattern = IS_OVERSEA || IS_BOE ? /^[\w\s]+$/ : /^[\w\s\u4e00-\u9fa5]+$/u;
  if (!namePattern.test(name)) {
    return false;
  }

  if (existingNames?.length && isDuplicateName(name, existingNames, excludeName)) {
    return false;
  }

  if (effectivePluginTypeCreationMethod === CLOUD_PLUGIN_COZE) {
    const url = String(values.url ?? '').trim();
    const authType =
      (values.auth_type as Array<number> | undefined) && values.auth_type.length
        ? (values.auth_type as Array<number>)
        : [0];

    if (!url || !authType?.length) {
      return false;
    }

    if (authType[0] === 1 && authType.at(-1) === 0) {
      const key = String(values.key ?? '').trim();
      const serviceToken = String(values.service_token ?? '').trim();
      if (!key || !serviceToken || !/^[\x00-\x7F]+$/.test(key)) {
        return false;
      }
    }
  }

  return true;
};

export interface CreatePluginFormProps {
  visible: boolean;
  isCreate?: boolean;
  editInfo?: PluginInfoProps;
  disabled?: boolean;
  onCancel?: () => void;
  onSuccess?: (pluginID?: string) => Promise<void> | void;
  projectId?: string;
  existingNames?: string[];
}

export const CreateFormPluginModal: FC<CreatePluginFormProps> = props => {
  const {
    onCancel,
    editInfo,
    isCreate = true,
    visible,
    onSuccess,
    disabled = false,
    projectId,
    existingNames,
  } = props;

  const { id } = useSpaceStore(store => store.space);
  const modalTitle = useMemo(() => {
    if (isCreate) {
      return (
        <div className="w-full flex justify-between items-center pr-[8px]">
          <div>{I18n.t('create_plugin_modal_title1')}</div>
          {/* <Space>
            <CodeModal
              onCancel={onCancel}
              onSuccess={onSuccess}
              projectId={projectId}
            />
            <ImportModal
              onCancel={onCancel}
              onSuccess={onSuccess}
              projectId={projectId}
            />
            <Divider layout="vertical" className="h-5" />
          </Space> */}
        </div>
      );
    }
    if (disabled) {
      return I18n.t('plugin_detail_view_modal_title');
    }
    return I18n.t('plugin_detail_edit_modal_title');
  }, [isCreate, disabled]);
  const [loading, setLoading] = useState(false);
  const [isSubmitDisabled, setIsSubmitDisabled] = useState(true);
  const pluginState = usePluginFormState();

  const {
    formApi,
    extItems,
    headerList,
    isValidCheckResult,
    setIsValidCheckResult,
    pluginTypeCreationMethod,
    defaultRuntime,
  } = pluginState;

  const updateSubmitDisabled = (partialValues?: Record<string, any>) => {
    const values = {
      ...(formApi.current?.getValues() ?? {}),
      ...partialValues,
    };
    setIsSubmitDisabled(
      !isPluginFormSubmittable({
        values,
        pluginTypeCreationMethod:
          partialValues?.creation_method ?? pluginTypeCreationMethod,
        existingNames,
        excludeName: editInfo?.meta_info?.name,
      }),
    );
  };

  useEffect(() => {
    if (!isCreate) {
      return;
    }
    if (visible) {
      // Scroll bar after display to top
      const modalContent = document.querySelector(
        '.create-plugin-modal-content .semi-modal-body',
      );
      if (modalContent) {
        modalContent.scrollTop = 0;
      }
      setTimeout(() => {
        updateSubmitDisabled(
          isCreate ? { creation_method: CLOUD_PLUGIN_COZE } : undefined,
        );
      }, 0);
    } else {
      // Reset form after hiding
      formApi?.current?.reset();
      setIsSubmitDisabled(true);
    }
  }, [visible, existingNames, pluginTypeCreationMethod]);

  const confirmBtn = async () => {
    await formApi.current?.validate();
    const type = isCreate ? 'create' : 'edit';
    const val = formApi.current?.getValues();
    if (!val || !pluginTypeCreationMethod) {
      return;
    }

    const json: Record<string, string> = {};
    extItems?.forEach(item => {
      if (item.key in val) {
        json[item.key] = val[item.key];
      }
    });

    const [pluginType, creationMethod] = pluginTypeCreationMethod.split('-');

    const params = convertPluginMetaParams({
      val,
      spaceId: String(id),
      headerList,
      projectId,
      creationMethod: Number(creationMethod) as unknown as CreationMethod,
      defaultRuntime,
      pluginType: Number(pluginType) as unknown as PluginType,
      extItemsJSON: json,
    });
    const action = {
      create: () => registerPluginMeta({ params }),
      edit: () => updatePluginMeta({ params, editInfo }),
    };

    try {
      setLoading(true);
      const pluginID = await action[type]();
      Toast.success({
        content: isCreate
          ? I18n.t('Plugin_new_toast_success')
          : I18n.t('Plugin_update_toast_success'),
        showClose: false,
      });
      onCancel?.();
      onSuccess?.(pluginID);
    } catch (error) {
      // @ts-expect-error -- linter-disable-autofix
      const { code, msg } = error;
      if (Number(code) === ERROR_CODE.SAFE_CHECK) {
        setIsValidCheckResult(false);
      } else {
        Toast.error({
          content: withSlardarIdButton(msg),
        });
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Modal
        title={modalTitle}
        className="[&_.semi-modal-header]:items-center"
        visible={visible}
        keepDOM={isCreate}
        onCancel={() => onCancel?.()}
        modalContentClass="create-plugin-modal-content"
        footer={
          !disabled && (
            <div>
              {!isValidCheckResult && (
                <div className={s['error-msg-box']}>
                  <span className={s['error-msg']}>
                    {I18n.t('plugin_create_modal_safe_error')}
                  </span>
                </div>
              )}
              {/* <Typography.Paragraph
                type="secondary"
                fontSize="12px"
                className="text-start mb-[16px]"
              >
                <IconCozInfoCircleFill className="coz-fg-hglt text-[14px] align-sub" />
                <span className="mx-[4px]">
                  {I18n.t('plugin_create_draft_desc')}
                </span>
                <PluginDocs />
              </Typography.Paragraph> */}
              <div className="plugin-modal-footer-buttons">
                <Button
                  type="tertiary"
                  onClick={() => {
                    onCancel?.();
                  }}
                >
                  {I18n.t('create_plugin_modal_button_cancel')}
                </Button>
                <Button
                  type="primary"
                  loading={loading}
                  disabled={isSubmitDisabled}
                  onClick={() => {
                    confirmBtn();
                  }}
                >
                  {I18n.t('create_plugin_modal_button_confirm')}
                </Button>
              </div>
            </div>
          )
        }
      >
        <PluginForm
          pluginState={pluginState}
          visible={visible}
          isCreate={isCreate}
          disabled={disabled}
          editInfo={editInfo}
          existingNames={existingNames}
          onNameValidationChange={values =>
            updateSubmitDisabled(values as Record<string, any> | undefined)
          }
        />
      </Modal>
    </>
  );
};
