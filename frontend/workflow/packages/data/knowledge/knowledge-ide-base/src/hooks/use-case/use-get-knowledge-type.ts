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

import { useEffect } from 'react';

import { useShallow } from 'zustand/react/shallow';
import {
  useKnowledgeParams,
  useKnowledgeStore,
} from '@coze-data/knowledge-stores';

import { useDataSetDetailReq } from '@/service/dataset';
export const useGetKnowledgeType = () => {
  const { datasetID = '' } = useKnowledgeParams();
  // Knowledge Base Details
  const { data: dataSetDetail, run: fetchDataSetDetail, loading } =
    useDataSetDetailReq();
  const { setDataSetDetail, dataSetDetail: storeDataSetDetail } =
    useKnowledgeStore(
      useShallow(state => ({
        setDataSetDetail: state.setDataSetDetail,
        dataSetDetail: state.dataSetDetail,
      })),
    );
  const currentDataSetDetail = storeDataSetDetail?.dataset_id
    ? storeDataSetDetail
    : dataSetDetail;

  const hasMatchingDetail =
    !!datasetID &&
    !!currentDataSetDetail?.dataset_id &&
    String(currentDataSetDetail.dataset_id) === String(datasetID);

  const isDetailLoading = !!datasetID && (!hasMatchingDetail || loading);

  useEffect(() => {
    if (!datasetID || hasMatchingDetail) {
      return;
    }
    fetchDataSetDetail({ datasetID });
  }, [datasetID, hasMatchingDetail, fetchDataSetDetail]);

  useEffect(() => {
    if (!dataSetDetail) {
      return;
    }
    setDataSetDetail(dataSetDetail);
  }, [dataSetDetail, setDataSetDetail]);

  useEffect(
    () => () => {
      setDataSetDetail({});
    },
    [setDataSetDetail],
  );

  return {
    dataSetDetail: currentDataSetDetail,
    isDetailLoading,
    loading,
  };
};
