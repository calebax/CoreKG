import { produce } from 'immer'
import { create } from 'zustand'
import type { SegmentRule } from '@/pages/app/docs/detail/components/ActionButtons/UploadButton/SegmentRuleModal'
import { uploadFileToS3 } from './UploadFileToS3'

export type Status =
  | 'waiting'
  | 'uploading'
  | 'finished'
  | 'error'
  | 'cancelled'
  | 'illegal' // 不合法的文件 直接不上传
export type UploadBaseInfo = {
  file: File
  /** 知识库专属前缀 例如/docs/cad/:id里的cad */
  forestPrefix: string
  forest_id: number
  parent_id: number
  segmentRule?: SegmentRule
  onSuccess?: (response?: any) => void
  onError?: (error: unknown) => void
}
export type FileItem = UploadBaseInfo & {
  key: string
  status: Status
  /** 上传成功后的文件id */
  file_id?: number
  /** 上次记录loaded的时间戳 */
  uploadTs?: number
  /** 已上传字节数 */
  loaded?: number
  /** 已上传的百分比 */
  percent?: number
  /** 上传速度 字节每秒 */
  speed?: number
  /** 剩余上传时间 秒 */
  restTime?: number
  cancel?: () => void
  error?: unknown
  // 不合法文件的原因
  illeagalReason?: string
}
export type UploadOption = {
  /** 判断文件是否过大 */
  isOversize?: (file: File) => boolean
}
type ForestUploadStore = {
  files: FileItem[]
  getFilesByOptions: (options?: {
    status?: Status | Status[]
    forest_id?: number | number[]
  }) => FileItem[]
  /** 上传一个文件(可能将其插入等待队列) */
  upload: (info: UploadBaseInfo, option?: UploadOption) => void
  /** 立刻上传正在等待的第一个文件 */
  uploadWaiting: () => void
  /** 重新上传失败的文件 */
  retryUpload: (key: string) => void
}
/** 最多同时上传的文件个数 */
const maxUploading = 3
/**
 * 知识库上传文件全局存储\
 * 不需要跨页面同步 也不持久化
 */
