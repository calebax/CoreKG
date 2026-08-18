import { FC } from 'react'
import { Button, Form, Input, Popover } from 'antd'
import { DownOutlined } from '@ant-design/icons'
import { useMount } from 'ahooks'
import { getSessionInfo } from '@/api'
import { cn } from '@/utils'
import { scroll } from '@/styles/scroll'
import { Dialog, DialogInput } from './DialogComponents'
import Params from './Params'
import { useDialog } from './useDialog'

export type InputParams = {
  // input
  name: string
  value?: string
  // name
  title: string
}[]
export type Content = {
  agentDetail: any
  session_id?: number
  answering: ValueController<boolean>
  /** 下次切换session_id本组件不会被重置 */
  setSessionId: (id: number) => void
}
export const Content: FC<Style & Content & { workflow?: boolean }> = (
  props,
) => {
  const { agentDetail, session_id, answering, setSessionId } = props
  const { dialog, startQA, startQAFirst, historyLoading } = useDialog(
    agentDetail,
    session_id,
    setSessionId,
    answering.onChange,
  )
  const [input, setInput] = useState<InputParams>()
  useMount(async () => {
    if (!session_id) return
    const res = await getSessionInfo({ id: session_id })
    const _input = res.input as any[]
    setInput(
      _input.map((item) => {
        const { name, value } = item
        const par = agentDetail.agent_info.params.find(
          (item: any) => item.input === name,
        )!
        return {
          name,
          value,
          title: par.name,
        }
      }),
    )
  })

  if (!session_id) {
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
  return (
    <div
      className={cn('flex-1 overflow-hidden', 'bg-[#F8FCFF]', ' flex flex-col')}
    >
      <div className={cn('flex-0 p-6 w-full max-w-250 mx-auto flex')}>
        <h1 className='flex-1 text-title font-bold text-[28px] text-center font-alimama-thin'>
          {agentDetail.show_name}
        </h1>
        <InputViewBtn input={input} />
      </div>
      <div
        className={cn('flex-1 overflow-hidden flex flex-col', props.className)}
      >
        <Dialog dialog={dialog} loading={historyLoading} />
        {props.workflow ? null : (
          <DialogInput
            onSend={startQA}
            disalbed={Boolean(historyLoading || answering.value)}
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
          className={cn(' w-150 max-w-150 max-h-100 p-4 overflow-auto', scroll)}
          layout='vertical'
        >
          {input?.map((param) => {
            if (!param.value) return
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
        className={cn('flex-0', { hidden: !input })}
      >
        查看
      </Button>
    </Popover>
  )
}
