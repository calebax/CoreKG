import { FC } from 'react'
import { App, Button, Form, Input, Result, Select } from 'antd'
import { useBoolean, useCountDown, useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { sleep } from '@/utils'
import { sendUpgradeFormCode, submitUpgradeForm } from '@/api/agent'
import { companyIndustry, companyScale } from './options'

type FormType = 'contact' | 'upgrade'
export const VersionForm: FC = () => {
  const { t } = useTranslation('pages')
  const formType: FormType = 'contact'
  const title = t('version.contactPreSales')
  const [success, { toggle }] = useBoolean()
  if (success) {
    return (
      <Result
        status='success'
        title={t('version.applicationSubmitSuccess')}
        subTitle={t('version.contactYouSoonWait')}
      ></Result>
    )
  }
  return (
    <div className='flex flex-col'>
      <span className='text-2xl font-semibold text-center mt-6'>{title}</span>
      <span className='text-[#181A1F] text-center mt-2'>
        {t('version.leaveAppInfoContactSoon')}
      </span>
      <FormContent
        type={formType}
        className='mx-[150px] mt-11'
        onSuccess={toggle}
      />
    </div>
  )
}

const FormContent: FC<Style & { onSuccess: () => void; type: FormType }> = (
  props,
) => {
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const { onSuccess, type } = props
  const { run, loading } = useRequest(
    async (formValue: any) => {
      await submitUpgradeForm({
        ...formValue,
        type,
      })
      onSuccess()
    },
    {
      manual: true,
    },
  )
  return (
    <Form
      layout='vertical'
      onFinish={run}
      className={props.className}
      style={props.style}
    >
      <Form.Item
        name='name'
        label={t('version.name')}
        rules={[
          {
            required: true,
            message: t('version.pleaseEnter', { target: t('version.name') }),
          },
        ]}
      >
        <Input />
      </Form.Item>
      <Form.Item
        name='phone'
        label={t('version.phone')}
        required
        rules={[
          {
            validator: async (_, value: string = '') => {
              const numbers = value
                .split('')
                .map(Number)
                .filter(Number.isInteger)
              if (numbers.length !== 11)
                throw new Error(t('version.pleasePhone'))
            },
          },
        ]}
      >
        <Input />
      </Form.Item>
      <Form.Item
        name='code'
        rules={[
          {
            required: true,
            message: t('version.pleaseEnter', { target: t('version.code') }),
          },
        ]}
      >
        <Code type={type} />
      </Form.Item>
      <Form.Item
        name='company_name'
        label={t('version.companyName')}
        rules={[
          {
            required: true,
            message: t('version.pleaseEnter', {
              target: t('version.companyName'),
            }),
          },
        ]}
      >
        <Input />
      </Form.Item>
      <Form.Item name='scale' label={t('version.companyScale')}>
        <Select
          options={companyScale.map((v) => {
            const val = t(v as any) as string
            return { label: val, value: val }
          })}
          showSearch
        ></Select>
      </Form.Item>
      <Form.Item name='industry' label={t('version.industryBelonging')}>
        <Select
          options={companyIndustry.map((item) => {
            return {
              label: t(item.label as any) as string,
              options: item.options.map((v) => {
                const val = t(v as any) as string
                return { label: val, value: val }
              }),
            }
          })}
          showSearch
        ></Select>
      </Form.Item>
      {type === 'contact' ? (
        <Form.Item name='claim' label={t('version.yourRequirement')}>
          <Input />
        </Form.Item>
      ) : null}
      <Form.Item>
        <Button loading={loading} type='primary' htmlType='submit' block>
          {tC('button.submit')}
        </Button>
      </Form.Item>
    </Form>
  )
}

/** 验证码 */
const Code: FC<ValueController<string> & { type: FormType }> = (props) => {
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const { value, onChange, type } = props
  const form = Form.useFormInstance()
  const { message } = App.useApp()
  const [targetDate, setTargetDate] = useState<number>()
  const [countdown] = useCountDown({
    targetDate,
  })
  const { loading, run } = useRequest(
    async () => {
      const { phone } = await form.validateFields(['phone'])
      await sendUpgradeFormCode({ phone, type })
      message.success(tM('verificationCodeSent'))
      setTargetDate(Date.now() + 120 * 1000)
      await sleep(0)
    },
    { manual: true },
  )
  return (
    <div className='flex gap-4'>
      <Input
        value={value}
        onChange={(e) => onChange?.(e.target.value)}
        placeholder={t('version.pleaseEnter', { target: t('version.code') })}
      />
      <Button disabled={countdown !== 0} onClick={run} loading={loading}>
        {t('version.sendVerificationCode')}
        {countdown !== 0 ? `(${Math.ceil(countdown / 1000)})` : null}
      </Button>
    </div>
  )
}
