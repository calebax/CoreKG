import { FC } from 'react'
import { Form, FormInstance, Input, Select } from 'antd'
import { useTranslation } from 'react-i18next'
import { ModelInfo } from '..'

const FormItem = Form.Item<ModelInfo>

export type ModelFormProps = {
  form: FormInstance<ModelInfo>
  initialValues?: any
  setTested: (val: boolean) => void
}
export const ModelForm: FC<ModelFormProps> = (props) => {
  const { t } = useTranslation('pages')
  const { initialValues, setTested, form } = props
  console.log(initialValues)
  return (
    <Form
      form={form}
      initialValues={initialValues}
      onValuesChange={(changed?: Partial<ModelInfo>) => {
        setTested(false)
        if (changed?.model_provider) {
          form.setFieldsValue(providerDefaultValue[changed.model_provider])
        }
      }}
      preserve={false}
    >
      <FormItem
        label={t('settings.name', { target: t('settings.model') })}
        name='show_name'
        rules={[
          {
            required: true,
            message: t('settings.inputContent', {
              target: t('settings.name', { target: t('settings.model') }),
            }),
          },
        ]}
      >
        <Input />
      </FormItem>
      <FormItem
        label={t('settings.supplier')}
        name='model_provider'
        rules={[{ required: true, message: t('settings.selectSupplier') }]}
      >
        <Select
          options={modelProviderOptions}
          placeholder={t('settings.selectSupplier')}
        />
      </FormItem>
      <FormItem
        label={t('settings.base', { target: t('settings.model') })}
        name='model_name'
        rules={[
          {
            required: true,
            message: t('settings.inputContent', {
              target: t('settings.base', { target: t('settings.model') }),
            }),
          },
        ]}
      >
        <Input />
      </FormItem>
      <FormItem
        label={t('settings.address', { target: t('settings.model') })}
        name='model_url'
        rules={[
          {
            required: true,
            message: t('settings.inputContent', {
              target: t('settings.address', { target: t('settings.model') }),
            }),
          },
        ]}
      >
        <Input />
      </FormItem>
      <FormItem label='API Key' name='api_key'>
        <Input.Password />
      </FormItem>
    </Form>
  )
}
const modelProviderOptions = [
  {
    value: 'deepseek',
    label: 'DeepSeek',
  },
  {
    value: 'aliyun',
    label: '千问',
  },
]

const providerDefaultValue: Record<
  ModelInfo['model_provider'],
  Pick<ModelInfo, 'model_name' | 'model_url'>
> = {
  deepseek: {
    model_name: 'deepseek-chat',
    model_url: 'https://api.deepseek.com/chat/completions',
  },
  aliyun: {
    model_name: 'qwen2.5-72b-instruct',
    model_url:
      'https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions',
  },
}
