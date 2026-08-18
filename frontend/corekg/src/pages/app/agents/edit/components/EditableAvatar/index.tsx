import { FC } from 'react'
import { Avatar } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'
import { useRequest } from 'ahooks'
import { cn } from '@/utils'
import { uploadImage } from '@/api/common'
import { loadFile } from '@/utils/loadFile'
import styles from './styles.module.scss'

export const EditableAvatar: FC<Style & ValueController<string>> = (props) => {
  const { value, onChange, className, style } = props
  const { run: uploadFile, loading } = useRequest(
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
    <div
      className={cn('relative', className, styles.agentAvator)}
      style={style}
    >
      {loading ? (
        <LoadingOutlined
          className={cn('z-10 absolute left-1/2 top-1/2 -translate-1/2')}
        />
      ) : (
        <Avatar className={cn('w-full h-full rounded-lg')} src={value} />
      )}
      {loading ? null : (
        <div
          className={cn('text-xs', styles.agentAvatorMask)}
          onClick={() => {
            if (loading) return
            loadFile(
              (file) => {
                uploadFile(file[0])
              },
              { accept: 'image/*' },
            )
          }}
        >
          更换
          <br />
          图标
        </div>
      )}
    </div>
  )
}
