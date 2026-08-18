import { FC } from 'react'
import { App, Avatar } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'
import { useRequest } from 'ahooks'
import { match } from 'ts-pattern'
import { cn } from '@/utils'
import { uploadAvatarImg, uploadWebsiteLogo } from '@/api/organization'
import { loadFile } from '@/utils/loadFile'
import styles from './styles.module.scss'

export const EditableAvatar: FC<
  Style & ValueController<string> & { type: 'company_logo' | 'website_logo' }
> = (props) => {
  const { value, onChange, type, className, style } = props
  const { message } = App.useApp()
  const { run: uploadFile, loading } = useRequest(
    async (file: File) => {
      const validTypes = ['image/png', 'image/jpeg', 'image/jpg']
      if (!validTypes.includes(file.type)) {
        message.error('仅支持 jpg/jpeg/png 格式的图片')
        return
      }
      const isLt5M = file.size / 1024 / 1024 < 5
      if (!isLt5M) {
        message.error('图片大小不能超过5M')
        return
      }

      const url = await match(type)
        .with('company_logo', () =>
          uploadAvatarImg(
            { file, purpose: 'company-logo' },
            {
              timeout: 0,
              headers: { 'Content-Type': 'multipart/form-data' },
            },
          ).then((res: any) => res.public_url),
        )
        .with('website_logo', () =>
          uploadWebsiteLogo(
            {
              file,
              purpose: 'company-logo',
            },
            {
              headers: { 'Content-Type': 'multipart/form-data' },
            },
          ).then((res: any) => res.public_url),
        )
        .exhaustive()
      onChange?.(url)
    },
    { manual: true },
  )
  return (
    <div
      className={cn('relative w-12 h-12', className, styles.agentAvator)}
      style={style}
    >
      {loading ? (
        <LoadingOutlined
          className={cn('z-10 absolute left-1/2 top-1/2 -translate-1/2')}
        />
      ) : (
        <Avatar
          shape='circle'
          className={cn('w-full h-full rounded-lg')}
          src={value}
        />
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
