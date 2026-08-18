import { FC } from 'react'
import { Button, Form, FormInstance } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { cn } from '@/utils'
import type { AgentEditValue } from '../..'
import { getAgentUrl } from '../../../utils/getAgentUrl'
import { useEditContext } from '../AgentContext'
import { Basic } from './components/Basic'
import { Forests } from './components/Forests'
import { Greeting } from './components/Greeting'
import { Manager } from './components/Manager'
import { ModelSelect } from './components/ModelSelect'
import { Params } from './components/Params'
import PermissionManagement from './components/PermissionManagement'
import { Prompt } from './components/Prompt'
import { PublicScope } from './components/PublicScope'
import { ScopeIds } from './components/ScopeIds'
import { Temperature } from './components/Temperature'

// eslint-disable-next-line react-refresh/only-export-components
export { useEditContext } from '../AgentContext'
export type { AgentEditValue } from '../..'

export const FormItem = Form.Item<AgentEditValue>

export const EditForm: FC<
  Style & {
    form: FormInstance<AgentEditValue>
  }
> = (props) => {
  const { form } = props
  const { agent } = useEditContext()
  const { coze_workflow_id, coze_space_id } = agent
  if (agent.type === 'workflow') {
    return (
      <Form
        form={form}
        initialValues={agent}
        className={cn('flex flex-col gap-2 h-full relative', props.className)}
        style={props.style}
      >
        <div className='mb-4 flex items-center'>
          <span className='text-base text-title font-medium'>高级配置</span>
          <Link
            target='_blank'
            to={getAgentUrl(0, 'workflow', false, {
              toCoze: true,
              coze_workflow_id,
              coze_space_id,
            })}
            className={cn(
              'text-[#0C99FF] text-base font-medium px-[2px]',
              'ml-auto flex items-center gap-1',
            )}
          >
            <PlusOutlined />
            去编排
          </Link>
        </div>
        <PermissionManagement form={form} />
        <div className='hidden'>
          <Manager />
          <PublicScope />
          <ScopeIds />
        </div>
      </Form>
    )
  }
  return (
    <Form
      form={form}
      initialValues={agent}
      className={cn('flex flex-col gap-2 h-full relative', props.className)}
      style={props.style}
    >
      {/* <Basic /> */}
      <ModelSelect />
      <Prompt />
      <Greeting />
      <Forests />
      <Params />
      <span className='text-base text-title font-medium mt-2 -mb-2'>
        模型参数
      </span>
      <Temperature />

      {/* 隐藏原有的Manager、PublicScope和ScopeIds组件，因为现在使用PermissionManagement */}
      {/* 使用绝对定位的隐藏容器，确保不占用任何布局空间 */}
      <div className='absolute -left-[9999px] opacity-0 pointer-events-none'>
        <Manager />
        <PublicScope />
        <ScopeIds />
      </div>

      {/* 新的权限管理组件 */}
      <span className='text-base text-title font-medium -mb-2'>权限管理</span>
      <PermissionManagement form={form} />
    </Form>
  )
}
