import React, { useState } from 'react'
import { Modal, Form, Input, Select, DatePicker, Button, message } from 'antd'
import locale from 'antd/es/date-picker/locale/zh_CN'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import { useTranslation } from 'react-i18next'

interface CreateApiKeyModalProps {
  visible: boolean
  onCancel: () => void
  onOk: (values: { name: string; expireDate: string }) => void
}

const CreateApiKeyModal: React.FC<CreateApiKeyModalProps> = ({
  visible,
  onCancel,
  onOk,
}) => {
  const { i18n, t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const [form] = Form.useForm()
  const [expireType, setExpireType] = useState<string | null>(null)
  const [customExpireDate, setCustomExpireDate] = useState<dayjs.Dayjs | null>(
    null,
  )
  const [loading, setLoading] = useState(false)

  // 设置dayjs语言为中文
  dayjs.locale(i18n.language)

  // 获取到期日期
  const getExpireDate = () => {
    const today = dayjs()

    switch (expireType) {
      case '一天':
        return today.add(1, 'day').format('YYYY-MM-DD')
      case '30天':
        return today.add(30, 'day').format('YYYY-MM-DD')
      case '自定义':
        return customExpireDate ? customExpireDate.format('YYYY-MM-DD') : ''
      default:
        return ''
    }
  }

  const handleOk = async () => {
    try {
      await form.validateFields()

      if (!expireType) {
        message.error(t('profile.selectExpiryDate'))
        return
      }

      if (expireType === '自定义' && !customExpireDate) {
        message.error(t('profile.selectExpiryDate'))
        return
      }

      const name = form.getFieldValue('name')
      const expireDate = getExpireDate()

      setLoading(true)
      onOk({ name, expireDate })

      form.resetFields()
      setExpireType(null)
      setCustomExpireDate(null)
    } catch (error) {
      console.error('表单验证失败', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCancel = () => {
    form.resetFields()
    setExpireType(null)
    setCustomExpireDate(null)
    onCancel()
  }

  return (
    <Modal
      title={
        <div className='text-center text-lg font-medium mt-4 mb-6'>
          {tC('button.create', { target: 'API key' })}
        </div>
      }
      open={visible}
      onCancel={handleCancel}
      footer={null}
      destroyOnHidden
      width={460}
      className='create-api-key-modal'
      centered
      closeIcon={<span className='text-[#86909C] text-3xl'>×</span>}
    >
      <Form
        form={form}
        layout='vertical'
        preserve={false}
        requiredMark={false}
        className='py-6'
      >
        <Form.Item
          name='name'
          label={
            <div className='flex items-center mb-2'>
              <span className='text-[#F53F3F] mr-1'>*</span>
              <span className='text-[#1D2129] text-base font-500'>
                {tC('custom.name')}:
              </span>
            </div>
          }
          rules={[
            { required: true, message: t('profile.pleaseEnterApiKeyName') },
          ]}
        >
          <Input
            placeholder={t('profile.pleaseEnterApiKeyName')}
            className='h-12 border border-gray-300 hover:!border-[#BEDAFF] focus:!border-[#BEDAFF] !rounded !text-base'
          />
        </Form.Item>

        <Form.Item
          label={
            <div className='flex items-center mb-2'>
              <span className='text-[#F53F3F] mr-1'>*</span>
              <span className='text-[#1D2129] text-base font-500'>
                到期日期：
              </span>
            </div>
          }
        >
          <div className='flex w-full gap-4'>
            {expireType === '自定义' ? (
              <>
                <Select
                  value={expireType}
                  onChange={(value) => setExpireType(value)}
                  className='w-[45%] !h-12'
                  options={[
                    {
                      value: '一天',
                      label: (
                        <div className='flex justify-start items-center gap-2 w-full py-2 text-[#1D2129] text-base'>
                          <span>
                            {t('profile.day', { target: t('profile.one') })}
                          </span>
                          <span>
                            ({dayjs().add(1, 'day').format('YYYY-MM-DD')})
                          </span>
                        </div>
                      ),
                      className: 'hover:!bg-[#E8F3FF]',
                    },
                    {
                      value: '30天',
                      label: (
                        <div className='flex justify-start items-center gap-2 w-full py-2 text-[#1D2129] text-base'>
                          <span>{t('profile.day', { target: 30 })}</span>
                          <span>
                            ({dayjs().add(30, 'day').format('YYYY-MM-DD')})
                          </span>
                        </div>
                      ),
                      className: 'hover:!bg-[#E8F3FF]',
                    },
                    { value: '自定义', label: t('profile.customize') },
                  ]}
                />
                <DatePicker
                  value={customExpireDate}
                  onChange={(date) => setCustomExpireDate(date)}
                  className='w-[55%] h-12 border-gray-300 hover:!border-[#BEDAFF] focus:!border-[#BEDAFF] !rounded !text-base'
                  placeholder={t('profile.pleaseSelectDate')}
                  format='YYYY-MM-DD'
                  disabledDate={(current) =>
                    current && current < dayjs().endOf('day')
                  }
                  locale={locale}
                />
              </>
            ) : (
              <Select
                value={expireType}
                onChange={(value) => setExpireType(value)}
                placeholder={t('profile.pleaseSelectExpirationDate')}
                className='w-full !h-12 p-[6px]'
                popupClassName='expire-select-dropdown'
                options={[
                  {
                    value: '一天',
                    label: (
                      <div className='flex justify-start items-center gap-2 w-full py-2 text-[#1D2129] text-base'>
                        <span>
                          {t('profile.day', { target: t('profile.one') })}
                        </span>
                        <span>
                          ({dayjs().add(1, 'day').format('YYYY-MM-DD')})
                        </span>
                      </div>
                    ),
                    className: 'hover:!bg-[#E8F3FF]',
                  },
                  {
                    value: '30天',
                    label: (
                      <div className='flex justify-start items-center gap-2 w-full py-2 text-[#1D2129] text-base'>
                        <span>{t('profile.day', { target: 30 })}</span>
                        <span>
                          ({dayjs().add(30, 'day').format('YYYY-MM-DD')})
                        </span>
                      </div>
                    ),
                    className: 'hover:!bg-[#E8F3FF]',
                  },
                  {
                    value: '自定义',
                    label: (
                      <div className='py-2 text-[#1D2129] text-base'>
                        {t('profile.customize')}
                      </div>
                    ),
                    className: 'hover:!bg-[#E8F3FF]',
                  },
                ]}
                listHeight={180}
                dropdownStyle={{
                  padding: '4px 0',
                }}
              />
            )}
          </div>
        </Form.Item>

        <Form.Item className='!mb-0 flex justify-end pt-10'>
          <Button
            className='mr-4 h-12  border border-gray-300 !text-base'
            onClick={handleCancel}
          >
            {tC('button.cancel')}
          </Button>
          <Button
            type='primary'
            onClick={handleOk}
            loading={loading}
            className='h-12 bg-blue-500 !text-base'
          >
            {tC('button.confirm')}
          </Button>
        </Form.Item>
      </Form>
    </Modal>
  )
}

export default CreateApiKeyModal
