import { FC, useMemo } from 'react'
import { Form, Input, Modal, Skeleton, Button, Spin } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'
import { GraphBaseInfo } from 'Graph'
import { useRequest } from 'ahooks'
import { cn, uniqueArray } from '@/utils'
import useLocalStore from '@/stores/local'
import { useAdmin } from '@/utils/useAdmin'
import PermissionSection from './PermissionSection'
import { GraphMetaValues } from './types'

export type GraphModal = {
  loading?: boolean
  title?: string
  onCancel?: () => void
  okText?: string
  onOk?: (val: GraphMetaValues) => any
  /** 有此值则是编辑 */
  initialValues?: GraphMetaValues
}
const FormItem = Form.Item<GraphMetaValues>

export const GraphModal: FC<GraphModal> = (props) => {
  const { loading, initialValues, onCancel, onOk: _onOk, ...rest } = props
  const { adminIds } = useAdmin()
  const currentUin: number = useLocalStore(
    (state) => state.userInfo.uinId,
  ) as any

  /** 新建图谱时默认的管理员 */
  const defaultManagerIds = useMemo(() => {
    return [currentUin]
  }, [adminIds, currentUin])

  const [form] = Form.useForm<GraphMetaValues>()

  const isEdit = !!initialValues

  // 监听表单字段变化
  const managerIds =
    Form.useWatch('manager_ids', form) ||
    (isEdit ? initialValues?.manager_ids : defaultManagerIds)
  const publicScope =
    Form.useWatch('public_scope', form) ||
    initialValues?.public_scope ||
    'custom'
  const scopeIds =
    Form.useWatch('scope_ids', form) ||
    initialValues?.scope_ids ||
    defaultManagerIds

  const { run: onOk, loading: submitting } = useRequest(
    async () => {
      const formValue = await form.validateFields()
      await _onOk?.(formValue)
    },
    { manual: true },
  )

  const modalDisabled = submitting || loading

  return (
    <Modal
      open={true}
      onCancel={onCancel}
      {...rest}
      footer={
        loading ? null : (
          <div className='flex justify-end gap-2'>
            <button
              className='px-6 py-2 bg-[#F5F5F5] text-[#0C1F17] rounded-md text-sm font-medium cursor-pointer hover:bg-[#F5F5F5] transition-colors'
              onClick={onCancel}
              disabled={modalDisabled}
            >
              取消
            </button>
            <button
              className={`px-6 py-2 rounded-md text-sm font-medium flex items-center gap-2 cursor-pointer ${
                !modalDisabled
                  ? 'bg-[#0C99FF] text-[#ffffff] hover:bg-[#0C99FF]'
                  : 'bg-[#0C99FF] text-[#ffffff] opacity-50 cursor-not-allowed'
              }`}
              onClick={onOk}
              disabled={modalDisabled}
            >
              {submitting ? (
                <LoadingOutlined className='text-white' spin />
              ) : null}
              {rest.okText || (isEdit ? '提交' : '确定')}
            </button>
          </div>
        )
      }
      keyboard={!modalDisabled}
      maskClosable={!modalDisabled}
      closable={!modalDisabled}
    >
      {loading ? (
        <Skeleton active />
      ) : (
        <Form
          initialValues={
            initialValues ?? {
              manager_ids: defaultManagerIds,
              scope_ids: defaultManagerIds,
              public_scope: 'custom',
            }
          }
          form={form}
          layout='vertical'
        >
          <FormItem
            name={'name'}
            label='名称'
            rules={[{ required: true, message: '请填写名称' }]}
          >
            <Input maxLength={60} showCount />
          </FormItem>
          <FormItem name={'description'} label='描述'>
            <Input.TextArea maxLength={100} showCount />
          </FormItem>

          {/* 隐藏的权限字段 - 确保表单包含这些字段 */}
          <FormItem name={'manager_ids'} hidden>
            <Input />
          </FormItem>
          <FormItem name={'public_scope'} hidden>
            <Input />
          </FormItem>
          <FormItem name={'scope_ids'} hidden>
            <Input />
          </FormItem>

          {/* 权限管理部分 */}
          <div className='mb-4 hidden'>
            <PermissionSection
              isEdit={isEdit}
              managerIds={managerIds || []}
              publicScope={publicScope || 'company'}
              scopeIds={scopeIds || []}
              onManagerIdsChange={(ids) => {
                form.setFieldValue('manager_ids', ids)
              }}
              onPublicScopeChange={(scope) => {
                form.setFieldValue('public_scope', scope)
              }}
              onScopeIdsChange={(ids) => {
                form.setFieldValue('scope_ids', ids)
              }}
              form={form}
            />
          </div>
        </Form>
      )}
    </Modal>
  )
}
