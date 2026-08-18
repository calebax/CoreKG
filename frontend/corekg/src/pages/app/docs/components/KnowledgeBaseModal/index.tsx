import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Modal, Form, Input, message, Radio, Spin, Tooltip, Select } from 'antd'
import { useRequest } from 'ahooks'
import { uniqueArray } from '@/utils'
import {
  createKnowledgeBase,
  getKnowledgeBaseDetail,
  updateForestWithPerm,
} from '@/api/knowledge'
import SelectUser from '@/components/form/SelectUser'
import useLocalStore from '@/stores/local'
import scrollStyles from '@/styles/scroll/styles.module.scss'
import { useAdmin } from '@/utils/useAdmin'
import DB from './images/db.svg'
import Excel from './images/excel.svg'
import ODS from './images/ods.svg'
import QA from './images/qa.svg'
import styles from './index.module.scss'

const KNOWLEDGE_BASE_TYPES = [
  {
    key: 'file',
    name: '多模态',
    icon: ODS,
    description:
      '支持上传多种类型的文件，适用于汇总文档资料、图片等多源信息的综合问答场景。',
    iconText: 'ODS',
  },

  {
    key: 'excel',
    name: '表格',
    icon: Excel,
    description:
      '支持上传 Excel、CSV 等表格文件，适用于商品清单、财务报表等结构化数据场景。',
    iconText: 'Excel',
  },
  {
    key: 'db',
    name: '数据库',
    icon: DB,
    description:
      '可连接MySQL等外部数据库，适用于需要实时访问业务系统数据的场景。',
    iconText: 'DB',
  },

  {
    key: 'qa',
    name: '问答对',
    icon: QA,
    description:
      '支持支持导入标准问答对（如常见问题与答案），适用于客服资料、产品FAQ等问答场景。',
    iconText: 'QA',
  },
] as const

// 即将上线的知识库类型
const COMING_SOON_TYPES: string[] = ['cad']

interface KnowledgeBaseFormValues {
  type?: string
  title: string
  description: string
  manager_ids?: number[]
  public_scope: 'company' | 'private' | 'custom'
  scope_ids?: number[]
}

interface KnowledgeBaseData {
  id: number
  type?: string
  forest_type?: string
  name?: string
  description?: string
  loading?: boolean
  manager_ids?: number[]
  public_scope?: 'company' | 'private' | 'custom'
  scope_ids?: number[]
}

interface KnowledgeBaseModalProps {
  open: boolean
  mode?: 'add' | 'edit'
  data?: KnowledgeBaseData
  onCancel: () => void
  onSuccess?: () => void
}

