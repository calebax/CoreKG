import { FC } from 'react'
import { Alert, App, Button, Form, Input, InputNumber, Modal } from 'antd'
import { useBoolean, useRequest } from 'ahooks'
import { cn } from '@/utils'
import {
  createForestDBInstance,
  DatabaseInfo,
  getForestDBInstance,
  modifyForestDBInstance,
  testForestDBInstanceConnection,
} from '@/api/knowledge'
import { encryptPassword } from '@/utils/crypto'
import SettingIcon from '../../images/setting.svg?react'
import styles from '../../styles.module.scss'

/** 配置数据库按钮 */
export const ConfigBDBtn: FC<
  Style & { forest_id: number; afterConfig: () => void }
> = (props) => {
  const { forest_id, afterConfig, className, style } = props
  const [open, { toggle }] = useBoolean()
  const { data, loading, run, error } = useRequest(async () => {
    const res = await getForestDBInstance({
      // forest_db_instance_id: 46,
      forest_id,
    })
    // 判断是否存在实例
    if (!res.forest_db_instance_id) return undefined
    delete res.password
    return res
  })

  return (
    <>
      <Button
        className={cn(className, styles.createBtn)}
        style={style}
        onClick={toggle}
        loading={loading}
        disabled={Boolean(error)}
      >
        <SettingIcon className={styles.createBtnIcon} />
        配置数据库
      </Button>
      {open ? (
        <ConfigBDModal
          forest_id={forest_id}
          originConfig={data}
          onCancel={toggle}
          afterConfig={() => {
            run()
            afterConfig()
          }}
        />
      ) : null}
    </>
  )
}

const DBFormItem = Form.Item<DatabaseInfo>

const ConfigBDModal: FC<{
  onCancel: () => void
  forest_id: number
  afterConfig: () => void
  originConfig?: any
}> = (props) => {
  const { message } = App.useApp()
  const { forest_id, onCancel, afterConfig, originConfig } = props
  const [form] = Form.useForm<DatabaseInfo>()

  const {
    loading: testing,
    run: testConnect,
    data: status,
    mutate,
  } = useRequest(
    async () => {
      const dbInfo = await form.validateFields()
      const { connection_status, failure_reason } =
        await testForestDBInstanceConnection({
          ...dbInfo,
          password: encryptPassword(dbInfo.password),
          forest_id,
        })
      if (connection_status === 'valid') {
        message.success('测试成功')
        return 'success'
      } else {
        message.error(`测试失败，原因为: ${failure_reason}`)
        return 'error'
      }
    },
    {
      manual: true,
    },
  )
  const { loading: submiting, run: submit } = useRequest(
    async () => {
      const dbInfo = await form.validateFields()
      if (originConfig) {
        await modifyForestDBInstance({
          ...dbInfo,
          password: encryptPassword(dbInfo.password),
          forest_id,
        })
        message.success('修改成功')
      } else {
        await createForestDBInstance({
          ...dbInfo,
          password: encryptPassword(dbInfo.password),
          forest_id,
        })
        message.success('创建成功')
      }
      onCancel()
      afterConfig()
    },
    {
      manual: true,
    },
  )
  return (
    <Modal
      title='MySQL数据源'
      open
      onCancel={onCancel}
      keyboard={false}
      maskClosable={false}
      destroyOnHidden
      footer={null}
    >
      <Form
        autoComplete='off'
        form={form}
        requiredMark={false}
        initialValues={originConfig}
        onValuesChange={() => mutate(undefined)}
      >
        <DBFormItem
          label='数据源地址'
          name='host'
          rules={[{ required: true, message: '请输入数据源地址' }]}
        >
          <Input placeholder='请输入数据源地址' />
        </DBFormItem>
        <DBFormItem
          label='端口'
          name='port'
          rules={[{ required: true, message: '请输入端口号' }]}
        >
          <InputNumber
            placeholder='请输入端口'
            controls={false}
            className='w-40'
          />
        </DBFormItem>
        <DBFormItem
          label='数据库名称'
          name='database'
          rules={[{ required: true, message: '请输入数据库名称' }]}
        >
          <Input placeholder='请输入数据库名称' />
        </DBFormItem>
        <DBFormItem
          label='用户名'
          name='username'
          rules={[{ required: true, message: '请输入用户名' }]}
        >
          <Input placeholder='请输入用户名' />
        </DBFormItem>
        <DBFormItem
          label='密码'
          name='password'
          rules={[
            { required: true, message: '请输入密码' },
            { min: 8, message: '密码长度不能少于8位' },
            { max: 36, message: '密码长度不能超过36位' },
          ]}
        >
          <Input.Password placeholder='请输入密码' />
        </DBFormItem>
        {status ? (
          <Alert
            type={status}
            message={`连接测试${status === 'success' ? '成功' : '失败'}`}
            showIcon
            closable
          />
        ) : null}
        <div className='flex gap-4 mt-4'>
          <Button
            onClick={() => {
              mutate(undefined)
              testConnect()
            }}
            loading={testing}
            disabled={submiting}
          >
            测试连接
          </Button>
          <Button
            className='ml-auto'
            onClick={onCancel}
            disabled={testing || submiting}
          >
            取消
          </Button>
          <Button
            onClick={submit}
            loading={submiting}
            disabled={testing}
            className='bg-[#0C99FF] hover:bg-[#0C99FF] text-white'
          >
            确定
          </Button>
        </div>
      </Form>
    </Modal>
  )
}
