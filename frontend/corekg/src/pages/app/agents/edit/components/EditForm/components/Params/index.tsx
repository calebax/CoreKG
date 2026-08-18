import { FC, useState } from 'react'
import { Tooltip, Input, Button, Form } from 'antd'
import { FormListProps } from 'antd/es/form'
import { PlusOutlined, QuestionCircleOutlined } from '@ant-design/icons'
import { Agent } from 'Agent'
import { useBoolean } from 'ahooks'
import { cn } from '@/utils'
import { ParamModal } from './ParamModal'
import DelIcon from './images/delete.svg?react'
import EditIcon from './images/edit.svg?react'

export type { AgentEditValue } from '../..'
/** 指令型机器人参数 */
export const Params: FC<Style> = (props) => {
  const type = Form.useWatch<Agent['type']>('type', { preserve: true })
  if (type !== 'prompt') return null
  return (
    <Form.List
      name='params'
      rules={[
        {
          validator: async (_, val) => {
            if (!val || val.length === 0) {
              return Promise.reject('至少需要一个变量')
            }
          },
        },
      ]}
    >
      {(...args) => {
        return <InnerParams {...props} args={args} />
      }}
    </Form.List>
  )
}
const InnerParams: FC<
  Style & {
    args: Parameters<FormListProps['children']>
  }
> = (props) => {
  const [fields, operation, meta] = props.args
  const form = Form.useFormInstance()

  const [open, { toggle }] = useBoolean(false)
  const [paramIndex, setIndex] = useState<number | undefined>()
  return (
    <>
      <div
        className={cn('mt-2 flex flex-col gap-1', props.className)}
        style={props.style}
      >
        <div className='flex items-center justify-between'>
          <div className='flex items-center'>
            <span className='text-base text-title font-medium'>变量</span>
            <Tooltip
              placement='right'
              title='支持在提示词中引用自定义变量，在调用智能体时传入变量值，可对智能体的输出进行调整'
            >
              <QuestionCircleOutlined className='text-[#C9CDD4] ml-1' />
            </Tooltip>
          </div>
          <Button
            type='text'
            size='small'
            icon={<PlusOutlined />}
            className='flex items-center gap-1 text-[#0C99FF] text-base font-medium px-[2px]'
            onClick={() => {
              setIndex(undefined)
              toggle()
            }}
          >
            添加
          </Button>
        </div>
        {fields.map((field, i) => {
          const { name, key } = field
          return (
            <div className='flex gap-2.5 items-baseline' key={key}>
              <Form.Item
                name={[name, 'input']}
                rules={[{ required: true, message: '请输入变量名' }]}
                className='flex-1 m-0'
              >
                <Input />
              </Form.Item>
              <Form.Item
                name={[name, 'name']}
                rules={[{ required: true, message: '请输入显示名称' }]}
                className='flex-1 m-0'
              >
                <Input />
              </Form.Item>
              <EditIcon
                className='cursor-pointer'
                onClick={() => {
                  setIndex(i)
                  toggle()
                }}
              />
              <DelIcon
                className='cursor-pointer'
                onClick={() => {
                  operation.remove(i)
                }}
              />
            </div>
          )
        })}
        {meta.errors.length > 0 ? (
          <Form.Item className='m-0'>
            <Form.ErrorList errors={meta.errors}></Form.ErrorList>
          </Form.Item>
        ) : null}
      </div>
      <ParamModal
        open={open}
        onClose={toggle}
        title={Number.isInteger(paramIndex) ? '编辑变量' : '添加变量'}
        value={
          Number.isInteger(paramIndex)
            ? form.getFieldValue(['params', paramIndex])
            : undefined
        }
        onSubmit={(val) => {
          if (Number.isInteger(paramIndex)) {
            form.setFieldValue(['params', paramIndex], val)
          } else {
            operation.add(val)
          }
        }}
      />
    </>
  )
}
