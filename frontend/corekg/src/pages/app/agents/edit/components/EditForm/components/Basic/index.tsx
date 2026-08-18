import { FC } from 'react'
import { Form } from 'antd'
import { useBoolean } from 'ahooks'
import { cn } from '@/utils'
import { AgentModal } from './AgentModal'
import EditIcon from './images/edit.svg?react'

/** 头像 名称 描述 */
export const Basic: FC<Style> = (props) => {
  const [open, { toggle }] = useBoolean(false)
  const form = Form.useFormInstance()
  const title = Form.useWatch('title', { preserve: true })
  const avatar = Form.useWatch('avatar', { preserve: true })
  const description = Form.useWatch('description', { preserve: true })

  return (
    <>
      <div
        className={cn('w-full flex items-center gap-3', props.className)}
        style={props.style}
      >
        <div className='flex-none w-12.5 h-12.5 rounded-full overflow-hidden bg-[#E5E9EF]'>
          <img src={avatar} alt='' className='w-full h-full object-cover' />
        </div>
        <div className='flex-grow w-full'>
          <div className='flex items-center gap-2'>
            <h1 className='text-title font-medium'>{title}</h1>
            <EditIcon
              className='cursor-pointer text-[#86909C]'
              onClick={toggle}
            />
          </div>
          <h2 className='text-description text-sm'>{description}</h2>
        </div>
      </div>
      <AgentModal
        open={open}
        onClose={toggle}
        value={{ avatar, title, description }}
        onSubmit={(val) => {
          form.setFieldValue('avatar', val.avatar)
          form.setFieldValue('title', val.title)
          form.setFieldValue('description', val.description)
        }}
      />
    </>
  )
}
