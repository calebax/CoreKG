import axios from 'axios'
import {
  combineLatest,
  concatMap,
  filter,
  from,
  map,
  mergeAll,
  mergeMap,
  Observable,
  repeat,
  ReplaySubject,
  retry,
  scan,
  share,
  shareReplay,
  startWith,
  Subject,
  take,
  takeLast,
  takeUntil,
  tap,
  timer,
  withLatestFrom,
} from 'rxjs'
import SparkMD5 from 'spark-md5'
import {
  getPresignedUrl,
  PresignedUploadFileCallBack,
  abortUpload,
  getNewUploadUrl,
} from '@/api/knowledge'

/**
 * 分片上传文件 自动断点续传
 * @returns 取消函数
 */
export const uploadFileToS3 = (data: {
  file: File
  forest_id: number
  parent_id: number
  split_config: any
  onFinish: (file_id: number) => void
  onError: (e: unknown) => void
  onUploadProgress?: (e: { loaded: number; total: number }) => void
}) => {
  const {
    file,
    forest_id,
    parent_id,
    split_config,
    onFinish,
    onError,
    onUploadProgress,
  } = data

  const abort$ = new ReplaySubject<void>()

  const hash$ = from(getFileMD5(file))
  // 获取预签名信息
  const presignedInfo$ = hash$.pipe(
    concatMap((hash) =>
      getPresignedUrl({
        forest_id,
        parent_id,
        files: [
          {
            hash,
            filename: file.name,
            size: file.size,
            split_config,
            content_type: file.type,
          },
        ],
      }),
    ),
    map((res) => {
      const info = res.files[0] as {
        upload_id: string
        hash: string
        exists: boolean
        multipart: {
          enabled: boolean
          chunk_size: number
          part_count: number
        }
        upload_urls: Record<number, string>
      }
      return info
    }),
    map((value) => {
      let chunkSize = file.size
      let totalPart = 1
      let maxUrl = 1
      if (value.multipart?.enabled) {
        ;({ chunk_size: chunkSize, part_count: totalPart } = value.multipart)
        maxUrl = Object.keys(value.upload_urls).length
      }
      return {
        ...value,
        chunkSize,
        totalPart,
        maxUrl,
      }
    }),
    retryWhenOffline(),
    takeUntil(abort$),
    take(1),
    shareReplay(1),
  )

  // 取消上传
  combineLatest([abort$, presignedInfo$, hash$])
    .pipe(
      mergeMap((v) => {
        const [, presignedInfo, hash] = v
        const { upload_id } = presignedInfo
        return abortUpload({ upload_id, hash })
      }),
      retryWhenOffline(),
    )
    .subscribe()

  // 重复上传
  combineLatest([presignedInfo$, hash$])
    .pipe(
      filter((v) => v[0].exists),
      concatMap((value) => {
        const [{ upload_id }, hash] = value
        return PresignedUploadFileCallBack({
          forest_id,
          upload_id,
          hash,
          filename: file.name,
        })
      }),
      map((res) => {
        return res.file_id as number
      }),
      retryWhenOffline(),
      takeUntil(abort$),
    )
    .subscribe({
      next: (file_id) => {
        onUploadProgress?.({ loaded: file.size, total: file.size })
        onFinish(file_id)
      },
      error: onError,
    })

  // 新上传
  // 始终按顺序上传片 同时存在的url不超过maxUrl 前一组上传结束后获取下一组的url

  const uploadInfo$ = presignedInfo$.pipe(filter((v) => !v.exists))

  type FilePartInfo = { partNumber: number; loaded?: number; etag?: string }
  /** 接受上传事件 */
  const uploadTrigger$ = new Subject<FilePartInfo>()
  /** 统计当前上传进度 获得一个partNumber为key的map */
  const uploadProgress$ = uploadTrigger$.pipe(
    scan((result, current) => {
      if (current.etag) {
        current.etag = current.etag.replace(/"/g, '')
      }
      const { partNumber } = current
      if (!result.has(partNumber)) {
        result.set(partNumber, { partNumber })
      }
      const target = result.get(partNumber)!
      result.set(partNumber, {
        ...target,
        ...current,
      })
      return result
    }, new Map<number, FilePartInfo>()),
    startWith(new Map<number, FilePartInfo>()),
    shareReplay(1),
  )
  /** 是否需要获取新的预签名url */
  let shouldGetNewUrls = false
  /** 片索引和对应的url */
  const latestUrls$ = combineLatest([uploadInfo$, hash$]).pipe(
    withLatestFrom(uploadProgress$),
    concatMap(async ([[info, hash], progress]) => {
      const {
        upload_id,
        upload_urls: defaultUrls,
        multipart,
        totalPart,
        maxUrl,
      } = info
      if (!shouldGetNewUrls) return defaultUrls
      // progress由后续逻辑确保当前所有片上传完毕 并且是从1开始的连续数列
      const completed_parts = [...progress.values()].map(
        (item) => item.partNumber,
      )
      // 1-completedPartNumber的片都上传了 下一个片从completedPartNumber开始
      const completedPartNumber = completed_parts.length
      const expired_parts = new Array<void>(
        Math.min(maxUrl, totalPart - completedPartNumber),
      )
        .fill()
        .map((_, i) => {
          return completedPartNumber + i + 1
        })

      const res = await getNewUploadUrl({
        hash,
        upload_id,
        completed_parts: multipart.enabled ? completed_parts : undefined,
        expired_parts: multipart.enabled ? expired_parts : undefined,
      })
      // key从completedPartNumber开始
      const renewed_urls = res.renewed_urls as Record<number, string>
      return renewed_urls
    }),
    map((urls) => {
      const urlList = Object.entries(urls).map(([k, url]) => {
        const partNumber = parseInt(k)
        return {
          partNumber,
          url,
        }
      })
      return urlList
    }),
    retryWhenOffline(),
  )

  //并发传输每个片
  combineLatest([latestUrls$, uploadInfo$])
    .pipe(
      mergeMap(([parts, info]) => {
        return parts.map((part) => [part, info] as const)
      }),
      // 用流表示片的进度并扁平化
      mergeMap(([part, info]) => {
        const { partNumber, url } = part
        const { chunkSize } = info
        const start = (partNumber - 1) * chunkSize
        const end = Math.min(file.size, partNumber * chunkSize)
        const putOnePart$ = new Observable<FilePartInfo>((subscriber) => {
          const controller = new AbortController()
          axios
            .put(url, file.slice(start, end), {
              onUploadProgress: (e) => {
                subscriber.next({
                  partNumber,
                  loaded: e.loaded,
                })
              },
              signal: controller.signal,
            })
            .then((res) => {
              subscriber.next({
                partNumber,
                etag: res.headers.etag,
              })
              subscriber.complete()
            })
            .catch((e) => subscriber.error(e))
          subscriber.add(() => {
            subscriber.next({
              partNumber,
              loaded: 0,
            })
            controller.abort()
          })
        })
        return putOnePart$
      }),
      // 重新订阅管道
      repeat({
        delay: () =>
          combineLatest([uploadProgress$, uploadInfo$]).pipe(
            map(([progress, info]) => {
              // 上游逻辑确保progress是已经上传的片
              const shouldRepeat = progress.size !== info.totalPart
              return shouldRepeat
            }),
            take(1),
            filter(Boolean),
            tap(() => {
              shouldGetNewUrls = true
            }),
          ),
      }),
      retryWhenOffline(),
      retry({
        delay: (e) => {
          if (e?.response?.status === 403) {
            // 鉴权失败的情况
            shouldGetNewUrls = true
            return timer(0)
          }
          throw e
        },
      }),
      takeUntil(abort$),
    )
    // 统一收集上传进度
    .subscribe(uploadTrigger$)

  // 向外上传进度 结束后使用回调
  uploadProgress$.subscribe((progress) => {
    let totalLoaded = 0
    ;[...progress.values()].forEach((v) => (totalLoaded += v.loaded ?? 0))
    onUploadProgress?.({ loaded: totalLoaded, total: file.size })
  })
  const uploadComplete$ = combineLatest([uploadInfo$, uploadProgress$]).pipe(
    filter(([info, progress]) => {
      return progress.size === info.totalPart
    }),
    map((val) => val[1]),
    takeLast(1),
    shareReplay(1),
  )
  combineLatest([uploadComplete$, uploadInfo$, hash$])
    .pipe(
      concatMap(async ([result, info, hash]) => {
        const { upload_id, multipart } = info
        const { file_id } = await PresignedUploadFileCallBack({
          forest_id,
          hash,
          upload_id,
          filename: file.name,
          parts: multipart.enabled
            ? [...result.values()]
                .map((item) => {
                  return {
                    part_number: item.partNumber,
                    etag: item.etag!,
                  }
                })
                .sort((v1, v2) => v1.part_number - v2.part_number)
            : undefined,
        })
        onFinish(file_id)
      }),
      retryWhenOffline(),
      takeUntil(abort$),
    )
    .subscribe({
      error: onError,
    })

  return () => abort$.next()
}

const getFileMD5 = async (file: File) => {
  const spark = new SparkMD5.ArrayBuffer()
  const reader = new FileReader()
  const chunkSize = 2 * 1024 * 1024
  const totalChunks = Math.ceil(file.size / chunkSize)
  for (let i = 0; i < totalChunks; i++) {
    const start = i * chunkSize
    const end = Math.min(start + chunkSize, file.size)
    const arrayBuffer = await new Promise<ArrayBuffer>((resolve, reject) => {
      reader.onload = (e) => {
        resolve(e.target?.result as ArrayBuffer)
      }
      reader.onerror = reject
      reader.readAsArrayBuffer(file.slice(start, end))
    })
    spark.append(arrayBuffer)
  }
  return `md5:${spark.end()}`
}

// 断网重试
const online$ = new Observable<void>((subscriber) => {
  const handleOnline = () => {
    subscriber.next()
  }
  window.addEventListener('online', handleOnline)
  subscriber.add(() => window.removeEventListener('online', handleOnline))
}).pipe(share())

/** 断网重试运算符 */
function retryWhenOffline<T>() {
  return retry<T>({
    delay: (e) => {
      if (!navigator.onLine) return online$
      throw e
    },
  })
}
