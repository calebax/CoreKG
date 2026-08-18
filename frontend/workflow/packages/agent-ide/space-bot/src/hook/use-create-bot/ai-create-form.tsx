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

import { useState, useEffect, useRef } from 'react';

import classNames from 'classnames';
import { I18n } from '@coze-arch/i18n';
import { Button, TextArea, Avatar, Typography } from '@coze-arch/coze-design';
import { axiosInstance } from '@coze-arch/bot-http';
import { useSpaceStore } from '@coze-arch/bot-studio-store';

import LoadingIcon from '../../assets/image/loadingIcon.svg';

import s from './index.module.less';

const { Text } = Typography;

// loading步骤配置
const LOADING_STEPS = [
  '正在分析用户需求...',
  '正在构思智能体人设...',
  '正在匹配工具插件...',
  '正在撰写智能体提示词...',
] as const;

// 每个步骤的持续时间（毫秒）
const STEP_DURATION = 2000;

// AI创建接口返回数据类型
export interface AICreateBotResponse {
  name: string;
  description: string;
  icon_url: string;
  icon_uri: string;
  prompt: string;
  suggested_questions: string[];
  plugin_apis?: any; // 可选字段，接口可能不返回
  workflow?: any[]; // 接口返回的 workflow 字段
  prologue: string;
}

export interface AICreateFormProps {
  onCancel: () => void;
  onConfirm: (data: AICreateBotResponse) => void;
}

