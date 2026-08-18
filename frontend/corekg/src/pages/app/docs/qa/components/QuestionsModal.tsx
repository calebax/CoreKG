import { useState, useEffect } from 'react'
import { Drawer, Form, Input, Button, message, Space } from 'antd'
import { PlusOutlined, DeleteOutlined, CloseOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { createQAPair, modifyQAPair } from '@/api/knowledge'

interface QAPair {
  id: string
  qa_question: string
  qa_answer: string
  sub_questions?: Array<{
    id: string
    question: string
    created_at: string
  }>
}

interface QuestionsModalProps {
  open: boolean
  mode: 'add' | 'edit'
  data?: QAPair
  forestId: number
  onCancel: () => void
  onSuccess: () => void
}

interface FormValues {
  question: string
  answer: string
  subQuestions: Array<{
    id?: string // 原有子问题的ID
    question: string
    created_at?: string
  }>
}

export default function QuestionsModal({
  open,
  mode,
  data,
  forestId,
  onCancel,
  onSuccess,
}: QuestionsModalProps) {
  const { t: tM } = useTranslation('messages')
  const { t: tC } = useTranslation('common')
  const { t } = useTranslation('pages')
  const [form] = Form.useForm<FormValues>()
  const [loading, setLoading] = useState(false)
  // 保存原始子问题数据，用于追踪删除
  const [originalSubQuestions, setOriginalSubQuestions] = useState<
    Map<string, any>
  >(new Map())

  useEffect(() => {
    if (open && mode === 'edit' && data) {
      // 构建原始子问题映射
      const originalMap = new Map()
      data.sub_questions?.forEach((sq) => {
        originalMap.set(sq.id, sq)
      })
      setOriginalSubQuestions(originalMap)

      form.setFieldsValue({
        question: data.qa_question,
        answer: data.qa_answer,
        subQuestions: data.sub_questions?.map((sq) => ({
          id: sq.id,
          question: sq.question,
          created_at: sq.created_at,
        })) || [{ question: '' }],
      })
    } else if (open && mode === 'add') {
      form.resetFields()
      setOriginalSubQuestions(new Map())
      form.setFieldsValue({
        subQuestions: [{ question: '' }],
      })
    }
  }, [open, mode, data, form])

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      if (mode === 'add') {
        await createQAPair({
          forest_id: forestId,
          question: values.question,
          answer: values.answer,
          sub_question: values.subQuestions
            .map((sq) => sq.question)
            .filter((q) => q.trim()),
        })
        message.success(tM('createSuccess'))
      } else {
        // 处理编辑模式的子问题变更
        const currentSubQuestions = values.subQuestions.filter((sq) =>
          sq.question.trim(),
        )
        const childQuestions: Array<{
          id?: string
          question: string
          is_deleted: boolean
          created_at?: string
        }> = []

        // 1. 处理当前表单中的子问题（新增或修改）
        currentSubQuestions.forEach((sq) => {
          childQuestions.push({
            id: sq.id,
            question: sq.question,
            is_deleted: false,
            created_at: sq.created_at,
          })
        })

        // 2. 找出被删除的子问题
        const currentIds = new Set(
          currentSubQuestions.map((sq) => sq.id).filter(Boolean),
        )
        originalSubQuestions.forEach((originalSq, id) => {
          if (!currentIds.has(id)) {
            // 这个原有的子问题被删除了
            childQuestions.push({
              id: id,
              question: originalSq.question,
              is_deleted: true,
              created_at: originalSq.created_at,
            })
          }
        })

        console.log('提交的子问题数据:', childQuestions)

        await modifyQAPair({
          main: {
            id: data!.id,
            forest_id: forestId,
            type: 'FQA',
            qa_question: values.question,
            qa_answer: values.answer,
          },
          child: childQuestions,
        })
        message.success(tM('modifySuccess'))
      }

      onSuccess()
    } catch (error) {
      console.error('操作失败:', error)
      message.error(tM(mode === 'add' ? 'createFailure' : 'modifyFailure'))
    } finally {
      setLoading(false)
    }
  }

  const title = tC(mode === 'add' ? 'button.addNew' : 'button.edit', {
    target: t('app.docs.qaContent'),
  })

  return (
    <Drawer
      title={
        <div className='flex justify-between items-center w-full'>
          <span>{title}</span>
          <Button
            type='text'
            icon={<CloseOutlined />}
            onClick={onCancel}
            className='text-gray-400 hover:text-gray-600'
          />
        </div>
      }
      placement='right'
      width={560}
      open={open}
      onClose={onCancel}
      closable={false}
      extra={
        <Space className='absolute bottom-4 right-4'>
          <Button onClick={onCancel}>{tC('button.cancel')}</Button>
          <Button loading={loading} onClick={handleSubmit} className='bg-[#0C99FF] hover:bg-[#0C99FF] text-white'>
            {tC('button.save')}
          </Button>
        </Space>
      }
    >
      <Form form={form} layout='vertical'>
        <Form.Item
          label={t('app.docs.question')}
          name='question'
          rules={[
            {
              required: true,
              message: t('app.docs.inputContent', {
                target: t('app.docs.question'),
              }),
            },
          ]}
        >
          <Input.TextArea
            placeholder={t('app.docs.inputContent', {
              target: t('app.docs.question'),
            })}
            rows={4}
          />
        </Form.Item>

        <Form.Item
          label={t('app.docs.similarContent', {
            target: t('app.docs.question'),
          })}
        >
          <Form.List name='subQuestions'>
            {(fields, { add, remove }) => (
              <div className='space-y-2'>
                {fields.map(({ key, name, ...restField }, index) => (
                  <div key={key} className='flex items-center gap-2'>
                    <span className='text-gray-500 w-3'>{index + 1}.</span>
                    <Form.Item
                      {...restField}
                      name={[name, 'question']}
                      className='flex-1 mb-0'
                    >
                      <Input
                        placeholder={t('app.docs.inputContent', {
                          target: t('app.docs.similarContent', {
                            target: t('app.docs.question'),
                          }),
                        })}
                      />
                    </Form.Item>
                    {fields.length > 1 && (
                      <Button
                        type='text'
                        icon={<DeleteOutlined />}
                        onClick={() => remove(name)}
                        className='text-gray-400 hover:text-red-500'
                      />
                    )}
                    {/* 隐藏字段保存ID和创建时间 */}
                    <Form.Item
                      {...restField}
                      name={[name, 'id']}
                      className='hidden mb-0'
                    >
                      <Input type='hidden' />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'created_at']}
                      className='hidden mb-0'
                    >
                      <Input type='hidden' />
                    </Form.Item>
                  </div>
                ))}
                <Button
                  type='dashed'
                  onClick={() => add({ question: '' })}
                  icon={<PlusOutlined />}
                  className='w-full'
                >
                  {t('app.docs.addContent', {
                    target: t('app.docs.similarContent', {
                      target: t('app.docs.question'),
                    }),
                  })}
                </Button>
              </div>
            )}
          </Form.List>
        </Form.Item>

        <Form.Item
          label={t('app.docs.answer')}
          name='answer'
          rules={[
            {
              required: true,
              message: t('app.docs.inputContent', {
                target: t('app.docs.answer'),
              }),
            },
          ]}
        >
          <Input.TextArea
            placeholder={t('app.docs.inputContent', {
              target: t('app.docs.answer'),
            })}
            rows={8}
            showCount
            maxLength={1000}
          />
        </Form.Item>
      </Form>
    </Drawer>
  )
}
