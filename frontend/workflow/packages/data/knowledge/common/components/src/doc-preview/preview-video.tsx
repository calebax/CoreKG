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

interface IPreviewVideoProps {
  fileUrl: string;
}

export const PreviewVideo = (props: IPreviewVideoProps) => {
  const { fileUrl } = props;

  if (!fileUrl) {
    return null;
  }

  return (
    <div className="flex items-center justify-center w-full h-full flex-1 p-4">
      <video
        controls
        src={fileUrl}
        className="max-w-full max-h-full"
      />
    </div>
  );
};
