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

import { I18n } from '@coze-arch/i18n';

export const normalizeResourceName = (name?: string | null): string =>
  (name ?? '').trim();

export const isDuplicateName = (
  name: string,
  existingNames: string[],
  excludeName?: string,
): boolean => {
  const normalized = normalizeResourceName(name);
  if (!normalized) {
    return false;
  }
  const normalizedExclude = normalizeResourceName(excludeName);
  return existingNames.some(existingName => {
    const normalizedExisting = normalizeResourceName(existingName);
    return (
      Boolean(normalizedExisting) &&
      normalizedExisting !== normalizedExclude &&
      normalizedExisting === normalized
    );
  });
};

export const createDuplicateNameRule = (
  existingNames: string[],
  excludeName?: string,
) => ({
  validator: (_: unknown, value: string) => {
    if (isDuplicateName(value, existingNames, excludeName)) {
      return new Error(I18n.t('name_already_taken'));
    }
    return true;
  },
});

export const createWorkflowDuplicateNameValidator = (
  existingNames: string[],
  excludeName?: string,
) => ({
  validator: (_: unknown, value: string) => {
    if (isDuplicateName(value, existingNames, excludeName)) {
      return new Error(I18n.t('name_already_taken'));
    }
    return true;
  },
});
