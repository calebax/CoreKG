import { FC } from 'react'
import {
  App,
  Button,
  Divider,
  Form,
  FormInstance,
  Modal,
  Skeleton,
  Tooltip,
} from 'antd'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { getCustomModelDetail, testCustomModel } from '@/api'
import { encryptPassword } from '@/utils/crypto'
import { ModelBaseInfo } from '../type'
import { ModelForm } from './ModelForm'
import './index.scss'

export type ModelInfo = ModelBaseInfo & {
  api_key?: string
}

export type ModelModal = {
  title: string
  onClose: () => void
  /** 提交或更新 由外部控制 */
  onSubmit: (val: ModelInfo) => any
  id?: number
}
/**
 * 不传open 每次需要销毁重建
 */
export const ModelModal: FC<ModelModal> = (props) => {
  const { t: tC } = useTranslation('common')
  const { t } = useTranslation('pages')
  const { title, onClose, id, onSubmit: _onSubmit } = props
  const [form] = Form.useForm<ModelInfo>()

  const { loading, data } = useRequest(
    async () => {
      if (!id) return undefined
      const { data } = await getCustomModelDetail({ id })
      return data as ModelInfo
    },
    { refreshDeps: [id], ready: Boolean(id) },
  )

  // 测试连接
  const { testing, testModel, tested, setTested } = useTest(form)
  const { run: onSubmit, loading: submiting } = useRequest(
    async () => {
      const formValue = await form.validateFields()
      await _onSubmit(formValue)
      onClose()
    },
    {
      manual: true,
    },
  )
  return (
    <Modal
      title={
        <span className='text-title text-base font-medium !leading-6'>
          {title}
        </span>
      }
      open
      onCancel={onClose}
      closable={true}
      keyboard={false}
      maskClosable={false}
      footer={null}
      destroyOnHidden
    >
      {loading ? (
        <Skeleton active />
      ) : (
        <>
          {/* 表单部分 */}
          <div className='pt-6 px-[38px] border-b border-b-[#F0F0F0]'>
            <ModelForm form={form} setTested={setTested} initialValues={data} />
          </div>

          {/* 底部按钮区域 */}
          <div className='flex items-center justify-between py-2.5 px-4'>
            <Button
              type='primary'
              ghost
              onClick={testModel}
              loading={testing}
              disabled={submiting}
              className='py-1 px-2.5 rounded-sm border border-[#3473EC] text-sm text-[#3473EC] font-medium'
            >
              {t('settings.connectionTest')}
            </Button>
            <Button disabled={submiting || testing} onClick={onClose}>
              {tC('button.cancel')}
            </Button>
            <Tooltip title={tested ? null : t('settings.connectionTestTip')}>
              <Button
                type='primary'
                loading={submiting}
                disabled={testing || !tested}
                onClick={onSubmit}
              >
                {tC('button.confirm')}
              </Button>
            </Tooltip>
          </div>
        </>
      )}
    </Modal>
  )
}

// 测试连接
const useTest = (form: FormInstance<ModelInfo>) => {
  const { t: tM } = useTranslation('messages')
  const { message } = App.useApp()
  const { loading, run, data, mutate } = useRequest(
    async () => {
      const testValue: Pick<ModelInfo, 'api_key' | 'model_name' | 'model_url'> =
        await form.validateFields(['api_key', 'model_name', 'model_url'])
      const { pass } = await testCustomModel({
        ...testValue,
        api_key: testValue.api_key ? encryptPassword(testValue.api_key) : '',
      })
      if (pass) {
        message.success(tM('modelConnectionTestPass'))
        return true
      } else {
        message.error(tM('testFailCheckConfigRetry'))
        return false
      }
    },
    { manual: true },
  )
  return {
    testing: loading,
    testModel: run,
    tested: data,
    setTested: mutate,
  }
}
