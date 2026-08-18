import { FC, useState, useEffect } from 'react'
import { Modal, Form, Input, Button, message } from 'antd'
import { createIndustryTerm, updateIndustryTerm } from '@/api/knowledge'

interface IndustryTermModalProps {
  open: boolean
  onCancel: () => void
  onSuccess: () => void
  editingItem?: any
  detailData?: any
}

const CharacterCount: FC<{ value: string; max: number }> = ({
  value = '',
  max,
}) => (
  <span className='text-gray-400 text-xs'>
    {value?.length || 0}/{max}
  </span>
)

const IndustryTermModal: FC<IndustryTermModalProps> = ({
  open,
  onCancel,
  onSuccess,
  editingItem,
  detailData,
}) => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const word = Form.useWatch('word', form)
  const description = Form.useWatch('description', form)

  // 检查是否可以保存
  const isSaveDisabled = !word?.trim() || !description?.trim()

  useEffect(() => {
    if (open) {
      if (detailData) {
        form.setFieldsValue({
          word: detailData.word,
          description: detailData.description,
        })
      } else {
        form.resetFields()
      }
    }
  }, [open, detailData, form])

  const handleFinish = async (values: any) => {
    setLoading(true)
    try {
      if (editingItem) {
        await updateIndustryTerm({
          id: editingItem.ID,
          word: values.word,
          description: values.description,
        })
        message.success('修改成功')
      } else {
        await createIndustryTerm({
          word: values.word,
          description: values.description,
        })
        message.success('创建成功')
      }
      onSuccess()
    } catch (error) {
      console.log(error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title={editingItem ? '编辑行业术语' : '创建行业术语'}
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
        className='mt-6'
        labelCol={{ span: 5, style: { paddingRight: '12px' } }}
        wrapperCol={{ span: 19 }}
        labelAlign='left'
      >
        <Form.Item label='术语名称' required className='mb-6'>
          <Form.Item
            name='word'
            noStyle
            rules={[
              { required: true, message: '请输入术语名称' },
              { max: 20, message: '术语名称最多20个字符' },
            ]}
          >
            <Input
              placeholder='请输入'
              maxLength={20}
              suffix={<CharacterCount value={word} max={20} />}
            />
          </Form.Item>
        </Form.Item>

        <Form.Item label='定义' required className='mb-6'>
          <Form.Item
            name='description'
            noStyle
            rules={[
              { required: true, message: '请输入定义' },
              { max: 200, message: '定义最多200个字符' },
            ]}
          >
            <Input.TextArea
              placeholder='请输入'
              maxLength={200}
              rows={4}
              showCount={{
                formatter: ({ count, maxLength }) => (
                  <span className='text-gray-400 text-xs'>
                    {count}/{maxLength}
                  </span>
                ),
              }}
            />
          </Form.Item>
        </Form.Item>
      </Form>
    </Modal>
  )
}

export default IndustryTermModal