export const AICreateForm = ({ onCancel, onConfirm }: AICreateFormProps) => {
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingStep, setLoadingStep] = useState(0);
  const [result, setResult] = useState<AICreateBotResponse | null>(null);
  const loadingTimerRef = useRef<NodeJS.Timeout | null>(null);
  const getSpaceId = useSpaceStore(state => state.getSpaceId);

  // 清理定时器
  useEffect(() => {
    return () => {
      if (loadingTimerRef.current) {
        clearTimeout(loadingTimerRef.current);
      }
    };
  }, []);

  // 处理生成
  const handleGenerate = async () => {
    if (!query.trim()) {
      return;
    }

    setLoading(true);
    setLoadingStep(0);
    setResult(null);

    // 清理之前的定时器
    if (loadingTimerRef.current) {
      clearTimeout(loadingTimerRef.current);
    }

    let apiCompleted = false;
    let apiResult: AICreateBotResponse | null = null;
    let apiError: any = null;

    // 获取 space_id
    const spaceId = getSpaceId();

    // 真实接口调用
    const apiPromise = axiosInstance
      .request({
        url: '/api/playground_api/bot_config/create',
        method: 'POST',
        data: {
          scene: 8,
          query: query.trim(),
          space_id: spaceId,
        },
      })
      .then((res: any) => {
        // axios 响应拦截器返回的是整个 response 对象
        // res 是 axios response 对象，res.data 是接口返回的数据
        // 根据拦截器逻辑，如果接口返回格式是 { code: 0, data: {...} }，则实际数据在 res.data.data 中
        // 如果接口返回格式是直接的数据对象（没有 code 和 data 包裹），则数据在 res.data 中
        // 优先尝试 res.data.data（标准格式），如果没有则使用 res.data（直接数据格式）
        let data = res?.data?.data || res?.data;

        // 如果 data 是 null 或 undefined，尝试直接使用 res
        if (!data && res) {
          data = res;
        }

        if (!data || typeof data !== 'object' || Array.isArray(data)) {
          throw new Error('数据解析失败：无法获取有效的响应数据');
        }

        apiCompleted = true;
        apiResult = data;
        return data;
      })
      .catch((error: any) => {
        // 如果错误是 ApiError 且 httpStatusCode 是 200，说明可能是数据格式问题
        // 接口返回的是直接数据对象（没有 code 字段），被拦截器当作错误处理了
        // 尝试从 error.response.data 或 error.raw 中提取数据
        if (
          error?.httpStatusCode === '200' ||
          error?.response?.status === 200
        ) {
          // ApiError 的 raw 字段存储的是 response.data
          // 尝试从 error.response.data 或 error.raw 中获取数据
          const errorData = error?.response?.data || error?.raw;

          // 如果 errorData 是直接的数据对象（有 name 字段），说明这就是我们要的数据
          if (
            errorData &&
            typeof errorData === 'object' &&
            !Array.isArray(errorData) &&
            errorData.name
          ) {
            apiCompleted = true;
            apiResult = errorData;
            // 返回数据，让 promise 变成 resolved，这样后续代码可以正常处理
            return errorData;
          }
        }

        apiCompleted = true;
        apiError = error;
        throw error;
      });

    // 开始定时切换loading步骤
    let currentStep = 0;
    const startLoadingAnimation = () => {
      if (currentStep < LOADING_STEPS.length) {
        setLoadingStep(currentStep);
        loadingTimerRef.current = setTimeout(() => {
          currentStep++;
          if (currentStep < LOADING_STEPS.length) {
            startLoadingAnimation();
          } else {
            // 所有步骤都展示完了
            // 检查接口是否完成
            if (apiCompleted) {
              if (apiError) {
                setLoading(false);
                // 这里可以添加错误提示
              } else if (apiResult) {
                setResult(apiResult);
                setLoading(false);
              } else {
                setLoading(false);
              }
            } else {
              // 接口还没完成，等待接口完成
              apiPromise
                .then((data: AICreateBotResponse) => {
                  setResult(data);
                  setLoading(false);
                })
                .catch(() => {
                  setLoading(false);
                });
            }
          }
        }, STEP_DURATION);
      }
    };

    // 开始loading动画
    startLoadingAnimation();
  };

  // 处理重新生成
  const handleRegenerate = () => {
    handleGenerate();
  };

  // 处理确认
  const handleConfirm = () => {
    if (result) {
      onConfirm(result);
    }
  };

  return (
    <div className={s['ai-create-form']}>
      {/* Loading状态：生成过程中，整个内容区域只显示loading */}
      {loading ? (
        <>
          <div className={s['ai-loading']}>
            <img
              src={LoadingIcon}
              alt="loading"
              className={s['loading-icon']}
            />
            <div className={s['loading-content']}>
              <div className={s['loading-circle']} />
              <Text className={s['loading-text']}>
                {LOADING_STEPS[loadingStep]}
              </Text>
            </div>
          </div>
          {/* Loading时底部显示取消和生成按钮（生成按钮显示loading状态） */}
          <div className={s['ai-form-footer']}>
            <Button type="tertiary" onClick={onCancel}>
              {I18n.t('Cancel')}
            </Button>
            <Button
              type="primary"
              onClick={handleGenerate}
              disabled={!query.trim() || loading}
              loading={loading}
            >
              {I18n.t('bot_create_ai_generate') || '生成'}
            </Button>
          </div>
        </>
      ) : (
        <>
          {/* 输入框区域：只在非loading状态显示 */}
          <div className={s['ai-input-section']}>
            <TextArea
              value={query}
              onChange={value => setQuery(value)}
              placeholder="请描述您希望创建一个什么样的智能体"
              disabled={loading}
              autosize={{ minRows: 3 }}
              className={s['ai-textarea']}
            />
            {result && (
              <div className={s['ai-input-actions']}>
                <Button
                  type="tertiary"
                  theme="solid"
                  onClick={handleRegenerate}
                  className={classNames(
                    s['regenerate-btn'],
                    'regenerate-btn-global',
                  )}
                >
                  {I18n.t('bot_create_ai_regenerate') || '重新生成'}
                </Button>
              </div>
            )}
          </div>

          {/* 结果展示区域 */}
          {result && (
            <div className={s['ai-result']}>
              <div className={s['ai-result-header']}>
                <Avatar
                  src={result.icon_url}
                  size="large"
                  className={s['ai-result-avatar']}
                >
                  {result.name}
                </Avatar>
                <div className={s['ai-result-info']}>
                  <Text className={s['ai-result-name']}>{result.name}</Text>
                  <Text
                    className={s['ai-result-description']}
                    ellipsis={{
                      rows: 2,
                      showTooltip: {
                        type: 'tooltip',
                        opts: {
                          style: {
                            maxWidth: 400,
                            backgroundColor: 'black',
                            color: 'white',
                            wordBreak: 'break-word',
                          },
                        },
                      },
                    }}
                  >
                    {result.description}
                  </Text>
                </div>
              </div>
            </div>
          )}

          {/* 初始状态：底部显示取消和生成按钮 */}
          {!result && (
            <div className={s['ai-form-footer']}>
              <Button type="tertiary" onClick={onCancel}>
                {I18n.t('Cancel')}
              </Button>
              <Button
                type="primary"
                onClick={handleGenerate}
                disabled={!query.trim()}
              >
                {I18n.t('bot_create_ai_generate') || '生成'}
              </Button>
            </div>
          )}

          {/* 有结果后：底部显示取消和确认按钮 */}
          {result && (
            <div className={s['ai-form-footer']}>
              <Button type="tertiary" onClick={onCancel}>
                {I18n.t('Cancel')}
              </Button>
              <Button type="primary" onClick={handleConfirm}>
                {I18n.t('Confirm')}
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
};
