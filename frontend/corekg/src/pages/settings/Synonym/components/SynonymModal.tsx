import { FC, useState, useEffect } from 'react'
import { Modal, Form, Input, Button, message } from 'antd'
import { PlusOutlined, MinusOutlined } from '@ant-design/icons'
import { createSynonymKeyword, updateSynonymKeyword } from '@/api/knowledge'

interface SynonymModalProps {
  open: boolean
  onCancel: () => void
  onSuccess: () => void
  editingItem?: any
  detailData?: any
}

const CharacterCount: FC<{ value: string; max?: number }> = ({
  value = '',
  max = 20,
}) => (
  <span className='text-gray-400 text-xs'>
    {value?.length || 0}/{max}
  </span>
)

const SynonymModal: FC<SynonymModalProps> = ({
  open,
  onCancel,
  onSuccess,
  editingItem,
  detailData,
}) => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const word = Form.useWatch('word', form)
  const synonyms = Form.useWatch('synonyms', form)

  // 检查是否可以保存
  const isSaveDisabled =
    !word?.trim() || !synonyms?.some((s: any) => s?.word?.trim())

  useEffect(() => {
    if (open) {
      if (detailData) {
        console.log(detailData)
        form.setFieldsValue({
          word: detailData.word,
          synonyms: detailData.synonym_keywords.map((item: any) => ({
            id: item.ID,
            word: item.word,
          })),
        })
      } else {
        form.resetFields()
        form.setFieldsValue({
          synonyms: [{ word: '' }],
        })
      }
    }
  }, [open, detailData, form])

  const handleFinish = async (values: any) => {
    const { word, synonyms } = values
    const validSynonyms = synonyms.filter((s: any) => s && s.word.trim())

    if (validSynonyms.length === 0) {
      message.warning('请至少输入一个有效的同义词')
      return
    }

    setLoading(true)
    try {
      if (editingItem) {
        await updateSynonymKeyword({
          id: editingItem.ID,
          word,
          child_words: validSynonyms.map((s: any) => s.word),
        })
        message.success('修改成功')
      } else {
        await createSynonymKeyword({
          word,
          child_words: validSynonyms.map((s: any) => s.word),
        })
        message.success('创建成功')
      }
      onSuccess()
    } catch (error) {
      console.error(error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title={editingItem ? '编辑同义词组' : '创建同义词组'}
      open={open}
      onCancel={onCancel}
      onOk={() => form.submit()}
      confirmLoading={loading}
      okButtonProps={{
        disabled: isSaveDisabled,
        className: `h-8 px-[15px] rounded-[6px] w-[77px] ${
          isSaveDisabled
            ? '!bg-[#0c99ff] !opacity-50 !text-white'
            : '!bg-[#0c99ff] !text-white hover:!bg-[#0a8ae6]'
        }`,
      }}
      cancelButtonProps={{
        className:
          'h-8 px-[15px] rounded-[6px] w-[77px] !bg-[#f1f5f9] !text-[#0c1f17] !border-0 hover:!bg-[#e2e8f0]',
      }}
      okText='保存'
      cancelText='取消'
      width={500}
      destroyOnClose
    >
      <Form
        form={form}
        layout='horizontal'
        onFinish={handleFinish}
        initialValues={{ synonyms: [{ word: '' }] }}
        className='mt-6'
        labelCol={{ span: 5, style: { paddingRight: '12px' } }}
        wrapperCol={{ span: 19 }}
        labelAlign='left'
      >
        <Form.Item label='主词' required className='mb-6'>
          <div className='flex items-center gap-2'>
            <Form.Item
              name='word'
              noStyle
              rules={[
                { required: true, message: '请输入主词' },
                { max: 20, message: '主词最多20个字符' },
              ]}
            >
              <Input
                placeholder='请输入主词'
                maxLength={20}
                suffix={<CharacterCount value={word} />}
                className='flex-1'
              />
            </Form.Item>
            <div className='w-[80px]' />
          </div>
        </Form.Item>

        <div className='mb-6 text-sm font-medium text-[rgba(0,0,0,0.88)]'>
          同义词组
        </div>

        <Form.List name='synonyms'>
          {(fields, { add, remove }) => (
            <>
              {fields.map((field, index) => (
                <Form.Item
                  {...field}
                  key={field.key}
                  label={`同义词${index + 1}`}
                  required
                  className='mb-6'
                  name={undefined} // 必须清除，否则会冲突
                >
                  <div className='flex items-center gap-2'>
                    <Form.Item
                      {...field}
                      name={[field.name, 'word']}
                      noStyle
                      rules={[
                        { required: true, message: '请输入同义词' },
                        { max: 20, message: '最多20个字符' },
                      ]}
                    >
                      <Input
                        placeholder='请输入同义词'
                        maxLength={20}
                        suffix={
                          <CharacterCount
                            value={synonyms?.[field.name]?.word}
                          />
                        }
                        className='flex-1'
                      />
                    </Form.Item>
                    <div className='flex items-center gap-2 w-[80px] shrink-0'>
                      <div className='w-[32px] flex justify-start'>
                        {index === fields.length - 1 && fields.length < 5 ? (
                          <Button
                            type='link'
                            onClick={() => add()}
                            className='p-0 h-auto text-[#0C99FF] flex items-center'
                          >
                            添加
                          </Button>
                        ) : (
                          fields.length > 1 && (
                            <Button
                              type='link'
                              danger
                              onClick={() => remove(field.name)}
                              className='p-0 h-auto text-red-500 hover:text-red-400 flex items-center'
                            >
                              删除
                            </Button>
                          )
                        )}
                      </div>
                    </div>
                  </div>
                </Form.Item>
              ))}
            </>
          )}
        </Form.List>
      </Form>
    </Modal>
  )
}

export default SynonymModal
