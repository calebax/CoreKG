import { FC } from 'react'
import { Avatar } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'
import { useRequest } from 'ahooks'
import { cn } from '@/utils'
import { uploadImage } from '@/api/common'
import { loadFile } from '@/utils/loadFile'
import { getDefaultAvatar } from '../../getDefaultAvatar'
import AddAvatarIcon from '../images/add-avatar.svg?react'
import styles from './styles.module.scss'
import Icon from './uploadIcon.svg?react'
import { BasicAgentInfo } from 'Agent'
export type AgentAvator = {
  className?: string
  style?: React.CSSProperties
  value?: string
  onChange?: (val: string) => void
  type?: BasicAgentInfo['type']
}
export const AgentAvator: FC<AgentAvator> = (props) => {
  const { value, onChange, type, className, style } = props
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
    <div
      className={cn('w-16 h-16 relative', className, styles.agentAvator)}
      style={style}
    >
      {loading ? (
        <LoadingOutlined
          className={cn('z-10 absolute left-1/2 top-1/2 -translate-1/2')}
        />
      ) : (
        <Avatar
          className={cn('w-full h-full rounded-lg')}
          src={value === 'default' ? getDefaultAvatar(type ?? 'role_play') : value}
        />
      )}
      <div
        className={styles.agentAvatorMask}
        onClick={() => {
          if (loading) return
          loadFile(
            (file) => {
              run(file[0])
            },
            { accept: 'image/*' },
          )
        }}
      >
        更换图标
      </div>
      {/* <AddAvatarIcon
        className={cn('absolute right-[-3px] bottom-[-4px] z-10', {
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
      /> */}
    </div>
  )
}
