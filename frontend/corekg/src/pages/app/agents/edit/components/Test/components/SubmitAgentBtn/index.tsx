import { FC } from 'react'
import { useNavigate } from 'react-router-dom'
import { App, Button } from 'antd'
import { useRequest } from 'ahooks'
import { publishAgent } from '@/api'
import { getAgentUrl } from '@/pages/app/agents/utils/getAgentUrl'
import { AgentEditValue } from '../..'
import { useEditContext } from '../../../AgentContext'

export const SubmitAgentBtn: FC<
  Style & {
    tested?: boolean
    getAgentEditValue: () => Promise<AgentEditValue>
  }
> = (props) => {
  const { getAgentEditValue, tested } = props
  const { message } = App.useApp()
  const navigate = useNavigate()
  const { agent } = useEditContext()
  const { run: submit, loading } = useRequest(
    async (formValue: AgentEditValue) => {
      await publishAgent({
        ...formValue,
        agent_type: formValue.type,
        avatar_url: formValue.avatar,
        show_name: formValue.title,
        chat_model_ids: formValue.chat_models?.map((item) => item.id),
        doc_forest_ids:
          formValue.type !== 'prompt'
            ? formValue.forests?.map((item) => item.id)
            : null,
      })
      message.success('发布成功')
      const { coze_workflow_id, coze_space_id } = agent
      navigate(
        getAgentUrl(formValue.id, formValue.type, false, {
          coze_workflow_id,
          coze_space_id,
        }),
      )
    },
    { manual: true },
  )
  return (
    <Button
      className='bg-[#0C99FF] hover:bg-[#0C99FF] text-white'
      loading={loading}
      onClick={() => {
        if (!tested) {
          message.warning('调试未通过暂不支持发布')
          return
        }
        getAgentEditValue().then(submit, () => {
          message.warning('请完善智能体配置')
        })
      }}
    >
      发布
    </Button>
  )
}
