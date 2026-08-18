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

import { type ReactNode } from 'react';

import { BotMode } from '@coze-arch/bot-api/developer_api';

import { ChangeButton } from './change-button';

export interface ModeLabelProps {
  icon: ReactNode;
  isDisabled: boolean;
  isSelected: boolean;
  title: ReactNode;
  desc: ReactNode;
}

export interface ModeOption
  extends Omit<ModeLabelProps, 'isSelected' | 'isDisabled'> {
  value: BotMode;
  showText: boolean;
  getIsDisabled: (params: { currentMode: BotMode }) => boolean;
}

export interface ModeChangeViewProps {
  modeValue: BotMode;
  // onModeChange: (value: BotMode) => Promise<void>; // Removed: No longer interactive
  isReadOnly: boolean;
  optionList: ModeOption[];
  tooltip?: string;
}

export const ModeChangeView = (props: ModeChangeViewProps) => {
  const {
    modeValue = BotMode.SingleMode,
    tooltip,
    optionList,
  } = props;

  const modeInfo = optionList.find(option => option.value === modeValue);

  return (
    <ChangeButton disabled={false} tooltip={tooltip} modeInfo={modeInfo} />
  );
};