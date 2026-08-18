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

import { type FC, useEffect, useState } from 'react';

import { useParams, useNavigate } from 'react-router-dom';

import { IconCozArrowLeft } from '@coze-arch/coze-design/icons';
import { Avatar, IconButton } from '@coze-arch/coze-design';
import { Spin } from '@coze-arch/coze-design';
import { PlaygroundApi } from '@coze-arch/bot-api';
import { useUserInfo } from '@coze-foundation/account-base';
import { useBotInfoStore } from '@coze-studio/bot-detail-store/bot-info';
import { useBotSkillStore } from '@coze-studio/bot-detail-store/bot-skill';
import { usePageRuntimeStore } from '@coze-studio/bot-detail-store/page-runtime';
import { BotDebugChatAreaProviderAdapter } from '@coze-agent-ide/chat-area-provider-adapter';
import { BotDebugChatArea } from '@coze-agent-ide/chat-debug-area';

import './bot-chat.less';

export const BotChat: FC = () => {
  const navigate = useNavigate();
  const { space_id, bot_id } = useParams<{
    space_id: string;
    bot_id: string;
  }>();

  const accountUserInfo = useUserInfo();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 从 store 获取 bot 信息
  const botName = useBotInfoStore(state => state.name);
  const botIconUrl = useBotInfoStore(state => state.icon_url);

  // 初始化 store
  useEffect(() => {
    const initStores = async () => {
      if (!bot_id || !space_id) {
        setError('缺少必要参数');
        setLoading(false);
        return;
      }

      try {
        setLoading(true);

        // 获取 bot 详情
        const response = await PlaygroundApi.GetDraftBotInfoAgw({
          bot_id,
        });

        if (!response?.data) {
          throw new Error('获取智能体信息失败');
        }

        const botData = response.data;
        const botInfo = botData.bot_info;

        // 初始化 BotInfoStore
        useBotInfoStore.getState().initStore({
          ...botData,
          bot_info: botInfo,
        });

        // 初始化 BotSkillStore
        useBotSkillStore.getState().initStore({
          ...botData,
          bot_info: botInfo,
        });

        // 初始化 PageRuntimeStore
        usePageRuntimeStore.getState().initStore({
          editable: botData.editable,
          has_unpublished_change: botData.has_unpublished_change,
        });

        setLoading(false);
      } catch (err) {
        console.error('初始化失败:', err);
        setError('加载智能体信息失败');
        setLoading(false);
      }
    };

    initStores();

    // 清理 store
    return () => {
      useBotInfoStore.getState().clear();
      useBotSkillStore.getState().clear();
      usePageRuntimeStore.getState().clear();
    };
  }, [bot_id, space_id]);

  if (loading) {
    return (
      <div className="w-full h-full flex items-center justify-center coz-bg-max">
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="w-full h-full flex flex-col items-center justify-center coz-bg-max">
        <div className="text-[16px] coz-fg-secondary mb-[16px]">{error}</div>
        <button
          type="button"
          className="px-[16px] py-[8px] rounded-[8px] coz-bg-hglt text-white cursor-pointer"
          onClick={() => navigate(`/space/${space_id}/develop`)}
        >
          返回列表
        </button>
      </div>
    );
  }

  return (
    <div className="w-full h-full flex flex-col coz-bg-max overflow-hidden">
      {/* 顶部标题栏 */}
      <div className="h-[56px] flex items-center px-[24px] gap-[12px] border-b border-solid coz-stroke-primary flex-shrink-0">
        {/* 返回按钮 */}
        <IconButton
          icon={<IconCozArrowLeft />}
          onClick={() => {
            navigate(`/space/${space_id}/develop`);
          }}
        />

        {/* 面包屑导航 */}
        <div className="flex items-center gap-[8px] text-[14px]">
          <span
            className="coz-fg-secondary cursor-pointer hover:coz-fg-primary"
            onClick={() => navigate(`/space/${space_id}/develop`)}
          >
            智能体
          </span>
          <span className="coz-fg-secondary">/</span>
          <span className="coz-fg-primary">对话</span>
        </div>

        {/* 机器人信息 */}
        <div className="ml-[16px] flex items-center gap-[8px]">
          <Avatar
            className="w-[28px] h-[28px] rounded-[6px]"
            shape="square"
            src={botIconUrl}
            style={
              !botIconUrl
                ? {
                    backgroundColor: '#0C99FF',
                    color: '#fff',
                    fontSize: '12px',
                  }
                : {}
            }
          >
            {!botIconUrl ? '🤖' : null}
          </Avatar>
          <span className="text-[14px] font-[500] coz-fg-primary">
            {botName || '智能助手'}
          </span>
        </div>
      </div>

      {/* 主体内容区域 - 直接使用预览调试组件 */}
      <div className="flex-1 overflow-hidden bot-chat-area">
        <BotDebugChatAreaProviderAdapter
          botId={bot_id || ''}
          userId={accountUserInfo?.user_id_str}
        >
          <BotDebugChatArea readOnly={false} />
        </BotDebugChatAreaProviderAdapter>
      </div>
    </div>
  );
};
