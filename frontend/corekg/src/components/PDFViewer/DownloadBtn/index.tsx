import { FC, type ReactNode } from 'react'
import { Button } from 'antd'
import { useBoolean } from 'ahooks'
import { downloadForestFile } from '@/utils/Forest'
import DownLoadIcon from './download.svg?react'

export type DownLoadBtnProps = {
  id: any
  name: string
  icon?: ReactNode
}
export const DownLoadBtn: FC<DownLoadBtnProps> = (props) => {
  const { id, name, icon } = props
  const [loading, { toggle }] = useBoolean(false)
  return (
    <Button
      type='text'
      icon={icon ?? <DownLoadIcon />}
      loading={loading}
      onClick={async () => {
        toggle()
        await downloadForestFile(Number(id), name)
        toggle()
      }}
      className='!min-w-[32px] !w-8 !h-8 hover:bg-gray-100'
    />
  )
}
