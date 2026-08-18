import { useParams } from 'react-router-dom'
import { Breadcrumb, Spin } from 'antd'
import { useRequest } from 'ahooks'
import { getFileInfo } from '@/api/knowledge'
import FilePreview from '../FilePreview'
import ResizablePanel from '../ResizablePanel'
import RightPanel from '../RightPanel'
import { FileLocationProvider } from './FileLocationContext'

export default function FileDetailView() {
  const { id, fileId } = useParams()

  const { data: fileDetail } = useRequest(
    async () => {
      const res = await getFileInfo({
        file_id: Number(fileId),
      })
      return {
        forest_name: res.Forest.name,
        file_name: res.name,
        parent_ids: res.parent_id_arr ?? [],
        parent_paths: res.parent_path_arr ?? [],
      }
    },
    { refreshDeps: [id, fileId] },
  )

  const fileType = useMemo(() => {
    const file_name = fileDetail?.file_name
    if (!file_name) return null
    return file_name.split('.').at(-1)
  }, [fileDetail?.file_name])

  const breadcrumbItems = useMemo(() => {
    if (!fileDetail) return []
    const { file_name, forest_name, parent_ids, parent_paths } = fileDetail
    const className =
      'text-[#86909C] font-medium text-base cursor-pointer hover:text-[#0C99FF]'
    const titles = [
      <Link className={className} to={'/docs'}>
        知识库
      </Link>,
    ]
    if (forest_name) {
      titles.push(
        <Link className={className} to={`/docs/cad/${id}`}>
          {forest_name}
        </Link>,
      )
    }
    const parentFolders = (parent_ids as number[]).map(
      (id: number, i: number) => {
        return { id, name: parent_paths[i], level: i }
      },
    )
    titles.push(
      ...parentFolders.map((folder, i) => {
        const pathToFolder = parentFolders.slice(0, i)
        const params = encodeURIComponent(JSON.stringify(pathToFolder))
        return (
          <Link
            className={className}
            to={`/docs/cad/${id}/folder/${folder.id}?path=${params}`}
          >
            {folder.name}
          </Link>
        )
      }),
    )
    titles.push(
      <span className='text-[#000000E5] font-medium text-base'>
        {file_name}
      </span>,
    )
    return titles.map((title) => ({ title }))
  }, [fileDetail, id])

  if (!fileDetail) {
    return (
      <div className='w-full h-full flex justify-center items-center bg-gray-50'>
        <Spin size='large' />
      </div>
    )
  }

  return (
    <FileLocationProvider>
      <div className='h-full overflow-hidden p-6 bg-white'>
        <div className='w-full h-full rounded-lg overflow-hidden flex flex-col gap-5'>
          {/* 骨架导航 */}
          <div className='flex-shrink-0'>
            <Breadcrumb
              separator='>'
              className='text-base'
              items={breadcrumbItems}
            />
          </div>

          {/* 主要内容区域 - 可拖动分割的左右布局 */}
          <div className='flex-1 overflow-hidden'>
            <ResizablePanel
              leftPanel={
                <FilePreview
                  fileName={fileDetail.file_name}
                  fileType={fileType}
                />
              }
              rightPanel={<RightPanel />}
              initialLeftWidth={50}
              minLeftWidth={30}
              minRightWidth={30}
            />
          </div>
        </div>
      </div>
    </FileLocationProvider>
  )
}
