import { useRequest } from 'ahooks'
import { listParseHistory } from '@/api/knowledge'
import { AnalyzeFile } from '../ForestAnalyzeFiles'

/** 解析队列 */
export const useAnalyzeFiles = () => {
  const { data, run } = useRequest(
    async () => {
      const res = await listParseHistory({
        filters: [
          {
            field: 'recent_days',
            value: ['5'],
          },
        ],
        orderBy: ['created_at desc'],
      })
      const data: any[] = res.Data ?? []
      const files: AnalyzeFile[] = data.map((item) => {
        const {
          name,
          size,
          forest_id,
          parent_id,
          ID: file_id,
          file_status,
          file_progress,
          forest_type,
          data_source_type,
        } = item
        const status: AnalyzeFile['status'] = (() => {
          switch (file_status) {
            case 'success':
              return 'finished'
            case 'running':
              return 'analyzing'
            case 'pending':
              return 'waiting'
            default:
              return 'error'
          }
        })()
        const forestPrefix = (() => {
          switch (forest_type) {
            case 'data': {
              switch (data_source_type) {
                case 'excel':
                  return 'excel'
                case 'db':
                  return 'db'
              }
              break
            }
            case 'file':
              return 'detail'
            default:
              return forest_type
          }
        })()
        const percent = Number(file_progress)

        return {
          name,
          size,
          percent,
          status,
          forest_id,
          parent_id,
          file_id,
          forestPrefix,
        }
      })
      return files
    },
    { pollingInterval: 15000 },
  )
  return {
    analyzeFiles: data,
    reload: run,
  }
}
