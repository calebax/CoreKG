import { FC } from 'react'
import { Button, Form, Input, Select } from 'antd'
import { useBoolean } from 'ahooks'
import { cn } from '@/utils'
import { scroll } from '@/styles/scroll'
import { type InputParams } from '..'

interface ParamsProps {
  agentDetail: any
  startQA: (input: InputParams) => Promise<void>
}

export default function Params({ agentDetail, startQA }: ParamsProps) {
  const [loading, { toggle }] = useBoolean()
  const params: any[] = agentDetail.agent_info.params
  return (
    <div
      className={cn('flex-1 overflow-auto p-6', 'flex flex-col', scroll)}
      style={{
        background: 'linear-gradient(90deg, #F7F7FF 0.03%, #F5FCFF 101.25%)',
      }}
    >
      <h1 className='text-title text-center font-bold text-[35px]'>
        {agentDetail.show_name}
      </h1>
      <Form
        layout='vertical'
        className='max-w-[620px] w-full mx-auto mt-10'
        onFinish={async (formValue) => {
          toggle()
          try {
            const input: InputParams = params.map((p) => {
              const name = p.input
              const value = formValue[name]
              const title = p.name
              return {
                title,
                value,
                name,
              }
            })
            await startQA(input)
          } catch {
            toggle()
          }
        }}
      >
        {params.map((param) => (
          <Form.Item
            key={param.input}
            name={param.input}
            rules={[
              {
                required: param.is_required,
                message:
                  (param.input_type === 'text' ? '请输入' : '请选择') +
                  param.name,
              },
            ]}
            label={
              <div className='flex items-center gap-2'>
                <span className='text-lg'>{param.name}</span>
                <span className='text-description text-sm'>
                  ({param.input})
                </span>
              </div>
            }
          >
            <ParamItem param={param} />
          </Form.Item>
        ))}
        <Form.Item>
          <Button
            type='primary'
            className='w-full h-11! flex items-center justify-center gap-2'
            loading={loading}
            htmlType='submit'
          >
            <span className='text-base font-medium'>立即生成</span>
          </Button>
        </Form.Item>
      </Form>
    </div>
  )
}

const ParamItem: FC<ValueController<string> & { param: any }> = (props) => {
  const { param, value, onChange } = props
  if (param.input_type === 'text') {
    return (
      <Input.TextArea
        placeholder={`请输入${param.name}`}
        value={value}
        onChange={(e) => onChange?.(e.target.value)}
        autoSize
        className='min-h-11! max-h-100'
      />
    )
  }
  return (
    <Select
      placeholder={`请选择${param.name}`}
      value={value}
      onChange={onChange}
      options={param.input_array.map((v: string) => ({ label: v, value: v }))}
    ></Select>
  )
}