export const useForestUploadStore = create<ForestUploadStore>((_set, get) => {
  const immerSet = (fn: (val: ForestUploadStore) => void) =>
    _set((prev) => produce(prev, fn))
  return {
    files: [],
    getFilesByOptions: (options = {}) => {
      const { status = [], forest_id = [] } = options
      const statusList = typeof status === 'string' ? [status] : status
      const idList = typeof forest_id === 'number' ? [forest_id] : forest_id
      return get().files.filter((item) => {
        if (status.length !== 0 && !statusList.includes(item.status)) {
          return false
        }
        if (idList.length !== 0 && !idList.includes(item.forest_id)) {
          return false
        }
        return true
      })
    },
    upload: (info, option = {}) =>
      immerSet((draft) => {
        const { files } = draft
        const key = `${performance.now()}-${JSON.stringify({ ...info, file: info.file.name })}`
        // 根据option判断文件是否上传
        if (option.isOversize?.(info.file)) {
          files.push({
            ...info,
            key,
            status: 'illegal',
            illeagalReason: '文件超出大小',
          })
          return
        }
        files.push({
          ...info,
          key,
          status: 'waiting',
          cancel: () =>
            immerSet((draft) => {
              const target = draft.files.find((item) => item.key === key)
              if (!target) return
              const wasWaiting = target.status === 'waiting'
              target.status = 'error'
              target.cancel = undefined
              if (wasWaiting) {
                Promise.resolve().then(get().uploadWaiting)
              }
            }),
        })
        Promise.resolve().then(get().uploadWaiting)
      }),
    uploadWaiting: () =>
      immerSet((draft) => {
        const { files } = draft
        if (
          files.filter((item) => item.status === 'uploading').length >=
          maxUploading
        )
          return
        const target = files.find((item) => item.status === 'waiting')
        if (!target) return
        const { key } = target
        const _cancel = PresignedUploadFile({
          ...target,
          onUploadProgress: (e) => {
            const { loaded, total } = e
            const ts = performance.now()
            immerSet((draft) => {
              const target = draft.files.find((item) => item.key === key)
              if (!target) return
              const lastTs = target.uploadTs!
              target.uploadTs = ts
              const lastLoaded = target.loaded!
              target.loaded = loaded
              target.percent = loaded / total
              if (lastTs) {
                target.speed = ((loaded - lastLoaded) / (ts - lastTs)) * 1000
                target.restTime = Math.max(total - loaded, 0) / target.speed
              }
            })
          },
          onFinish: (file_id, response) => {
            immerSet((draft) => {
              const target = draft.files.find((item) => item.key === key)
              if (!target) return
              target.status = 'finished'
              target.cancel = undefined
              target.file_id = file_id
              // 调用自定义成功回调，传递响应数据
              target.onSuccess?.(response)
              Promise.resolve().then(get().uploadWaiting)
            })
          },
          onError: (e) => {
            immerSet((draft) => {
              const target = draft.files.find((item) => item.key === key)
              if (!target) return
              target.status = 'error'
              target.error = e
              target.cancel = undefined
              // 调用自定义错误回调
              target.onError?.(e)
              Promise.resolve().then(get().uploadWaiting)
            })
          },
        })
        const cancel = () => {
          _cancel()
          immerSet((draft) => {
            const target = draft.files.find((item) => item.key === key)
            if (!target) return
            target.status = 'error'
            target.cancel = undefined
            Promise.resolve().then(get().uploadWaiting)
          })
        }
        target.status = 'uploading'
        target.cancel = cancel
        target.uploadTs = performance.now()
        target.loaded = 0
        target.percent = 0
      }),
    retryUpload: (key: string) =>
      immerSet((draft) => {
        const target = draft.files.find((item) => item.key === key)
        if (!target || target.status !== 'error') return
        target.status = 'waiting'
        target.error = undefined
        target.loaded = 0
        target.percent = 0
        target.uploadTs = undefined
        target.speed = undefined
        target.restTime = undefined

        target.cancel = () =>
          immerSet((draft) => {
            const target = draft.files.find((item) => item.key === key)
            if (!target) return
            const wasWaiting = target.status === 'waiting'
            target.status = 'error'
            target.cancel = undefined
            if (wasWaiting) {
              Promise.resolve().then(get().uploadWaiting)
            }
          })
        Promise.resolve().then(get().uploadWaiting)
      }),
  }
})

/**
 * 预签名上传\
 * 返回一个取消函数
 */
function PresignedUploadFile(
  uploadInfo: UploadBaseInfo & {
    onFinish: (file_id: number, response?: any) => void
    onError: (e: unknown) => void
    onUploadProgress?: (e: { loaded: number; total: number }) => void
  },
) {
  const {
    file,
    forest_id,
    parent_id,
    segmentRule,
    onUploadProgress,
    onFinish,
    onError,
  } = uploadInfo

  // 构建分段配置 - 无论是默认规则还是自定义规则，都需要传完整的参数
  const split_config = {
    split_mode: segmentRule?.type || 'auto',
    // 默认值或用户设置的值
    chunk_size: segmentRule?.type === 'rule' ? segmentRule.segmentLength : 256,
    split_mark:
      segmentRule?.type === 'rule' && segmentRule.segmentSeparator
        ? [segmentRule.segmentSeparator]
        : ['\n'],
    // split_overlap需要转换为小数：25% -> 0.25
    split_overlap:
      segmentRule?.type === 'rule'
        ? (segmentRule.segmentOverlap ?? 30) / 100
        : 0.25,
    preprocessing_rules: {
      remove_empty_line:
        segmentRule?.type === 'rule'
          ? segmentRule.textPreprocessing?.removeExtraSpaces || false
          : false,
      remove_url:
        segmentRule?.type === 'rule'
          ? segmentRule.textPreprocessing?.removeLineBreaks || false
          : false,
      remove_email:
        segmentRule?.type === 'rule'
          ? segmentRule.textPreprocessing?.removeSpecialChars ?? true
          : false,
    },
  }

  const abortUpload = uploadFileToS3({
    file,
    forest_id,
    parent_id,
    split_config,
    onUploadProgress,
    onFinish,
    onError,
  })
  return abortUpload
}
