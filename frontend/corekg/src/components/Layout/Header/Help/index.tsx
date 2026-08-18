import { FC } from 'react'
import { Tooltip } from 'antd'
import { cn } from '@/utils'
import HelpIcon from './help.svg?react'

export const Help: FC<Style> = (props) => {
  return null
  return (
    <Tooltip placement='bottom' title='帮助文档'>
      <a
        target='_blank'
        href='https://docs.corekg.com/docs/corekg/'
        className={cn('cursor-pointer', props.className)}
        style={props.style}
      >
        <HelpIcon />
      </a>
    </Tooltip>
  )
}
