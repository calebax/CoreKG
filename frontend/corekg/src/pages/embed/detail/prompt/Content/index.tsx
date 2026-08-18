import { FC, useState } from 'react'
import { Button, Form, Input, Popover } from 'antd'
import { DownOutlined } from '@ant-design/icons'
import { cn } from '@/utils'
import { scroll } from '@/styles/scroll'
import { Dialog, DialogInput } from './DialogComponents'
import Params from './Params'
import { useIframePromptDialog } from './useIframePromptDialog'

export type InputParams = {
  // input
  name: string
  value?: string
  // name
  title: string
}[]

export type Content = {
  agentDetail: any
  answering: ValueController<boolean>
  workflow?: boolean
}

export const Content: FC<Style & Content> = (props) => {
  const { agentDetail, answering, workflow } = props
  const { dialog, startQA, startQAFirst, historyLoading, hasStarted } =
    useIframePromptDialog(agentDetail, answering.onChange, workflow)
  const [input, setInput] = useState<InputParams>()

  // 如果还没开始对话，显示参数输入界面
  if (!hasStarted) {
    return (
      <Params
        agentDetail={agentDetail}
        startQA={async (val) => {
          setInput(val)
          await startQAFirst(val)
        }}
      />
    )
  }

  // 已开始对话，显示对话界面
  return (
    <div
      className={cn(
        'flex-1 overflow-hidden w-full h-full',
        'bg-[#F8FCFF]',
        'flex flex-col',
      )}
    >
      <div className={cn('flex-none p-6 w-full max-w-250 mx-auto flex')}>
        <h1 className='flex-1 text-title font-bold text-[28px] text-center font-alimama-thin'>
          {agentDetail.show_name}
        </h1>
        <InputViewBtn input={input} />
      </div>
      <div
        className={cn('flex-1 overflow-hidden flex flex-col', props.className)}
      >
        <Dialog dialog={dialog} loading={historyLoading} />
        {workflow ? null : (
          <DialogInput
            onSend={startQA}
            disabled={Boolean(historyLoading || answering.value)}
          />
        )}
      </div>
    </div>
  )
}

export const InputViewBtn: FC<{ input?: InputParams }> = (props) => {
  const { input } = props
  return (
    <Popover
      trigger={['click']}
      placement='bottom'
      content={
        <Form
          className={cn('w-150 max-w-150 max-h-100 p-4 overflow-auto', scroll)}
          layout='vertical'
        >
          {input?.map((param) => {
            if (!param.value) return null
            return (
              <Form.Item
                initialValue={param.value}
                key={param.name}
                name={param.name}
                layout='vertical'
                label={
                  <div className='flex items-center gap-2'>
                    <span className='text-lg'>{param.title}</span>
                    <span className='text-description text-sm'>
                      ({param.name})
                    </span>
                  </div>
                }
              >
                <Input disabled />
              </Form.Item>
            )
          })}
        </Form>
      }
    >
      <Button
        type='link'
        icon={<DownOutlined />}
        iconPosition='end'
        className={cn('flex-none', { hidden: !input })}
      >
        查看
      </Button>
    </Popover>
  )
}
