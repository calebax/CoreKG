import { useNavigate } from 'react-router-dom'
import { App } from 'antd'
import { match } from 'ts-pattern'
import { useLoginGlobalData } from '@/utils/useLoginGlobalData'

export type QuotaLimitType = 'knowledge' | 'agent' | 'member'

/**
 * 显示额度受限弹窗
 * @example
 * const { show } = useQuotaLimitModal()
 * show({ type: 'knowledge' })
 * show({ type: 'agent', content: '自定义内容' })
 */
export const useQuotaLimitModal = () => {
  const { modal } = App.useApp()
  const navigate = useNavigate()
  const { version } = useLoginGlobalData()

  const show = (options: { type?: QuotaLimitType; content?: string }) => {
    const { type, content: customContent } = options

    const defaultContent = (() => {
      const actionText = match(type)
        .with('knowledge', () => '知识库容量已达上限')
        .with('agent', () => '无法创建新的智能体')
        .with('member', () => '无法创建新的成员')
        .otherwise(() => '')
      return `${actionText}，具体使用情况可前往「个人信息」页面查看,如需扩容可继续续费。`
    })()
    const content = customContent || defaultContent

    modal.confirm({
      title: '温馨提示',
      icon: null,
      content: <div className='text-[#616373] text-base'>{content}</div>,
      okText: '立即扩容',
      onOk: () => {
        navigate('/settings/profile')
      },
      centered: true,
    })
  }

  return {
    show,
  }
}
