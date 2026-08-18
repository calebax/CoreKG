import { FC } from 'react'
import { Button, Spin } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'
import { useBoolean } from 'ahooks'
import { Link } from 'react-router-dom'
import { cn } from '@/utils'
import DownIcon from '@/assets/icons/down.svg?react'
import UpIcon from '@/assets/icons/up.svg?react'
import { AIDialog } from '..'

export type ReferenceFiles = {
  className?: string
  loading: boolean
  files: AIDialog['reference']
}
/** 参考文献 */
export const ReferenceFiles: FC<ReferenceFiles> = (props) => {
  const { files, loading, className } = props
  const [expanded, { toggle }] = useBoolean(false)
  if (!loading && files.length === 0) {
    // 无参考文献且加载过程已经结束
    return null
  }
  return (
    <div className={cn('flex flex-col gap-2', className)}>
      <span className='self-start flex items-center relative'>
        {loading ? (
          <Spin indicator={<LoadingOutlined spin />} size='small' />
        ) : null}
        {files.length === 0 ? (
          <span className='text-[#1E1F28] text-base font-normal leading-[22px] ml-[10px]'>
            正在搜索知识库资料...
          </span>
        ) : (
          <Button
            type='text'
            size='small'
            className='p-0 text-[#1E1F28] text-base font-normal leading-[22px]'
            style={{
              marginLeft: loading ? '10px' : '0',
              fontFamily: 'Inter , sans-serif',
            }}
            onClick={toggle}
            icon={expanded ? <UpIcon /> : <DownIcon />}
            iconPosition='end'
          >
            找到了{files.length}篇知识库文件作为参考
          </Button>
        )}
      </span>

      {expanded ? (
        <div className={cn('py-2.5 flex flex-col gap-2.5', 'rounded')}>
          {files.slice(0, 200).map((item, i) => (
            <FileTitle {...item} index={i} key={i} />
          ))}
        </div>
      ) : null}
    </div>
  )
}

type FileTitle = ReferenceFiles['files'][number] & { index: number }
const FileTitle: FC<FileTitle> = (props) => {
  const { file_name, forest_id, file_id, index } = props
  return (
    <Link
      className={cn(
        'pl-[6px] text-[#616373] text-base font-medium leading-[22px] py-1 px-1.5',
        'hover:text-[#1E1F28] hover:bg-[#EFF0F6]',
      )}
      to={`/docs/detail/${forest_id}/file/${file_id}`}
      target='_blank'
    >
      {index + 1}.{file_name}
    </Link>
  )
}