function KnowledgeBaseModal({
  open,
  mode = 'add',
  data,
  onCancel,
  onSuccess,
}: KnowledgeBaseModalProps) {
  const { adminIds } = useAdmin()

  const { uinId } = useLocalStore((state) => state.userInfo)
  const [form] = Form.useForm<KnowledgeBaseFormValues>()
  const navigate = useNavigate()
  const publicScope = Form.useWatch('public_scope', form)
  const description = Form.useWatch('description', form)
  const title = Form.useWatch('title', form)
  const selectedType = Form.useWatch('type', form)
  const [submitLoading, setSubmitLoading] = useState(false)
  const [isDataLoading, setIsDataLoading] = useState(false)

  useEffect(() => {
    if (!open) {
      form.resetFields()
      setIsDataLoading(false)
      return
    }

    if (mode === 'add') {
      const initManagerIds = [uinId as any as number]
      // 新建模式：重置表单并设置默认值
      form.resetFields()
      form.setFieldsValue({
        type: 'file', // 默认选中多模态类型
        description: '',
        public_scope: 'custom',
        manager_ids: initManagerIds,
        scope_ids: initManagerIds,
      })
      setSubmitLoading(false)
    } else if (mode === 'edit' && data) {
      // 编辑模式：处理数据加载
      setIsDataLoading(data.loading || false)

      if (!data.loading && data.name) {
        const _scope_ids: number[] = data.scope_ids ?? []
        const _manager_ids: number[] = data.manager_ids ?? []
        const formValues: KnowledgeBaseFormValues = {
          // 编辑模式下不设置 type 字段
          title: data.name || '',
          description: data.description || '',
          manager_ids: data.manager_ids || [],
          public_scope: data.public_scope || 'custom',
          scope_ids: uniqueArray(_manager_ids, _scope_ids),
        }
        form.setFieldsValue(formValues)
        setIsDataLoading(false)
      }
    }
  }, [open, mode, data, form, uinId, adminIds])

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      console.log(values)

      setSubmitLoading(true)
      const data_source_type: 'standard' | 'excel' | 'db' = (() => {
        switch (values.type) {
          case 'excel':
            return 'excel'
          case 'db':
            return 'db'
          default:
            return 'standard'
        }
      })()
      const data_source_subtype: 'standard' | 'excel' | 'mysql' = (() => {
        switch (data_source_type) {
          case 'excel':
            return 'excel'
          case 'db':
            return 'mysql'
          default:
            return 'standard'
        }
      })()
      if (mode === 'add') {
        // 新建知识库
        const createBody = {
          forest_type: ['excel', 'db'].includes(values.type!)
            ? 'data'
            : values.type!,
          name: values.title,
          description: values.description,
          public_scope: values.public_scope,
          data_source_type,
          data_source_subtype,
        }

        const { forest_id } = await createKnowledgeBase(createBody)

        // await updateForestWithPerm({
        //   id: forest_id,
        //   ...createBody,
        //   manager_ids: [uinId as any],
        //   public_scope: values.public_scope,
        //   scope_ids: values.scope_ids,
        // })

        message.success('创建成功')

        form.resetFields()
        onCancel()

        // 根据知识库类型跳转到不同页面
        if (values.type === 'file') {
          navigate(`/docs/detail/${forest_id}`)
        } else if (values.type === 'qa') {
          navigate(`/docs/qa/${forest_id}`, {
            state: { is_admin: true },
          })
        } else if (values.type === 'cad') {
          navigate(`/docs/cad/${forest_id}`)
        } else if (values.type === 'excel') {
          // 表格知识库创建成功后与多模态一致进入通用详情页
          navigate(`/docs/detail/${forest_id}`)
        } else if (values.type === 'db') {
          navigate(`/docs/db/${forest_id}`)
        }
        onSuccess?.()
      } else {
        // 编辑知识库
        const updateBody = {
          id: data!.id,
          name: values.title,
          description: values.description,
          manager_ids: values.manager_ids!,
          public_scope: values.public_scope,
          scope_ids: values.scope_ids,
          data_source_type,
          data_source_subtype,
        }
        await updateForestWithPerm(updateBody)
        message.success('修改成功')
        form.resetFields()
        onCancel()
        onSuccess?.()
      }
    } catch (error) {
      console.error('操作失败:', error)
    } finally {
      setSubmitLoading(false)
    }
  }

  const handleComingSoonClick = (typeName: string) => {
    message.info(`${typeName}即将上线，敬请期待`)
  }

  const modalTitle = mode === 'add' ? '创建知识库' : '编辑知识库'
  const isLoading = mode === 'edit' && isDataLoading

  const shouldDisableCreateButton =
    mode === 'add' &&
    (!selectedType || // 未选择知识库类型
      COMING_SOON_TYPES.includes(selectedType) ||
      !title || // 未输入知识库名称
      !title.trim()) // 知识库名称为空字符串

  const renderKnowledgeBaseTypeSelector = () => (
    <Form.Item
      label='知识库类型'
      name='type'
      rules={[{ required: true, message: '请选择知识库类型' }]}
      className={styles.createFormItem}
    >
      <div className='grid grid-cols-4 gap-[12px]'>
        {KNOWLEDGE_BASE_TYPES.map((type) => {
          const isComingSoon = COMING_SOON_TYPES.includes(type.key)
          const isSelected = selectedType === type.key

          const typeCard = (
            <div
              key={type.key}
              className={`
                relative py-4 px-3 border-[1.5px] rounded-lg transition-all
                ${
                  isSelected
                    ? 'border-[#0C99FF]'
                    : isComingSoon
                      ? 'border-gray-200 bg-gray-50 cursor-not-allowed opacity-60'
                      : 'border-gray-200 hover:border-gray-300 cursor-pointer'
                }
              `}
              onClick={() => {
                if (isComingSoon) {
                  handleComingSoonClick(type.name)
                } else if (!isSelected) {
                  form.setFieldValue('type', type.key)
                }
              }}
            >
              <div className='flex flex-col items-start gap-3'>
                <div className='flex items-center gap-[12px]'>
                  <div className='flex items-center justify-center border-1 border-[#E5E7EB] rounded-[5px] bg-[#fff] w-[52px] h-[52px] shadow-[4px_4px_20px_0px_#00000014]'>
                    <img
                      src={type.icon}
                      alt={type.name}
                      className={`w-6 h-6 ${isComingSoon ? 'opacity-60' : ''}`}
                    />
                  </div>
                  <div
                    className={`font-medium text-[18px] text-[#0C1F17] ${
                      isComingSoon ? 'text-gray-400' : 'text-gray-900'
                    }`}
                  >
                    {type.name}
                  </div>
                </div>
                <div
                  className={`text-[12px] leading-[1.4em] ${
                    isComingSoon ? 'text-gray-400' : 'text-[#6E757F]'
                  }`}
                >
                  {type.description}
                </div>
              </div>
            </div>
          )

          // 如果是即将上线的类型，包装在 Tooltip 中
          if (isComingSoon) {
            return (
              <Tooltip
                key={type.key}
                title='即将上线，敬请期待'
                placement='top'
              >
                {typeCard}
              </Tooltip>
            )
          }

          return typeCard
        })}
      </div>
    </Form.Item>
  )

  const renderKnowledgeBaseTypeDisplay = () => {
    const typeKey = data?.forest_type || data?.type
    const currentType =
      typeKey === 'data'
        ? KNOWLEDGE_BASE_TYPES[3]
        : KNOWLEDGE_BASE_TYPES.find((type) => type.key === typeKey)

    if (!currentType) {
      return (
        <div className='mb-6'>
          <span className='text-gray-700'>知识库类型：</span>
          <span className='text-gray-900 font-medium'>未知类型</span>
        </div>
      )
    }

    return (
      <div className='mb-6'>
        <span className='text-gray-700'>知识库类型：</span>
        <span className='text-gray-900 font-medium'>{currentType.name}</span>
      </div>
    )
  }

  const scope_ids = Form.useWatch('scope_ids', { form, preserve: true }) ?? []
  const manager_ids =
    Form.useWatch('manager_ids', { form, preserve: true }) ?? []
  return (
    <Modal
      title=''
      open={open}
      onCancel={onCancel}
      onOk={handleSubmit}
      okText={mode === 'add' ? '创建' : '确定'}
      cancelText='取消'
      confirmLoading={submitLoading}
      destroyOnClose
      width={mode === 'add' ? 912 : 520}
      okButtonProps={{
        disabled: shouldDisableCreateButton,
        className: styles.submit,
      }}
      cancelButtonProps={{
        className: styles.cancel,
      }}
      style={{ top: 40 }}
      bodyStyle={{
        padding: '0',
        maxHeight: 'calc(100vh - 150px)',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <div className='flex items-center justify-between mb-6'>
        <h1 className='text-gray-900 text-lg font-medium'>{modalTitle}</h1>
      </div>

      <div className={`flex-1 overflow-y-auto pr-2 ${scrollStyles.scroll}`}>
        {isLoading ? (
          <div className='flex justify-center items-center py-20'>
            <Spin tip='加载中...' />
          </div>
        ) : (
          <Form
            form={form}
            layout='vertical'
            className='space-y-6'
            onValuesChange={(changed) => {
              if ('manager_ids' in changed) {
                if (changed.manager_ids.length > manager_ids.length) {
                  // 如果选中了管理员 将其添加至scoped_ids
                  form.setFieldValue(
                    'scope_ids',
                    uniqueArray(adminIds, changed.manager_ids, scope_ids),
                  )
                }
              } else if ('scope_ids' in changed) {
                // 管理员不能被从scope_id中删除
                form.setFieldValue(
                  'scope_ids',
                  uniqueArray(adminIds, manager_ids, changed.scope_ids),
                )
              }
            }}
          >
            {/* 知识库类型选择（新建模式）或显示（编辑模式） */}
            {mode === 'add'
              ? renderKnowledgeBaseTypeSelector()
              : renderKnowledgeBaseTypeDisplay()}

            <Form.Item
              label='知识库名称'
              name='title'
              className={styles.createFormItem}
            >
              <div className='w-full'>
                <Input
                  className={`h-[32px] ${styles.inputWrap}`}
                  placeholder='请输入知识库名称'
                  maxLength={50}
                  value={title}
                />
                {/* <p className='mt-1 text-xs text-gray-500 text-right'>
                  {title?.length || 0}/15
                </p> */}
              </div>
            </Form.Item>

            <Form.Item
              className={styles.createFormItem}
              label='知识库描述'
              name='description'
            >
              <div className='w-full'>
                <Input.TextArea
                  placeholder='请输入知识库描述'
                  maxLength={200}
                  rows={2}
                  value={description}
                  className={`resize-none ${styles.inputWrap}`}
                />
              </div>
            </Form.Item>
            {/* {mode !== 'edit' && ['excel', 'db'].includes(selectedType!) ? (
              <Form.Item
                className={styles.createFormItem}
                name='type'
                label='数据类型'
              >
                <Radio.Group>
                  <Radio value='excel'>Excel文件</Radio>
                  <Radio value='db'>MySQL数据库</Radio>
                </Radio.Group>
              </Form.Item>
            ) : null} */}
            {/* 管理员字段仅在编辑模式显示 */}
            {mode === 'edit' && (
              <Form.Item
                className={styles.createFormItem}
                label={<span className='text-gray-900'>管理员</span>}
                name='manager_ids'
                rules={[{ required: true, message: '请选择管理员' }]}
              >
                <SelectUser
                  className='w-full'
                  placeholder='请选择管理员'
                  mode='multiple'
                  value={manager_ids}
                  onChange={(value: number[]) =>
                    form.setFieldValue(
                      'manager_ids',
                      uniqueArray(manager_ids, value),
                    )
                  }
                />
              </Form.Item>
            )}

            <Form.Item
              className={styles.createFormItem}
              style={{
                display: 'none',
              }}
              label={<span className='text-gray-900'>公开范围</span>}
              name='public_scope'
              rules={[{ required: true, message: '请选择公开范围' }]}
            >
              <Radio.Group className='flex gap-6'>
                <Radio value='company'>公司</Radio>
                <Radio value='custom'>自定义</Radio>
              </Radio.Group>
            </Form.Item>

            {mode === 'edit' && publicScope === 'custom' && (
              <Form.Item
                name='scope_ids'
                rules={[{ required: true, message: '请选择自定义范围' }]}
              >
                <SelectUser
                  className='w-full'
                  placeholder='请选择范围'
                  mode='multiple'
                  value={scope_ids}
                  onChange={(value: number[]) =>
                    form.setFieldValue('scope_ids', value)
                  }
                />
              </Form.Item>
            )}
          </Form>
        )}
      </div>
    </Modal>
  )
}

export default KnowledgeBaseModal
