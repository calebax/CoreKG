import { saveAs } from 'file-saver'
import { createFolder, getPreviewFileURL } from '@/api/knowledge'
import { UploadBaseInfo } from '@/stores/ForestUploadStore'
import { type Directory } from '../loadFile'

/**
 * 获取知识库文件的url
 * @param forest_id 知识库id
 * @param file_id 文件id
 * @param location 用于定位文件内容,pdf的页数、视频的时间等
 */
export function getFileURL(
  forest_id: number,
  file_id: number,
  location?: number[],
) {
  const url = `/docs/detail/${forest_id}/file/${file_id}`
  if (!location) return url
  const searchParams = new URLSearchParams()
  searchParams.append('location', encodeURIComponent(JSON.stringify(location)))
  return `${url}?${searchParams.toString()}`
}

/**
 * 下载知识库文件
 * @param file_id 文件id
 */
export async function downloadForestFile(file_id: number, fileName: string) {
  const { url } = await getPreviewFileURL({ file_id,is_download:true })
  const blob = await fetch(url).then((res) => res.blob())
  saveAs(blob, fileName)
}

/** 上传一个文件夹至当前目录内 */
export const uploadDir = async (
  forestPrefix: string,
  forest_id: number,
  parent_id: number,
  dir: Directory,
  upload: (info: UploadBaseInfo) => void,
) => {
  await Promise.allSettled(
    Array.from(dir.entries()).map(async ([name, file]) => {
      if (file instanceof File) {
        upload({ file, forest_id, parent_id, forestPrefix })
      } else {
        const { id: newId } = await createFolder({
          forest_id,
          parent_id,
          name,
        })
        await uploadDir(forestPrefix, forest_id, newId, file, upload)
      }
    }),
  )
}
