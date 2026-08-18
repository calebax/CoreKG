import { FC } from 'react'
import { Button } from 'antd'
import { cn } from '@/utils'
import { ForestAnalyzeFiles, type AnalyzeFile } from './ForestAnalyzeFiles'
import { ForestUploadFiles } from './ForestUploadFiles'

export const ForestUpload: FC<{
  forest_id: number
  closeModal?: () => void
  analyzeFiles?: AnalyzeFile[]
  reloadAnalyzeFiles: () => void
}> = (props) => {
  const { forest_id, closeModal, analyzeFiles, reloadAnalyzeFiles } = props
  const [type, setType] = useState<'analyze' | 'upload'>('upload')

  return (
    <div className='w-full h-full overflow-hidden'>
      <div
        className={cn(
          'rounded-sm bg-white overflow-hidden',
          'p-0 w-full h-full flex gap-2.5',
        )}
      >
        <div
          className={cn(
            'border border-[#D7D9E5] rounded-sm',
            'flex-none w-44 p-4 flex flex-col gap-3',
          )}
        >
          <Button
            type='text'
            block
            className={cn('justify-start', {
              'bg-[#EFF0F6]': type === 'analyze',
            })}
            onClick={() => setType('analyze')}
          >
            文件解析{analyzeFiles ? `（${analyzeFiles.length}）` : null}
          </Button>
          <Button
            type='text'
            block
            className={cn('justify-start', {
              'bg-[#EFF0F6]': type === 'upload',
            })}
            onClick={() => setType('upload')}
          >
            文件上传
          </Button>
        </div>
        <div className='flex-1 overflow-hidden'>
          <ForestAnalyzeFiles
            files={analyzeFiles}
            reload={reloadAnalyzeFiles}
            closeModal={closeModal}
            className={cn({ hidden: type !== 'analyze' })}
          />
          <ForestUploadFiles
            forest_id={forest_id}
            onUploadOne={reloadAnalyzeFiles}
            closeModal={closeModal}
            className={cn({ hidden: type !== 'upload' })}
          />
        </div>
      </div>
    </div>
  )
}
