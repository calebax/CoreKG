import { FC } from 'react'
import { Avatar } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'
import { useRequest } from 'ahooks'
import { cn } from '@/utils'
import { uploadImage } from '@/api/common'
import { loadFile } from '@/utils/loadFile'
import Icon from './uploadIcon.svg?react'

export type AgentAvator = {
  className?: string
  style?: React.CSSProperties
  value?: string
  onChange?: (val: string) => void
}
export const AgentAvator: FC<Style & ValueController<string>> = (props) => {
  const { value, onChange, className, style } = props
  const { loading, run } = useRequest(
    async (file: File) => {
      const { url } = (await uploadImage(
        { file, purpose: 'yg-chat' },
        {
          timeout: 0,
          headers: { 'Content-Type': 'multipart/form-data' },
        },
      )) as any
      onChange?.(url)
    },
    { manual: true },
  )
  return (
    <div className={cn('w-16 h-16 relative', className)} style={style}>
      {loading ? (
        <LoadingOutlined
          className={cn('z-10 absolute left-1/2 top-1/2 -translate-1/2')}
        />
      ) : (
        <Avatar className={cn('w-full h-full')} src={value} />
      )}

      <Icon
        className={cn('absolute right-0 bottom-0 z-10', {
          'cursor-pointer': !loading,
        })}
        onClick={() => {
          if (loading) return
          loadFile(
            (file) => {
              run(file[0])
            },
            { accept: 'image/*' },
          )
        }}
      />
    </div>
  )
}
