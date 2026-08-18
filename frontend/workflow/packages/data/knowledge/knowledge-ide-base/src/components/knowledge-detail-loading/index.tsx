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

import classNames from 'classnames';
import { Spin } from '@coze-arch/coze-design';

export const KnowledgeDetailLoadingContent = () => (
  <div
    className={classNames(
      'flex grow w-full items-center justify-center',
      'min-h-[480px] coz-bg-max rounded-[8px]',
      'border border-solid coz-stroke-primary',
    )}
  >
    <Spin size="large" />
  </div>
);
